package bot

import (
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
			return nil
		}
	}
}
