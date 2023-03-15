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

func (c *Chat) addMessage(role string, content string) {
	c.Add(openai.ChatCompletionMessage{Role: role, Content: content})
}

func (c *Chat) AddSystemMessage(content string) {
	c.addMessage(openai.ChatMessageRoleSystem, content)
}

func (c *Chat) AddUserMessage(content string) {
	c.addMessage(openai.ChatMessageRoleUser, content)
}

func (c *Chat) AddAssistantMessage(content string) {
	c.addMessage(openai.ChatMessageRoleAssistant, content)
}

func (c *Chat) LastContent() string {
	if len(c.Messages) == 0 {
		return ""
	}
	return c.Messages[len(c.Messages)-1].Content
}
