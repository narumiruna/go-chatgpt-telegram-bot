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

func (g *ChatGPTService) handleTranslateCommand(c tele.Context, target string) error {
	message := c.Message()

	chat := types.NewChat()
	systemContent := fmt.Sprintf("You are a translation assistant and teacher. You will translate all messages to %s.", target)

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
	return g.handleTranslateCommand(c, "Japanese, and and provide comprehensive explanations of the relevant grammar points and usage in Traditional Chinese. Be sure to include clear examples and sample sentences to enhance understanding.")
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
