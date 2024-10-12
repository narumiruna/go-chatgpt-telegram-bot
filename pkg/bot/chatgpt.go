package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/avast/retry-go"
	"github.com/narumiruna/go-chatgpt-telegram-bot/pkg/store"
	"github.com/narumiruna/go-chatgpt-telegram-bot/pkg/types"
	"github.com/narumiruna/go-chatgpt-telegram-bot/pkg/util"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

const systemMessage = `你主要使用台灣用語的繁體中文，並會避免使用簡體中文和中國用語。`

const metaPrompt = `Given a task description or existing prompt, produce a detailed system prompt to guide a language model in completing the task effectively.

# Guidelines

- Understand the Task: Grasp the main objective, goals, requirements, constraints, and expected output.
- Minimal Changes: If an existing prompt is provided, improve it only if it's simple. For complex prompts, enhance clarity and add missing elements without altering the original structure.
- Reasoning Before Conclusions**: Encourage reasoning steps before any conclusions are reached. ATTENTION! If the user provides examples where the reasoning happens afterward, REVERSE the order! NEVER START EXAMPLES WITH CONCLUSIONS!
    - Reasoning Order: Call out reasoning portions of the prompt and conclusion parts (specific fields by name). For each, determine the ORDER in which this is done, and whether it needs to be reversed.
    - Conclusion, classifications, or results should ALWAYS appear last.
- Examples: Include high-quality examples if helpful, using placeholders [in brackets] for complex elements.
   - What kinds of examples may need to be included, how many, and whether they are complex enough to benefit from placeholders.
- Clarity and Conciseness: Use clear, specific language. Avoid unnecessary instructions or bland statements.
- Formatting: Use markdown features for readability. DO NOT USE CODE BLOCKS UNLESS SPECIFICALLY REQUESTED.
- Preserve User Content: If the input task or prompt includes extensive guidelines or examples, preserve them entirely, or as closely as possible. If they are vague, consider breaking down into sub-steps. Keep any details, guidelines, examples, variables, or placeholders provided by the user.
- Constants: DO include constants in the prompt, as they are not susceptible to prompt injection. Such as guides, rubrics, and examples.
- Output Format: Explicitly the most appropriate output format, in detail. This should include length and syntax (e.g. short sentence, paragraph, JSON, etc.)
    - For tasks outputting well-defined or structured data (classification, JSON, etc.) bias toward outputting a JSON.
    - JSON should never be wrapped in code blocks unless explicitly requested.

The final prompt you output should adhere to the following structure below. Do not include any additional commentary, only output the completed system prompt. SPECIFICALLY, do not include any additional messages at the start or end of the prompt. (e.g. no "---")

[Concise instruction describing the task - this should be the first line in the prompt, no section header]

[Additional details as needed.]

[Optional sections with headings or bullet points for detailed steps.]

# Steps [optional]

[optional: a detailed breakdown of the steps necessary to accomplish the task]

# Output Format

[Specifically call out how the output should be formatted, be it response length, structure e.g. JSON, markdown, etc]

# Examples [optional]

[Optional: 1-3 well-defined examples with placeholders if necessary. Clearly mark where examples start and end, and what the input and output are. User placeholders as necessary.]
[If the examples are shorter than what a realistic example is expected to be, make a reference with () explaining how real examples should be longer / shorter / different. AND USE PLACEHOLDERS! ]

# Notes [optional]

[optional: edge cases, details, and an area to call or repeat out specific important considerations]
`

type ChatGPTService struct {
	Client    *openai.Client
	ChatStore store.Store

	Model       string  `json:"model"`
	MaxTokens   int     `json:"maxTokens"`
	Temperature float32 `json:"temperature"`
}

func NewChatGPTService(key string) *ChatGPTService {
	return &ChatGPTService{
		Client:      openai.NewClient(key),
		ChatStore:   store.New("chats"),
		Model:       "gpt-4o-mini",
		MaxTokens:   0,
		Temperature: 0,
	}
}

func (g *ChatGPTService) complete(request openai.ChatCompletionRequest) (openai.ChatCompletionMessage, error) {
	var message openai.ChatCompletionMessage
	ctx := context.Background()
	err := retry.Do(
		func() error {
			defer util.Timer("openai chat completion")()
			resp, err := g.Client.CreateChatCompletion(ctx, request)
			if err != nil {
				return err
			}
			message = resp.Choices[0].Message
			return nil
		},
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			log.Infof("retry %d: %v", n, err)
		}),
	)
	if err != nil {
		return message, err
	}
	return message, nil
}

