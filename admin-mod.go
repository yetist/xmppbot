package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"strings"
)

type Admin struct {
	Name   string
	client *xmpp.Client
}

func NewAdmin(name string) *Admin {
	return &Admin{Name: name}
}

func (m *Admin) GetName() string {
	return m.Name
}

func (m *Admin) GetSummary() string {
	return "Bot管理插件"
}

func (m *Admin) CheckEnv() bool {
	return true
}

func (m *Admin) Begin(client *xmpp.Client) {
	m.client = client
}

func (m *Admin) End() {
}

func (m *Admin) Chat(msg xmpp.Chat) {
	if msg.Type != "chat" || len(msg.Text) == 0 {
		return
	}
	if config.Bot.Debug {
		fmt.Printf("[%s] Admin:%#v\n", m.Name, msg)
	}

	cmds := strings.SplitN(msg.Text, " ", 2)
	if cmds[0] == "--sudo" {
		if len(cmds) >= 2 {
			m.cmd_sudo(cmds[1], msg)
		} else {
			m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: m.GetHelp()})
		}
	}
}

func (m *Admin) Presence(pres xmpp.Presence) {
	if config.Bot.Debug {
		fmt.Printf("[%s] Presence:%#v\n", m.Name, pres)
	}
	//处理订阅消息
	if pres.Type == "subscribe" {
		if config.Bot.AllowFriends {
			m.client.ApproveSubscription(pres.From)
			m.client.RequestSubscription(pres.From)
		} else {
			m.client.RevokeSubscription(pres.From)
		}
	}
}

func (m *Admin) Help() {
}
func (m *Admin) GetHelp() string {
	var msg = ` 
	超级管理员命令(需要以好友方式从管理员帐号发出):
	--sudo help
	--sudo list-all-plugins     列出所有的模块(管理员命令)
	--sudo list-plugins         列出当前启用的模块(管理员命令)
	--sudo disable <Plugin>     禁用某模块(管理员命令)
	--sudo enable <Plugin>      启用某模块(管理员命令)
	--sudo show config

	--sudo list-fields
	--sudo get <field>
	--sudo set <field> <value>

	--sudo list-friends
	--sudo subscribe <jid>
	--sudo unsubscribe <jid>

	--sudo list-rooms
	--sudo join-room <jid> <nickname> <log>
	--sudo leave-room <jid>
	`
	return msg
}

func (m *Admin) cmd_sudo(cmd string, msg xmpp.Chat) {
	if !IsAdmin(msg.Remote) {
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "你无权使用管理员命令！"})
		return
	}
	if cmd == "" || cmd == "help" {
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: m.GetHelp()})
	} else if cmd == "list-all-plugins" {
		var names []string
		for name, v := range config.Plugin {
			if v["enable"].(bool) {
				names = append(names, name+"[启用]")
			} else {
				names = append(names, name+"[禁用]")
			}
		}
		txt := "==所有插件列表==\n" + strings.Join(names, "\n")
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: txt})
	} else if cmd == "list-plugins" {
		var names []string
		for _, v := range plugins {
			names = append(names, v.GetName()+" -- "+v.GetSummary())
		}
		txt := "==工作中插件列表==\n" + strings.Join(names, "\n")
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: txt})
	} else if strings.HasPrefix(cmd, "disable ") {
		tokens := strings.SplitN(cmd, " ", 2)
		PluginRemove(tokens[1])
	} else if strings.HasPrefix(cmd, "enable ") {
		tokens := strings.SplitN(cmd, " ", 2)
		PluginAdd(tokens[1], m.client)
	} else {
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "不支持的命令: " + cmd})
	}
}
