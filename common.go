package main

import (
	"strings"
)

func SplitJID(jid string) (string, string) {
	if strings.Contains(jid, "/") {
		tokens := strings.SplitN(jid, "/", 2)
		return tokens[0], tokens[1]
	} else {
		return jid, ""
	}
}

func IsAdmin(jid string) bool {
	u, _ := SplitJID(jid)
	for _, admin := range config.Bot.Admin {
		if u == admin {
			return true
		}
	}
	return false
}

func IsValidStatus(status string) bool {
	switch status {
	case
		"away",
		"chat",
		"dnd",
		"xa":
		return true
	}
	return false
}
