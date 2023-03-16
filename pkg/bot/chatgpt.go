package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/avast/retry-go"
	"github.com/narumiruna/go-chatgpt-telegram-bot/pkg/types"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

const defaultTemperature = 1.0

type ChatGPT struct {
	client         *openai.Client
	chats          types.ChatMap
	systemContents map[int64]string
}

func NewChatGPT(key string) *ChatGPT {
	return &ChatGPT{
		client:         openai.NewClient(key),
		chats:          make(types.ChatMap),
		systemContents: make(map[int64]string),
	}
}

func (g *ChatGPT) complete(ctx context.Context, chat *types.Chat) (*types.Chat, error) {
	request := openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    chat.Messages,
		Temperature: defaultTemperature,
	}

	err := retry.Do(
		func() error {
			defer timer("openai chat completion")()
			resp, err := g.client.CreateChatCompletion(ctx, request)
			if err != nil {
				return err
			}
			chat.Add(resp.Choices[0].Message)
			return nil
		},
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			log.Infof("retry %d: %v", n, err)
		}),
	)

	if err != nil {
		return nil, err
	}

	return chat, nil
}

func (g *ChatGPT) handleNewChat(c tele.Context) error {
	message := c.Message()

	content := strings.TrimPrefix(message.Text, "/gpt ")
	if content == "" {
		log.Infof("ignore empty contenxt")
		return nil
	}

	chat := types.NewChat()
	if content, ok := g.systemContents[message.Chat.ID]; ok {
		chat.AddSystemMessage(content)
	}

	if message.IsReply() {
		chat.AddUserMessage(message.ReplyTo.Text)
	}

	chat.AddUserMessage(content)
	return g.reply(c, chat)
}

func (g *ChatGPT) handleReply(c tele.Context) error {
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
	chat, ok := g.chats[key]
	if !ok {
		// ignore if the replyTo message is not from the bot
		if c.Bot().Me.ID != message.ReplyTo.Sender.ID {
			return nil
		}

		chat.AddAssistantMessage(message.ReplyTo.Text)
	}

	chat.AddUserMessage(message.Text)

	return g.reply(c, chat)
}

func (g *ChatGPT) reply(c tele.Context, chat *types.Chat) error {
	chat, err := g.complete(context.Background(), chat)
	if err != nil {
		return err
	}

	replyMessage, err := c.Bot().Reply(c.Message(), chat.LastContent(), &tele.SendOptions{
		ParseMode: "Markdown",
	})
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%d@%d", replyMessage.ID, replyMessage.Chat.ID)
	log.Infof("message key: %s", key)
	g.chats[key] = chat
	return nil
}

func (g *ChatGPT) setSystemContent(c tele.Context) error {
	message := c.Message()

	content := strings.TrimPrefix(message.Text, "/set ")
	if content == "" {
		log.Infof("ignore empty contenxt")
		return nil
	}

	g.systemContents[message.Chat.ID] = content
	return nil
}
