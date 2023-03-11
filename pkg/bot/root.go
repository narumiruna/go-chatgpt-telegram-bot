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

func (g *ChatGPT) start(c tele.Context) error {
	payload := c.Message().Payload

	messages := Messages{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: payload,
		},
	}

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

func (g *ChatGPT) reply(c tele.Context) error {
	if !c.Message().IsReply() {
		log.Infof("Message is not a reply: %+v", c.Message())
		return nil
	}

	replyToID := c.Message().ReplyTo.ID
	log.Infof("Reply ID: %d", replyToID)
	messages, ok := g.messagesMap[replyToID]
	if !ok {
		return fmt.Errorf("no messages for reply ID: %d", replyToID)
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: c.Message().Text,
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

	bot.Handle("/gpt", chatGPT.start)
	bot.Handle(tele.OnText, chatGPT.reply)

	log.Infof("Starting bot")
	bot.Start()
}
