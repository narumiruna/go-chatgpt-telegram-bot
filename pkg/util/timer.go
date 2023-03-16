package util

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func Timer(s string) func() {
	start := time.Now()
	return func() {
		log.Infof("[timer] %s took %+v s", s, time.Since(start).Seconds())
	}
}
