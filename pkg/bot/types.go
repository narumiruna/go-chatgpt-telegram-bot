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

type OpenAIMessagesMap map[string]OpenAIMessages
