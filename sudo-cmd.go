package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"sort"
	"strings"
)

type Sudo struct {
	Name      string
	CmdPrefix string
	client    *xmpp.Client
}

func NewSudo(name string) *Sudo {
	return &Sudo{
		Name:      name,
		CmdPrefix: "--sudo",
	}
}

func (m *Sudo) GetName() string {
	return m.Name
}

func (m *Sudo) GetSummary() string {
	return "管理员模块，提供" + m.CmdPrefix + "开头的命令响应[内置]"
}

func (m *Sudo) CheckEnv() bool {
	return true
}

func (m *Sudo) Begin(client *xmpp.Client) {
	m.client = client
	m.client.Roster()
}

func (m *Sudo) End() {
}

func (m *Sudo) Chat(msg xmpp.Chat) {
	if msg.Type == "roster" {
		fmt.Printf("%#v\n", msg.Roster)
	}
	if msg.Type != "chat" || len(msg.Text) == 0 {
		return
	}
	if config.Bot.Debug {
		fmt.Printf("[%s] Chat:%#v\n", m.Name, msg)
	}

	/* 处理管理员命令 */
	cmds := strings.SplitN(msg.Text, " ", 2)
	if cmds[0] == m.CmdPrefix {
		if len(cmds) >= 2 {
			m.Command(cmds[1], msg)
		} else {
			m.cmd_help("", msg)
		}
	}
}

func (m *Sudo) Presence(pres xmpp.Presence) {
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

func (m *Sudo) Help() {
}

func (m *Sudo) Command(cmd string, msg xmpp.Chat) {
	if !IsAdmin(msg.Remote) {
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "你无权使用管理员命令！"})
		return
	}
	if cmd == "" || cmd == "help" {
		m.cmd_help(cmd, msg)
	} else if cmd == "list-all-plugins" {
		m.cmd_list_all_plugins(cmd, msg)
	} else if cmd == "list-plugins" {
		m.cmd_list_plugins(cmd, msg)
	} else if strings.HasPrefix(cmd, "disable ") {
		m.cmd_disable_plugin(cmd, msg)
	} else if strings.HasPrefix(cmd, "enable ") {
		m.cmd_enable_plugin(cmd, msg)
	} else if strings.HasPrefix(cmd, "subscribe ") {
		m.cmd_subscribe(cmd, msg)
	} else if strings.HasPrefix(cmd, "unsubscribe ") {
		m.cmd_unsubscribe(cmd, msg)
	} else if strings.HasPrefix(cmd, "status ") {
		m.cmd_status(cmd, msg)
	} else if cmd == "list-friends" {
		m.cmd_list_friends(cmd, msg)
	} else {
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "不支持的命令: " + cmd})
	}
}

func (m *Sudo) cmd_help(cmd string, msg xmpp.Chat) {

	help_msg := map[string]string{
		"help":                      "显示本信息",
		"list-all-plugins":          "列出所有的模块(管理员命令)",
		"list-plugins":              "列出当前启用的模块(管理员命令)",
		"disable <Plugin>":          "禁用某模块(管理员命令)",
		"enable <Plugin>":           "启用某模块(管理员命令)",
		"status <status> [message]": "设置机器人在线状态",
		"subscribe <jid>":           "请求加<jid>为好友",
		"unsubscribe <jid>":         "不再信认<jid>为好友",
		//"show config" : "",

		//"list-fields" : "",
		//"get <field>" : "",
		//"set <field> <value>" : "",

		//"list-friends" : "",
		//"list-rooms" : "",
		//"join-room <jid> <nickname> <log>" : "",
		//"leave-room <jid>" : "",
	}

	keys := make([]string, 0, len(help_msg))
	for key := range help_msg {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var help_list []string
	help_list = append(help_list, "==管理员命令==")
	for _, k := range keys {
		help_list = append(help_list, fmt.Sprintf("%s %s : %30s", m.CmdPrefix, k, help_msg[k]))
	}
	m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: strings.Join(help_list, "\n")})
}

func (m *Sudo) cmd_list_all_plugins(cmd string, msg xmpp.Chat) {
	var names []string
	names = append(names, m.Name+"[内置]")
	for name, v := range config.Plugin {
		if v["enable"].(bool) {
			names = append(names, name+"[启用]")
		} else {
			names = append(names, name+"[禁用]")
		}
	}
	txt := "==所有插件列表==\n" + strings.Join(names, "\n")
	m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: txt})
}

func (m *Sudo) cmd_list_plugins(cmd string, msg xmpp.Chat) {
	var names []string
	for _, v := range plugins {
		names = append(names, v.GetName()+" -- "+v.GetSummary())
	}
	txt := "==运行中插件列表==\n" + strings.Join(names, "\n")
	m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: txt})
}

func (m *Sudo) cmd_disable_plugin(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if tokens[1] == m.Name {
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: m.Name + "是内置模块，不允许禁用"})
	} else {
		PluginRemove(tokens[1])
	}
}

func (m *Sudo) cmd_enable_plugin(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	PluginAdd(tokens[1], m.client)
}

func (m *Sudo) cmd_subscribe(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		m.client.RequestSubscription(tokens[1])
	}
}

func (m *Sudo) cmd_unsubscribe(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		if IsAdmin(tokens[1]) {
			m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: tokens[1] + "是管理员，不允许从好友中删除！"})
		} else {
			m.client.RevokeSubscription(tokens[1])
		}
	}
}

func (m *Sudo) cmd_list_friends(cmd string, msg xmpp.Chat) {
	m.client.Roster()
}

func (m *Sudo) cmd_status(cmd string, msg xmpp.Chat) {
	// cmd is "status chat 正在聊天中..."
	var info = ""
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) == 3 {
		info = tokens[2]
	}
	if IsValidStatus(tokens[1]) {
		m.client.SendOrg(fmt.Sprintf("<presence xml:lang='en'><show>%s</show><status>%s</status></presence>", tokens[1], info))
	} else {
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "设置状态失败，有效的状态为: away, chat, dnd, xa."})
	}
}
