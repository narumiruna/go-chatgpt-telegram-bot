package bot

import (
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func messageLogger(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		m := c.Message()
		log.Infof("message ID: %d, from sender: %s (%d), text: %s, payload: %s",
			m.ID,
			m.Sender.Username,
			m.Sender.ID,
			m.Text,
			m.Payload,
		)
		return next(c)
	}
}
