package bot

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func whitelist(chats ...int64) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			for _, id := range chats {
				if c.Message().Chat.ID == id {
					return next(c)
				}
			}
			return c.Reply(fmt.Sprintf("chat ID: %d is not in the whitelist", c.Message().Chat.ID))
		}
	}
}
