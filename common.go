package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"sort"
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

// 设置状态消息
func SetStatus(client *xmpp.Client, status, info string) {
	client.SendOrg(fmt.Sprintf("<presence xml:lang='en'><show>%s</show><status>%s</status></presence>", status, info))
}

func MapDelete(dict map[string]interface{}, key string) {
	_, ok := dict[key]
	if ok {
		delete(dict, key)
	}
}

func SortMapKeys(dict map[string]string) []string {
	keys := make([]string, 0, len(dict))
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func ListDelete(list []string, key string) []string {
	for k, v := range list {
		if v == key {
			list = append(list[:k], list[k+1:]...)
		}
	}
	return list
}
