package main

import (
	"github.com/mattn/go-xmpp"
)

var plugins []BotInterface
var admin_plugin AdminInterface

type BotInterface interface {
	Help() string
	GetName() string
	GetSummary() string
	CheckEnv() bool
	Chat(chat xmpp.Chat)
	Presence(pres xmpp.Presence)
	Begin(client *xmpp.Client)
	End()
	Restart()
	GetOptions() map[string]string
	SetOption(key, val string)
}

// 新增模块在这里注册
func CreatePlugin(name string, opt map[string]interface{}) BotInterface {
	var plugin BotInterface
	if name == "auto-reply" {
		plugin = NewAutoReply(name, opt)
	} else if name == "url-helper" {
		plugin = NewUrlHelper(name, opt)
	} else if name == "tuling" {
		plugin = NewTuling(name, opt)
	} else if name == "logger" {
		plugin = NewLogger(name, opt)
	}
	return plugin
}

// Interface(), 初始化并加载所有模块
func PluginInit() {
	// 自动启用内置插件
	admin := NewAdmin("admin")
	plugins = append(plugins, admin)
	admin_plugin = admin

	for name, v := range config.Plugin {
		if v["enable"].(bool) { //模块是否被启用
			plugin := CreatePlugin(name, v)
			if plugin != nil && plugin.CheckEnv() { //模块运行环境是否满足
				plugins = append(plugins, plugin)
			}
		}
	}
}

// Interface(), 模块加载时的处理函数
func PluginBegin(client *xmpp.Client) {
	for _, v := range plugins {
		v.Begin(client)
	}
}

// Interface(), 模块卸载时的处理函数
func PluginEnd() {
	for _, v := range plugins {
		v.End()
	}
}

// Interface(), 重新载入并初始化各模块
func PluginRestart(client *xmpp.Client) {
	var disable_plugins []string

	// 对正在运行中的插件，调用Restart接口重启
	for name, _ := range config.Plugin {
		for _, v := range plugins {
			if name == v.GetName() {
				v.Restart()
				continue
			}
		}
		disable_plugins = append(disable_plugins, name)
	}
	// 对禁用的插件，重新启用
	for _, n := range disable_plugins {
		PluginAdd(n, client)
	}
}

// Interface(), 模块收到消息时的处理
func PluginChat(chat xmpp.Chat) {
	for _, v := range plugins {
		v.Chat(chat)
	}
}

// Interface(), 模块收到Presence消息时的处理
func PluginPresence(pres xmpp.Presence) {
	for _, v := range plugins {
		v.Presence(pres)
	}
}

//获取管理员模块
func GetAdminPlugin() AdminInterface {
	return admin_plugin
}

func GetPluginByName(name string) BotInterface {
	for _, v := range plugins {
		if name == v.GetName() {
			return v
		}
	}
	return nil
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
