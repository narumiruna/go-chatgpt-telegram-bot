package bot

import (
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func responseTime(next tele.HandlerFunc) tele.HandlerFunc {
	now := time.Now().UnixMilli()
	return func(c tele.Context) error {
		log.Infof("response time: %d ms", time.Now().UnixMilli()-now)
		return next(c)
	}
}
