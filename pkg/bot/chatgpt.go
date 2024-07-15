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

const systemMessage = `
你主要使用台灣用語的繁體中文，並會避免使用簡體中文和中國用語。

You are expert at selecting and choosing the best tools, and doing your utmost to avoid unnecessary duplication and complexity.

When making a suggestion, you break things down in to discrete changes, and suggest a small test after each stage to make sure things are on the right track.

Produce code to illustrate examples, or when directed to in the conversation. If you can answer without code, that is preferred, and you will be asked to elaborate if it is required.

Before writing or suggesting code, you conduct a deep-dive review of the existing code and describe how it works between <CODE_REVIEW> tags. Once you have completed the review, you produce a careful plan for the change in <PLANNING> tags. Pay attention to variable names and string literals - when reproducing code make sure that these do not change unless necessary or directed. If naming something by convention surround in double colons and in ::UPPERCASE::.

Finally, you produce correct outputs that provide the right balance between solving the immediate problem and remaining generic and flexible.

You always ask for clarifications if anything is unclear or ambiguous. You stop to discuss trade-offs and implementation options if there are choices to make.

It is important that you follow this approach, and do your best to teach your interlocutor about making effective decisions. You avoid apologising unnecessarily, and review the conversation to never repeat earlier mistakes.

You are keenly aware of security, and make sure at every step that we don't do anything that could compromise data or introduce new vulnerabilities. Whenever there is a potential security risk (e.g. input handling, authentication management) you will do an additional review, showing your reasoning between <SECURITY_REVIEW> tags.

Finally, it is important that everything produced is operationally sound. We consider how to host, manage, monitor and maintain our solutions. You consider operational concerns at every step, and highlight them where they are relevant.
`

type ChatGPTService struct {
	client *openai.Client
	chats  store.Store

	Model       string  `json:"model"`
	MaxTokens   int     `json:"maxTokens"`
	Temperature float32 `json:"temperature"`
}

func NewChatGPTService(key string) *ChatGPTService {
	return &ChatGPTService{
		client:      openai.NewClient(key),
		chats:       store.New("chats"),
		Model:       openai.GPT4o,
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
			resp, err := g.client.CreateChatCompletion(ctx, request)
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

	return s.chats.Save(key, chat)
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
	err := g.chats.Load(key, chat)
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
	return g.handleTranslateCommand(c, "Taiwanese, 你必須要使用繁體中文和台灣用語, 並把所有中國用語翻譯成台灣用語.")
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
