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
	Option    map[string]string
}

func NewSudo(name string) *Sudo {
	return &Sudo{
		Name: name,
		Option: map[string]string{
			"cmd":  config.Bot.AdminCmd,
			"help": config.Bot.HelpCmd,
		},
	}
}

func (m *Sudo) GetName() string {
	return m.Name
}

func (m *Sudo) GetSummary() string {
	return "管理员模块，提供" + m.Option["cmd"] + "开头的命令响应[内置]"
}

func (m *Sudo) GetOptions() map[string]string {
	return m.Option
}

func (m *Sudo) SetOption(key, val string) {
	// 如果有key的话，更新之
	m.Option[key] = val
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

func (m *Sudo) Restart() {
	LoadConfig(appName, appCfg)
	m.Option["cmd"] = config.Bot.AdminCmd
	m.Option["help"] = config.Bot.HelpCmd
	m.client.Roster()
	SetStatus(m.client, config.Account.Status, config.Account.StatusMessage)
}

func (m *Sudo) Chat(msg xmpp.Chat) {
	if msg.Type == "roster" {
		fmt.Printf("%#v\n", msg.Roster)
	}
	if msg.Type != "chat" || len(msg.Text) == 0 {
		return
	}

	// 仅处理好友消息
	if strings.HasPrefix(msg.Text, m.Option["help"]) {
		//cmd := strings.TrimSpace(msg.Text[len("--help"):])
		m.Help()
	} else if strings.HasPrefix(msg.Text, m.Option["cmd"]) {
		cmd := strings.TrimSpace(msg.Text[len(m.Option["cmd"]):])
		m.Command(cmd, msg)
	}
}

func (m *Sudo) Presence(pres xmpp.Presence) {
	if config.Bot.Debug {
		fmt.Printf("[%s] Presence:%#v\n", m.Name, pres)
	}
	//处理订阅消息
	if pres.Type == "subscribe" {
		if config.Bot.AutoSubscribe {
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
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "你不是管理员，无法执行管理员命令！"})
		return
	}
	if cmd == "" || cmd == "help" {
		m.cmd_help(cmd, msg)
	} else if cmd == "restart" {
		m.cmd_restart(cmd, msg)
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
	} else if cmd == "list-admin" {
		m.cmd_list_admin(cmd, msg)
	} else if strings.HasPrefix(cmd, "add-admin ") {
		m.cmd_add_admin(cmd, msg)
	} else if strings.HasPrefix(cmd, "del-admin ") {
		m.cmd_del_admin(cmd, msg)
		//} else if cmd == "list-friends" {
		//	m.cmd_list_friends(cmd, msg)
	} else {
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "不支持的命令: " + cmd})
	}
}

func (m *Sudo) cmd_help(cmd string, msg xmpp.Chat) {

	help_msg := map[string]string{
		"help":                        "显示本信息",
		"restart":                     "重新载入配置文件，初始化各模块",
		"list-all-plugins":            "列出所有的模块(管理员命令)",
		"list-plugins":                "列出当前启用的模块(管理员命令)",
		"disable <Plugin>":            "禁用某模块(管理员命令)",
		"enable <Plugin>":             "启用某模块(管理员命令)",
		"status <status> [message]":   "设置机器人在线状态",
		"subscribe <jid>":             "请求加<jid>为好友",
		"unsubscribe <jid>":           "不再信认<jid>为好友",
		"auto-subscribe <true|false>": "是否自动完成互加好友",
		"list-admin":                  "列出管理员帐号",
		"add-admin <jid>":             "新增管理员帐号",
		"del-admin <jid>":             "新增管理员帐号",

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
		help_list = append(help_list, fmt.Sprintf("%s %s : %30s", m.Option["cmd"], k, help_msg[k]))
	}
	m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: strings.Join(help_list, "\n")})
}

func (m *Sudo) cmd_restart(cmd string, msg xmpp.Chat) {
	m.Restart() //重启内置插件
	PluginRestart(m.client)
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

func (m *Sudo) cmd_status(cmd string, msg xmpp.Chat) {
	// cmd is "status chat 正在聊天中..."
	var info = ""
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) == 3 {
		info = tokens[2]
	}
	if IsValidStatus(tokens[1]) {
		SetStatus(m.client, tokens[1], info)
	} else {
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "设置状态失败，有效的状态为: away, chat, dnd, xa."})
	}
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

func (m *Sudo) cmd_auto_subscribe(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 {
		switch strings.ToLower(tokens[1]) {
		case
			"y",
			"yes",
			"1",
			"true",
			"T":
			return true
			config.Bot.AutoSubscribe = true
		}
		config.Bot.AutoSubscribe = false
	}
}

func (m *Sudo) cmd_list_admin(cmd string, msg xmpp.Chat) {
	txt := "==管理员列表==\n" + strings.Join(config.Bot.Admin, "\n")
	m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: txt})
}

func (m *Sudo) cmd_add_admin(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		if IsAdmin(tokens[1]) {
			m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: tokens[1] + " 已是管理员用户，不需再次增加！"})
		} else {
			m.client.RequestSubscription(tokens[1])
			config.Bot.Admin = append(config.Bot.Admin, tokens[1])
			m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "您已添加 " + tokens[1] + "为管理员!"})
			jid, _ := SplitJID(msg.Remote)
			m.client.Send(xmpp.Chat{Remote: tokens[1], Type: "chat", Text: jid + " 临时添加您为管理员!"})
		}
	}
}

func (m *Sudo) cmd_del_admin(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	jid, _ := SplitJID(msg.Remote)
	if IsAdmin(tokens[1]) && tokens[1] != jid {
		config.Bot.Admin = ListDelete(config.Bot.Admin, tokens[1])
		m.client.Send(xmpp.Chat{Remote: tokens[1], Type: "chat", Text: jid + " 临时取消了您的管理员身份!"})
	} else {
		m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "不能取消 " + tokens[1] + " 的管理员身份!"})
	}
}
