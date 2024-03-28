package types

import (
	"testing"

	"github.com/narumiruna/go-chatgpt-telegram-bot/pkg/store"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func compare(c1 *Chat, c2 *Chat) bool {
	length := len(c1.Messages)
	if length != len(c2.Messages) {
		return false
	}

	for i, m := range c1.Messages {
		n := c2.Messages[i]

		if m.Content != n.Content {
			return false
		}

		if m.Role != n.Role {
			return false
		}
	}
	return true
}

func Test_chat_store(t *testing.T) {
	cases := []struct {
		key       string
		namespace string
		roles     []string
		messages  []string
	}{
		{
			key:       "test_key",
			namespace: "test_namespace",
			roles:     []string{"system", "assistant", "user"},
			messages:  []string{"1", "2", "3"},
		},
	}

	for _, c := range cases {
		s := store.New(c.namespace)

		c1 := NewChat()
		for i, m := range c.messages {
			c1.Add(openai.ChatCompletionMessage{Role: c.roles[i], Content: m})
		}

		err1 := s.Save(c.key, c1)
		assert.NoError(t, err1)

		c2 := NewChat()
		err2 := s.Load(c.key, c2)
		assert.NoError(t, err2)

		assert.True(t, compare(c1, c2))
	}
}
