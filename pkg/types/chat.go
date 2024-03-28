package types

import "github.com/sashabaranov/go-openai"

type Chat struct {
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

func NewChat() *Chat {
	return &Chat{}
}

func (c *Chat) Add(message openai.ChatCompletionMessage) {
	c.Messages = append(c.Messages, message)
}

func (c *Chat) AddUserMessage(content string) {
	c.Add(openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: content})
}

func (c *Chat) AddAssistantMessage(content string) {
	c.Add(openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: content})
}

func (c *Chat) AddSystemMessage(content string) {
	c.Add(openai.ChatCompletionMessage{Role: openai.ChatMessageRoleSystem, Content: content})
}

func (c *Chat) LastContent() string {
	if len(c.Messages) == 0 {
		return ""
	}
	return c.Messages[len(c.Messages)-1].Content
}
