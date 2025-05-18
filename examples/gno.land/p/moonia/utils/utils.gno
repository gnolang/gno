package utils

import (
	"strconv"
	"time"
	"strings"
)

func GetBlockTime() int64 {
	return time.Now().Unix()
}

func FormatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("02 Jan 2006, 15:04")
}

func ParseIndex(indexStr string, max int) int {
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		panic("Invalid index.")
	}
	if index < 0 || index >= max {
		panic("Proposal does not exist.")
	}
	return index
}

func ParseQuery(path string) (daoID string, showStats bool) {
	parts := strings.Split(path, "?")
	if len(parts) < 2 {
		return "", false
	}

	args := strings.Split(parts[1], "&")
	for _, arg := range args {
		if strings.HasPrefix(arg, "dao=") {
			daoID = arg[len("dao="):]
		} else if arg == "stats" {
			showStats = true
		}
	}
	return
}
