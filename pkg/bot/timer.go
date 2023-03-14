package bot

import (
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func timer(s string) func() {
	start := time.Now()
	return func() {
		log.Infof("[timer] %s took %+v s", s, time.Since(start).Seconds())
	}
}

func responseTimer(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		defer timer("bot response")()
		return next(c)
	}
}
