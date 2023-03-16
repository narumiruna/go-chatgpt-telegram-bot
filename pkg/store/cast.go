package store

import (
	"fmt"
	"strconv"
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
		panic(fmt.Sprintf("unsupported type %T", v))
	}
}
