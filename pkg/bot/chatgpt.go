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

type ChatGPTService struct {
	client         *openai.Client
	chats          store.Store
	systemContents store.Store
}

func NewChatGPTService(key string) *ChatGPTService {
	return &ChatGPTService{
		client:         openai.NewClient(key),
		chats:          store.New("chats"),
		systemContents: store.New("contents"),
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

func (g *ChatGPTService) reply(c tele.Context, chat *types.Chat) error {
	message := c.Message()

	request := openai.ChatCompletionRequest{
		Model:    openai.GPT4,
		Messages: chat.Messages,
	}

	log.Infof("request: %+v", request)

	completedMessage, err := g.complete(request)
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

	return g.chats.Save(key, chat)
}

func (g *ChatGPTService) HandleNewChat(c tele.Context) error {
	message := c.Message()

	userContent := strings.TrimPrefix(message.Text, "/gpt ")
	if userContent == "" {
		log.Infof("ignore empty content")
		return nil
	}

	chat := types.NewChat()

	var systemContent string
	if err := g.systemContents.Load(message.Chat.ID, &systemContent); err == nil {
		log.Infof("found system content: %s", systemContent)
		chat.AddSystemMessage(systemContent)
	}

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

func (g *ChatGPTService) HandleResetCommand(c tele.Context) error {
	return g.systemContents.Delete(c.Message().Chat.ID)
}

func (g *ChatGPTService) HandleSetCommand(c tele.Context) error {
	message := c.Message()

	content := strings.TrimPrefix(message.Text, "/set ")
	if content == "" {
		log.Infof("ignore empty content")
		return nil
	}

	return g.systemContents.Save(message.Chat.ID, content)
}

func (g *ChatGPTService) HandleTCCommand(c tele.Context) error {
	message := c.Message()

	chat := types.NewChat()
	systemContent := "You are a translation assistant." +
		" You will translate all messages to Traditional Chinese with traditional characters."

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
