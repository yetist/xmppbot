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
	for _, admin := range config.Setup.Admin {
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

// 回复好友消息，或聊天室私聊消息
func ReplyAuto(client *xmpp.Client, recv xmpp.Chat, text string) {
	client.Send(xmpp.Chat{Remote: recv.Remote, Type: "chat", Text: text})
}

// 回复好友消息，或聊天室公共消息
func ReplyPub(client *xmpp.Client, recv xmpp.Chat, text string) {
	if recv.Type == "groupchat" {
		roomid, _ := SplitJID(recv.Remote)
		client.Send(xmpp.Chat{Remote: roomid, Type: recv.Type, Text: text})
	} else {
		ReplyAuto(client, recv, text)
	}
}

// 发送到好友消息，或聊天室私聊消息
func SendAuto(client *xmpp.Client, to, text string) {
	client.Send(xmpp.Chat{Remote: to, Type: "chat", Text: text})
}

// 发送聊天室公共消息
func SendPub(client *xmpp.Client, to, text string) {
	client.Send(xmpp.Chat{Remote: to, Type: "groupchat", Text: text})
}