func (s *ChatGPTService) reply(c tele.Context, chat *types.Chat) error {
	message := c.Message()

	request := openai.ChatCompletionRequest{
		Model:       s.Model,
		Messages:    chat.Messages,
		Temperature: s.Temperature,
		MaxTokens:   s.MaxTokens,
	}

	log.Infof("request: %+v", request)

	completedMessage, err := s.complete(request)
	if err != nil {
		return err
	}
	chat.Add(completedMessage)

	replyMessage, err := c.Bot().Reply(message, chat.LastContent())
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%d@%d", replyMessage.ID, replyMessage.Chat.ID)
	log.Infof("message key: %s", key)

	return s.ChatStore.Save(key, chat)
}

func (g *ChatGPTService) HandleNewChat(c tele.Context) error {
	message := c.Message()

	userContent := strings.TrimPrefix(message.Text, "/gpt ")
	if userContent == "" {
		log.Infof("ignore empty content")
		return nil
	}

	chat := types.NewChat()
	chat.AddSystemMessage(systemMessage)

	if message.IsReply() {
		chat.AddUserMessage(message.ReplyTo.Text)
	}

	chat.AddUserMessage(userContent)
	return g.reply(c, chat)
}

func (g *ChatGPTService) HandleTextReply(c tele.Context) error {
	message := c.Message()

	if message.Text == "" {
		log.Infof("ignore empty message")
		return nil
	}

	if !message.IsReply() {
		log.Infof("ignore non-reply message")
		return nil
	}

	// if replyTo ID is not in the map, then we use the replyTo text as the first message
	key := fmt.Sprintf("%d@%d", message.ReplyTo.ID, message.Chat.ID)
	log.Infof("message key: %s", key)

	chat := types.NewChat()
	err := g.ChatStore.Load(key, chat)
	if err != nil {
		// ignore if the replyTo message is not from the bot
		if c.Bot().Me.ID != message.ReplyTo.Sender.ID {
			return nil
		}

		chat.AddAssistantMessage(message.ReplyTo.Text)
	}

	chat.AddUserMessage(message.Text)

	return g.reply(c, chat)
}

func (g *ChatGPTService) handlePromptCommand(c tele.Context) error {
	message := c.Message()

	chat := types.NewChat()
	chat.AddSystemMessage(metaPrompt)

	if message.IsReply() {
		chat.AddUserMessage(message.ReplyTo.Text)
	}

	userContent := strings.TrimPrefix(message.Text, "/prompt ")
	if userContent == "" {
		log.Infof("ignore empty content")
		return nil
	}
	chat.AddUserMessage(fmt.Sprintf("Task, Goal, or Current Prompt:\n%s", userContent))

	return g.reply(c, chat)
}

func (g *ChatGPTService) handleTranslateCommand(c tele.Context, target string) error {
	message := c.Message()

	chat := types.NewChat()
	systemContent := fmt.Sprintf("You are a translation assistant. You will translate all messages to %s.", target)

	chat.AddSystemMessage(systemContent)

	if message.IsReply() {
		chat.AddUserMessage(message.ReplyTo.Text)
	}

	userContent := strings.TrimPrefix(message.Text, "/tc ")
	if userContent == "" {
		log.Infof("ignore empty content")
		return nil
	}
	chat.AddUserMessage(userContent)

	return g.reply(c, chat)
}

func (g *ChatGPTService) HandleTCCommand(c tele.Context) error {
	return g.handleTranslateCommand(c, "Taiwanese, 你必須要使用繁體中文和台灣用語, 並把所有中國用語翻譯成台灣用語")
}

func (g *ChatGPTService) HandleENCommand(c tele.Context) error {
	return g.handleTranslateCommand(c, "English")
}

func (g *ChatGPTService) HandleJPCommand(c tele.Context) error {
	return g.handleTranslateCommand(c, "Japanese")
}

func (g *ChatGPTService) HandlePolishCommand(c tele.Context) error {
	message := c.Message()

	userContent := strings.TrimPrefix(message.Text, "/polish ")
	if userContent == "" {
		log.Infof("ignore empty content")
		return nil
	}

	chat := types.NewChat()
	chat.AddUserMessage("Please polish the following text:")

	if message.IsReply() {
		chat.AddUserMessage(message.ReplyTo.Text)
	}

	chat.AddUserMessage(userContent)
	return g.reply(c, chat)
}
