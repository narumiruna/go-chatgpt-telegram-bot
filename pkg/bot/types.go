package bot

import (
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

type OpenAIMessages []openai.ChatCompletionMessage

func (m OpenAIMessages) LastContent() string {
	if len(m) == 0 {
		log.Errorf("Messages is empty")
		return ""
	}

	return m[len(m)-1].Content
}

type OpenAIMessagesMap map[int]OpenAIMessages

func (m OpenAIMessagesMap) AppendMessage(key int, message openai.ChatCompletionMessage) {
	if _, ok := m[key]; !ok {
		m[key] = OpenAIMessages{}
	}

	m[key] = append(m[key], message)
}
