package bot

import (
	"strconv"
	"strings"
)

func parseInt64(s string) ([]int64, error) {
	if s == "" {
		return []int64{}, nil
	}

	var integers []int64
	for _, intString := range strings.Split(s, ",") {
		i, err := strconv.ParseInt(intString, 10, 64)
		if err != nil {
			return nil, err
		}
		integers = append(integers, i)
	}
	return integers, nil
}
