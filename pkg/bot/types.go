package bot

import (
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

type Messages []openai.ChatCompletionMessage

func (m Messages) LastContent() string {
	if len(m) == 0 {
		log.Errorf("Messages is empty")
		return ""
	}

	return m[len(m)-1].Content
}

type messagesMap map[int]Messages

func (m messagesMap) AppendMessage(key int, message openai.ChatCompletionMessage) {
	if _, ok := m[key]; !ok {
		m[key] = Messages{}
	}

	m[key] = append(m[key], message)
}
