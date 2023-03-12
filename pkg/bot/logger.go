package bot

import (
	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func messageLogger(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		log.Infof("message text: %s", c.Message().Text)
		log.Infof("message payload: %s", c.Message().Payload)
		return next(c)
	}
}
