package types

import "github.com/sashabaranov/go-openai"

type Chat struct {
	Window int

	Messages []openai.ChatCompletionMessage `json:"messages"`
}

func NewChat() *Chat {
	return &Chat{Window: -1}
}

func NewChatWindow(window int) *Chat {
	return &Chat{Window: window}
}

func (c *Chat) Add(message openai.ChatCompletionMessage) {
	c.Messages = append(c.Messages, message)
	if c.Window > 0 && len(c.Messages) > c.Window {
		c.Messages = c.Messages[len(c.Messages)-c.Window:]
	}
}

func (c *Chat) addMessage(role string, content string) {
	c.Add(openai.ChatCompletionMessage{Role: role, Content: content})
}

func (c *Chat) AddUserMessage(content string) {
	c.addMessage(openai.ChatMessageRoleUser, content)
}

func (c *Chat) AddAssistantMessage(content string) {
	c.addMessage(openai.ChatMessageRoleAssistant, content)
}

func (c *Chat) SetSystemMessage(content string) {
	// c.System = content
	c.addMessage(openai.ChatMessageRoleSystem, content)
}

func (c *Chat) LastContent() string {
	if len(c.Messages) == 0 {
		return ""
	}
	return c.Messages[len(c.Messages)-1].Content
}

// func (c *Chat) Messages() []openai.ChatCompletionMessage {
// 	var messages []openai.ChatCompletionMessage

// 	if c.System != "" {
// 		messages = append(messages, openai.ChatCompletionMessage{
// 			Role:    openai.ChatMessageRoleAssistant,
// 			Content: c.System,
// 		})
// 	}

// 	messages = append(messages, c.RawMessages...)
// 	return messages
// }
