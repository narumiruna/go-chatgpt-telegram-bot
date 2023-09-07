package types

import "github.com/sashabaranov/go-openai"

type Chat struct {
	Window int

	system   string
	messages []openai.ChatCompletionMessage
}

func NewChat() *Chat {
	return &Chat{Window: -1}
}

func NewChatWindow(window int) *Chat {
	return &Chat{Window: window}
}

func (c *Chat) Add(message openai.ChatCompletionMessage) {
	c.messages = append(c.messages, message)
	if c.Window > 0 && len(c.messages) > c.Window {
		c.messages = c.messages[len(c.messages)-c.Window:]
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
	c.system = content
}

func (c *Chat) LastContent() string {
	if len(c.messages) == 0 {
		return ""
	}
	return c.messages[len(c.messages)-1].Content
}

func (c *Chat) Messages() []openai.ChatCompletionMessage {
	var messages []openai.ChatCompletionMessage

	if c.system != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: c.system,
		})
	}

	messages = append(messages, c.messages...)
	return messages
}
