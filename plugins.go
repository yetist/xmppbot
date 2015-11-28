package main

import (
	"github.com/mattn/go-xmpp"
)

type BotPlugin interface {
	GetName() string
	GetSummary() string
	CheckEnv() bool
	Begin(client *xmpp.Client)
	Chat(chat xmpp.Chat)
	Presence(pres xmpp.Presence)
	End()
	Help()
}

var plugins []BotPlugin

// 初始化并加载所有模块
func PluginInit() {
	// 自动启用内置插件
	plugins = append(plugins, NewSudo("sudo"))

	for name, v := range config.Plugin {
		if v["enable"].(bool) { //模块是否被启用
			plugin := CreatePlugin(name, v)
			if plugin != nil && plugin.CheckEnv() { //模块运行环境是否满足
				plugins = append(plugins, plugin)
			}
		}
	}
}

// 模块加载时的处理函数
func PluginBegin(client *xmpp.Client) {
	for _, v := range plugins {
		v.Begin(client)
	}
}

// 模块卸载时的处理函数
func PluginEnd() {
	for _, v := range plugins {
		v.End()
	}
}

// 模块收到消息时的处理
func PluginChat(chat xmpp.Chat) {
	for _, v := range plugins {
		v.Chat(chat)
	}
}

// 模块收到Presence消息时的处理
func PluginPresence(pres xmpp.Presence) {
	for _, v := range plugins {
		v.Presence(pres)
	}
}

// 按名称卸载某个模块
func PluginRemove(name string) {
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

// 按名称加载某个模块
func PluginAdd(name string, client *xmpp.Client) {
	for _, v := range plugins {
		if name == v.GetName() {
			return
		}
	}
	for n, v := range config.Plugin {
		if n == name && v["enable"].(bool) {
			plugin := CreatePlugin(name, v)
			if plugin != nil && plugin.CheckEnv() { //模块运行环境是否满足
				plugin.Begin(client)
				plugins = append(plugins, plugin)
			}
		}
	}
}

func CreatePlugin(name string, opt map[string]interface{}) BotPlugin {
	var plugin BotPlugin
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
