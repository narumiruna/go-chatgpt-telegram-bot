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

func (g *ChatGPT) isValidChatID(chatID int64) bool {
	for _, id := range g.validChatID {
		if id == chatID {
			return true
		}
	}
	return false
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

	if message.Payload == "" {
		log.Infof("ignore empty payload")
		return nil
	}

	// If chatIDs is not empty, then we only accept messages from those chatIDs
	chatID := message.Chat.ID
	if len(g.validChatID) != 0 && !g.isValidChatID(chatID) {
		return c.Reply(fmt.Sprintf("Sorry, I'm not allowed to talk to you :(. Add your chat ID: %d to the VALID_CHAT_ID env var.", chatID))
	}

	openAIMessages := OpenAIMessages{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: message.Payload,
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

	// If chatIDs is not empty, then we only accept messages from those chatIDs
	chatID := message.Chat.ID
	if len(g.validChatID) != 0 && !g.isValidChatID(chatID) {
		return c.Reply(fmt.Sprintf("Sorry, I'm not allowed to talk to you :(. Add your chat ID: %d to the VALID_CHAT_ID env var.", chatID))
	}

	// if replyTo ID is not in the map, then we use the replyTo text as the first message
	openAIMessages, ok := g.openAIMessagesMap[message.ReplyTo.ID]
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
	g.openAIMessagesMap[replyMessage.ID] = openAIMessages
	return nil
}
