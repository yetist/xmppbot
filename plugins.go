package main

import (
	"github.com/mattn/go-xmpp"
)

type BotModule interface {
	GetName() string
	GetSummary() string
	CheckEnv() bool
	Begin(client *xmpp.Client)
	Chat(chat xmpp.Chat)
	Presence(pres xmpp.Presence)
	End()
	Help()
}

var plugins []BotModule

// 初始化所有启用的bot插件。
func PluginInit() {
	// 自动启用内置插件
	plugins = append(plugins, NewAdmin("admin"))

	// 检查每个插件运行环境是否满足，如不满足，则从列表中去掉。
	for name, v := range config.Plugin {
		if name == "chat" && v["enable"].(bool) {
			plugin := NewChat(name, v)
			if plugin.CheckEnv() {
				plugins = append(plugins, plugin)
			}
		} else if name == "muc" && v["enable"].(bool) {
			plugin := NewMuc(name, v)
			if plugin.CheckEnv() {
				plugins = append(plugins, plugin)
			}
		}
	}
}

// 当bot登录成功后，调用每个插件处理函数。
func PluginBegin(client *xmpp.Client) {
	for _, v := range plugins {
		v.Begin(client)
	}
}

// 当收到聊天消息时，调用每个插件处理函数。
func PluginChat(chat xmpp.Chat) {
	for _, v := range plugins {
		v.Chat(chat)
	}
}

// 当收到Presence消息时，调用每个插件处理函数。
func PluginPresence(pres xmpp.Presence) {
	for _, v := range plugins {
		v.Presence(pres)
	}
}

// 当bot登录成功后，调用每个插件处理函数。
func PluginRemove(name string) {
	//不允许禁用内置模块
	if name == "admin" {
		return
	}

	id := -1
	for k, v := range plugins {
		if name == v.GetName() {
			v.End()
			id = k
		}
	}
	if id > -1 {
		plugins = append(plugins[:id], plugins[id+1:]...)
	}
}

func PluginAdd(name string, client *xmpp.Client) {
	for _, v := range plugins {
		if name == v.GetName() {
			return
		}
	}
	for n, v := range config.Plugin {
		if n == name && v["enable"].(bool) {
			plugin := CreatePlugin(name, v)
			if plugin.CheckEnv() {
				plugin.Begin(client)
				plugins = append(plugins, plugin)
			}
		}
	}
}

func CreatePlugin(name string, opt map[string]interface{}) BotModule {
	var plugin BotModule
	if name == "chat" {
		plugin = NewChat(name, opt)
	} else if name == "muc" {
		plugin = NewMuc(name, opt)
	}
	return plugin
}

////////////////////////////////////////
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
