package main

import (
	"github.com/mattn/go-xmpp"
)

type BotModule interface {
	CheckEnv() bool
	Prep(client *xmpp.Client)
	Chat(chat xmpp.Chat)
	Presence(pres xmpp.Presence)
	Help()
	GetName() string
}

var plugins []BotModule

// 初始化所有启用的bot插件。
func InitPlugins() {
	// 检查每个插件运行环境是否满足，如不满足，则从列表中去掉。
	for m, v := range config.Plugin {
		if m == "chat" && v["enable"].(bool) {
			chat := NewChat("chat", v)
			if chat.CheckEnv() {
				plugins = append(plugins, chat)
			}
		} else if m == "muc" && v["enable"].(bool) {
			muc := NewMuc("muc", v)
			if muc.CheckEnv() {
				plugins = append(plugins, muc)
			}
		}
	}
}

// 当bot登录成功后，调用每个插件处理函数。
func PrepPlugins(client *xmpp.Client) {
	for _, v := range plugins {
		v.Prep(client)
	}
}

// 当收到聊天消息时，调用每个插件处理函数。
func ChatPlugins(chat xmpp.Chat) {
	for _, v := range plugins {
		v.Chat(chat)
	}
}

// 当收到Presence消息时，调用每个插件处理函数。
func PresencePlugins(pres xmpp.Presence) {
	for _, v := range plugins {
		v.Presence(pres)
	}
}

func helpPlugins(mod string) {
	if mod == "" {
		for _, v := range plugins {
			v.Help()
		}
		return
	}
	if mod == "plugins" {
	} else {
		for _, v := range plugins {
			if v.GetName() == mod {
				v.Help()
			}
		}
	}
}
