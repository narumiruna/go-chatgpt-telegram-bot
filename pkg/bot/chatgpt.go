package bot

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

type ChatGPT struct {
	client            *openai.Client
	openAIMessagesMap OpenAIMessagesMap
	validChatID       []int64
}

func NewChatGPT(key string, validChatID []int64) *ChatGPT {
	messageMap := make(OpenAIMessagesMap)
	return &ChatGPT{
		client:            openai.NewClient(key),
		openAIMessagesMap: messageMap,
		validChatID:       validChatID,
	}
}

func (g *ChatGPT) complete(ctx context.Context, messages OpenAIMessages) (OpenAIMessages, error) {
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

func (g *ChatGPT) newChat(c tele.Context) error {
	message := c.Message()

	context := fmt.Sprintf("%s\n%s", message.Payload, message.Text)
	log.Infof("context: %s", context)

	openAIMessages := OpenAIMessages{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: context,
		},
	}

	return g.chat(c, openAIMessages)
}

func (g *ChatGPT) reply(c tele.Context) error {
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
	log.Infof("key: %s", key)
	openAIMessages, ok := g.openAIMessagesMap[key]
	if !ok {
		openAIMessages = OpenAIMessages{{
			Role:    openai.ChatMessageRoleAssistant,
			Content: message.ReplyTo.Text,
		}}
	}

	openAIMessages = append(openAIMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message.Text,
	})

	return g.chat(c, openAIMessages)
}

func (g *ChatGPT) chat(c tele.Context, openAIMessages OpenAIMessages) error {
	openAIMessages, err := g.complete(context.Background(), openAIMessages)
	if err != nil {
		return err
	}

	replyMessage, err := c.Bot().Reply(c.Message(), openAIMessages.LastContent(), &tele.SendOptions{
		ParseMode: "Markdown",
	})
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%d@%d", replyMessage.ID, replyMessage.Chat.ID)
	log.Infof("key: %s", key)
	g.openAIMessagesMap[key] = openAIMessages
	return nil
}
