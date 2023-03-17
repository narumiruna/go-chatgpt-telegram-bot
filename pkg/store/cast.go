package store

import (
	"strconv"

	log "github.com/sirupsen/logrus"
)

func cast(data interface{}) string {
	switch v := data.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(data.(int))
	case int64:
		return strconv.FormatInt(data.(int64), 10)
	default:
		log.Infof("unsupported type %T", v)
		return ""
	}
}
