package bot

import (
	"github.com/narumiruna/go-chatgpt-telegram-bot/pkg/util"
	tele "gopkg.in/telebot.v3"
)

func responseTimer(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		defer util.Timer("bot response")()
		return next(c)
	}
}
