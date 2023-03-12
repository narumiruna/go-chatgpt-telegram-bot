package bot

import (
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func responseTimer(next tele.HandlerFunc) tele.HandlerFunc {
	now := time.Now().UnixMilli()
	return func(c tele.Context) error {
		responseTime := time.Now().UnixMilli() - now
		log.Infof("response time: %d ms", responseTime)
		return next(c)
	}
}
