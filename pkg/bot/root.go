package bot

import (
	"context"
	"fmt"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

const defaultTimeout = 10 * time.Second

type ChatGPT struct {
	client      *openai.Client
	messagesMap messagesMap
}

func NewChatGPT(key string) *ChatGPT {
	messageMap := make(messagesMap)
	return &ChatGPT{
		client:      openai.NewClient(key),
		messagesMap: messageMap,
	}
}

func (g *ChatGPT) complete(ctx context.Context, messages Messages) (Messages, error) {
	request := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
	}

	resp, err := g.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, err
	}

	messages = append(messages, resp.Choices[0].Message)
	return messages, nil
}

func (g *ChatGPT) chat(c tele.Context) error {
	messages := Messages{}
	isReply := c.Message().IsReply()
	log.Infof("isReply: %t", isReply)

	if isReply {
		oldMessages, ok := g.messagesMap[c.Message().ReplyTo.ID]
		if !ok {
			return fmt.Errorf("no messages for ReplyTo.ID: %d", c.Message().ReplyTo.ID)
		}
		messages = append(messages, oldMessages...)
	}

	content := c.Message().Payload
	if isReply {
		content = c.Message().Text
	}
	log.Infof("user content: %s", content)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})

	messages, err := g.complete(context.Background(), messages)
	if err != nil {
		return err
	}

	teleMessage, err := c.Bot().Reply(c.Message(), messages.LastContent())
	if err != nil {
		return err
	}
	g.messagesMap[teleMessage.ID] = messages
	return nil
}

func Execute() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")

	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: defaultTimeout},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	chatGPT := NewChatGPT(openaiAPIKey)

	bot.Handle("/gpt", chatGPT.chat)
	bot.Handle(tele.OnText, chatGPT.chat)

	log.Infof("Starting bot")
	bot.Start()
}
