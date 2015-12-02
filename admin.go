package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"sort"
	"strings"
)

type Admin struct {
	Name   string
	client *xmpp.Client
	Option map[string]interface{}
	Rooms  []*Room
}

func NewAdmin(name string) *Admin {
	var rooms []*Room
	for _, i := range config.Setup.Rooms {
		var password string
		jid := i["jid"].(string)
		nickname := i["nickname"].(string)
		if i["password"] == nil {
			password = ""
		} else {
			password = i["password"].(string)
		}
		room := NewRoom(jid, nickname, password, config.Setup.Status, config.Setup.StatusMessage)
		rooms = append(rooms, room)
	}
	return &Admin{
		Name:  name,
		Rooms: rooms,
		Option: map[string]interface{}{
			"cmd":            config.Setup.AdminCmd,
			"help":           config.Setup.HelpCmd,
			"room":           "--room",
			"auto-subscribe": config.Setup.AutoSubscribe,
			"admin":          config.Setup.Admin,
		},
	}
}

func (m *Admin) GetName() string {
	return m.Name
}

func (m *Admin) GetSummary() string {
	return "管理员模块，提供" + m.Option["cmd"].(string) + "开头的命令响应[内置]"
}

func (m *Admin) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		switch k {
		case "cmd":
			opts[k] = v.(string) + "  (管理员命令前缀)"
		case "help":
			opts[k] = v.(string) + "  (帮助命令前缀)"
		case "room":
			opts[k] = v.(string) + "  (聊天室命令前缀)"
		case "auto-subscribe":
			opts[k] = BoolToString(v.(bool)) + "  (是否自动完成互加好友)"
		case "admin":
			opts[k] = strings.Join(v.([]string), "; ") + "  (管理员列表)"
		}
	}
	return opts
}

func (m *Admin) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		if key == "auto-subscribe" {
			m.Option[key] = StringToBool(val)
		} else if key == "admin" {
			//TODO: 忽略对管理员列表的设置
			return
		} else {
			m.Option[key] = val
		}
	}
}

func (m *Admin) CheckEnv() bool {
	return true
}

func (m *Admin) Begin(client *xmpp.Client) {
	m.client = client
	//m.client.Roster()
	for _, room := range m.Rooms {
		if len(room.Password) > 0 {
			m.client.JoinProtectedMUC(room.JID, room.Nickname, room.Password)
		} else {
			m.client.JoinMUC(room.JID, room.Nickname)
		}
		fmt.Printf("[%s] Join to %s as %s\n", m.Name, room.JID, room.Nickname)
	}
}

func (m *Admin) End() {
	for _, room := range m.Rooms {
		m.client.LeaveMUC(room.JID)
		fmt.Printf("[%s] Leave from %s\n", m.Name, room.JID)
	}
}

func (m *Admin) Restart() {
	m.End()
	LoadConfig(AppName, AppVersion, AppConfig)
	m.Option["cmd"] = config.Setup.AdminCmd
	m.Option["help"] = config.Setup.HelpCmd
	m.Option["auto-subscribe"] = config.Setup.AutoSubscribe
	m.client.Roster()
	SetStatus(m.client, config.Setup.Status, config.Setup.StatusMessage)

	var rooms []*Room
	for _, i := range config.Setup.Rooms {
		var password string
		jid := i["jid"].(string)
		nickname := i["nickname"].(string)
		if i["password"] == nil {
			password = ""
		} else {
			password = i["password"].(string)
		}
		room := NewRoom(jid, nickname, password, config.Setup.Status, config.Setup.StatusMessage)
		rooms = append(rooms, room)
	}
	m.Rooms = rooms
	m.Begin(m.client)
}

func (m *Admin) Chat(msg xmpp.Chat) {
	if msg.Type == "roster" {
		fmt.Printf("%#v\n", msg.Roster)
	}
	if len(msg.Text) == 0 {
		return
	}

	// 仅处理好友消息
	if msg.Text == "--version" {
		ReplyPub(m.client, msg, AppName+"-"+AppVersion)
	} else if strings.HasPrefix(msg.Text, m.Option["help"].(string)) {
		cmd := strings.TrimSpace(msg.Text[len(m.Option["help"].(string)):])
		m.HelpCommand(cmd, msg)
	} else if strings.HasPrefix(msg.Text, m.Option["room"].(string)) {
		cmd := strings.TrimSpace(msg.Text[len(m.Option["room"].(string)):])
		m.RoomCommand(cmd, msg)
	} else if strings.HasPrefix(msg.Text, m.Option["cmd"].(string)) {
		cmd := strings.TrimSpace(msg.Text[len(m.Option["cmd"].(string)):])
		m.AdminCommand(cmd, msg)
	}
}

func (m *Admin) Presence(pres xmpp.Presence) {
	if config.Setup.Debug {
		fmt.Printf("[%s] Presence:%#v\n", m.Name, pres)
	}
	//处理订阅消息
	if pres.Type == "subscribe" {
		if m.Option["auto-subscribe"].(bool) {
			m.client.ApproveSubscription(pres.From)
			m.client.RequestSubscription(pres.From)
		} else {
			m.client.RevokeSubscription(pres.From)
		}
	}
}

func (m *Admin) Help() string {
	msg := []string{
		"管理员模块为内置模块，提供了管理命令，用来设置及改变Bot的行为。",
		"支持以下命令：",
		"--help    查看帮助命令详情",
		"--sudo    查看管理员命令详情",
		"--room    查看聊天室命令详情",
		"--version 查看Bot版本",
	}
	return strings.Join(msg, "\n")
}

func (m *Admin) GetRooms() []*Room {
	return m.Rooms
}

func (m *Admin) IsAdmin(jid string) bool {
	u, _ := SplitJID(jid)
	for _, admin := range config.Setup.Admin {
		if u == admin {
			return true
		}
	}
	return false
}

func (m *Admin) HelpCommand(cmd string, msg xmpp.Chat) {
	if cmd == "" {
		m.cmd_help_all(cmd, msg)
	} else if len(cmd) > 0 {
		m.cmd_help_plugin(cmd, msg)
	} else {
		ReplyAuto(m.client, msg, "不支持的命令: "+cmd)
	}
}

func (m *Admin) cmd_help_all(cmd string, msg xmpp.Chat) {
	var helps []string
	for _, v := range plugins {
		helps = append(helps, "=="+v.GetName()+"模块帮助信息==")
		helps = append(helps, v.Help())
	}
	txt := "==所有插件帮助信息==\n" + strings.Join(helps, "\n")
	ReplyAuto(m.client, msg, txt)
}

func (m *Admin) cmd_help_plugin(cmd string, msg xmpp.Chat) {
	var helps []string
	names := strings.Split(cmd, " ")
	for _, name := range names {
		if v := GetPluginByName(name); v != nil {
			helps = append(helps, "=="+v.GetName()+"帮助信息==")
			helps = append(helps, v.Help())
		}
	}
	ReplyAuto(m.client, msg, strings.Join(helps, "\n"))
}

func (m *Admin) RoomCommand(cmd string, msg xmpp.Chat) {
	if cmd == "" || cmd == "help" {
		m.room_help(cmd, msg)
	} else if strings.HasPrefix(cmd, "msg ") {
		m.room_msg(cmd, msg)
	} else if strings.HasPrefix(cmd, "nick ") {
		m.room_nick(cmd, msg)
	} else if strings.HasPrefix(cmd, "list-blocks ") {
		m.room_list_blocks(cmd, msg)
	} else if strings.HasPrefix(cmd, "block ") {
		m.room_block(cmd, msg)
	} else if strings.HasPrefix(cmd, "unblock ") {
		m.room_unblock(cmd, msg)
	} else if strings.HasPrefix(cmd, "status ") {
		m.room_status(cmd, msg)
	} else {
		ReplyAuto(m.client, msg, "不支持的命令: "+cmd)
	}
}

func (m *Admin) room_help(cmd string, msg xmpp.Chat) {

	help_msg := map[string]string{
		"help":                                "显示本信息",
		"msg <jid|all> <msg>":                 "让机器人在聊天室中发送消息msg",
		"nick <jid|all> <NickName>":           "修改机器人在聊天室的昵称为NickName",
		"list-blocks <jid|all>":               "查看聊天室屏蔽列表",
		"block <jid|all> <who>":               "屏蔽who，对who发送的消息不响应",
		"unblock <jid|all> <who>":             "重新对who发送的消息进行响应",
		"status <jid|all> <status> [message]": "设置机器人在线状态",
	}

	keys := make([]string, 0, len(help_msg))
	for key := range help_msg {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var help_list []string
	help_list = append(help_list, "==聊天室命令==")
	for _, k := range keys {
		help_list = append(help_list, fmt.Sprintf("%s %s : %30s", m.Option["room"], k, help_msg[k]))
	}
	ReplyAuto(m.client, msg, strings.Join(help_list, "\n"))
}

func (m *Admin) room_msg(cmd string, msg xmpp.Chat) {
	//"msg <jid|all> <msg>":       "让机器人在聊天室中发送消息msg",
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) != 3 {
		return
	}
	if tokens[1] == "all" {
		for _, v := range m.Rooms {
			SendPub(m.client, v.JID, tokens[2])
		}
	} else {
		for _, v := range m.Rooms {
			if v.JID == tokens[1] {
				SendPub(m.client, tokens[1], tokens[2])
			} else {
				ReplyAuto(m.client, msg, "Bot未进入此聊天室")
			}
		}
	}
}

func (m *Admin) room_nick(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) != 3 {
		return
	}
	if tokens[1] == "all" {
		for _, v := range m.Rooms {
			v.SetNick(m.client, tokens[2])
		}
	} else {
		for _, v := range m.Rooms {
			if v.JID == tokens[1] {
				v.SetNick(m.client, tokens[2])
			} else {
				ReplyAuto(m.client, msg, "Bot未进入此聊天室")
			}
		}
	}
}

func (m *Admin) room_list_blocks(cmd string, msg xmpp.Chat) {
	//"list-blocks <jid|all>":     "屏蔽who，对who发送的消息不响应",
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) != 2 {
		return
	}
	var blocks []string
	if tokens[1] == "all" {
		for _, v := range m.Rooms {
			blocks = append(blocks, v.ListBlocks())
		}
	} else {
		for _, v := range m.Rooms {
			if v.JID == tokens[1] {
				blocks = append(blocks, v.ListBlocks())
			} else {
				ReplyAuto(m.client, msg, "Bot未进入此聊天室")
			}
		}
	}
	ReplyAuto(m.client, msg, strings.Join(blocks, "\n"))
}

func (m *Admin) room_block(cmd string, msg xmpp.Chat) {
	//"block <jid|all> <who>":     "屏蔽who，对who发送的消息不响应",
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) != 3 {
		return
	}
	if tokens[1] == "all" {
		for _, v := range m.Rooms {
			SendPub(m.client, v.JID, "/me 忽略了 "+tokens[2]+" 的消息")
			v.BlockOne(tokens[2])
		}
	} else {
		for _, v := range m.Rooms {
			if v.JID == tokens[1] {
				SendPub(m.client, v.JID, "/me 忽略了 "+tokens[2]+" 的消息")
				v.BlockOne(tokens[2])
			} else {
				ReplyAuto(m.client, msg, "Bot未进入此聊天室")
			}
		}
	}
}

func (m *Admin) room_unblock(cmd string, msg xmpp.Chat) {
	//"unblock <jid|all> <who>":   "重新对who发送的消息进行响应",
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) != 3 {
		return
	}
	if tokens[1] == "all" {
		for _, v := range m.Rooms {
			SendPub(m.client, v.JID, "/me 开始关注 "+tokens[2]+" 的消息")
			v.UnBlockOne(tokens[2])
		}
	} else {
		for _, v := range m.Rooms {
			if v.JID == tokens[1] {
				SendPub(m.client, v.JID, "/me 开始关注 "+tokens[2]+" 的消息")
				v.UnBlockOne(tokens[2])
			} else {
				ReplyAuto(m.client, msg, "Bot未进入此聊天室")
			}
		}
	}
}

func (m *Admin) room_status(cmd string, msg xmpp.Chat) {
	//"status <jid|all> <status> [message]": "设置机器人在线状态",
	tokens := strings.SplitN(cmd, " ", 4)
	if len(tokens) != 4 || len(tokens) != 3 {
		return
	}
	info := ""
	if len(tokens) == 4 {
		info = tokens[3]
	}
	if tokens[1] == "all" {
		for _, v := range m.Rooms {
			v.SetStatus(m.client, tokens[2], info)
		}
	} else {
		for _, v := range m.Rooms {
			if v.JID == tokens[1] {
				v.SetStatus(m.client, tokens[2], info)
			} else {
				ReplyAuto(m.client, msg, "Bot未进入此聊天室")
			}
		}
	}
}

func (m *Admin) AdminCommand(cmd string, msg xmpp.Chat) {
	if !m.IsAdmin(msg.Remote) {
		ReplyAuto(m.client, msg, "请确认您是管理员，并且通过好友消息发送了此命令。")
		return
	}
	if cmd == "" || cmd == "help" {
		m.admin_help(cmd, msg)
	} else if cmd == "restart" {
		m.admin_restart(cmd, msg)
	} else if cmd == "list-all-plugins" {
		m.admin_list_all_plugins(cmd, msg)
	} else if cmd == "list-plugins" {
		m.admin_list_plugins(cmd, msg)
	} else if strings.HasPrefix(cmd, "disable ") {
		m.admin_disable_plugin(cmd, msg)
	} else if strings.HasPrefix(cmd, "enable ") {
		m.admin_enable_plugin(cmd, msg)
	} else if strings.HasPrefix(cmd, "subscribe ") {
		m.admin_subscribe(cmd, msg)
	} else if strings.HasPrefix(cmd, "unsubscribe ") {
		m.admin_unsubscribe(cmd, msg)
	} else if strings.HasPrefix(cmd, "status ") {
		m.admin_status(cmd, msg)
	} else if cmd == "list-admin" {
		m.admin_list_admin(cmd, msg)
	} else if strings.HasPrefix(cmd, "add-admin ") {
		m.admin_add_admin(cmd, msg)
	} else if strings.HasPrefix(cmd, "del-admin ") {
		m.admin_del_admin(cmd, msg)
	} else if cmd == "list-options" {
		m.admin_list_options(cmd, msg)
	} else if strings.HasPrefix(cmd, "set-option ") {
		m.admin_set_option(cmd, msg)
	} else if cmd == "list-rooms" {
		m.admin_list_rooms(cmd, msg)
	} else if strings.HasPrefix(cmd, "join-room") {
		m.admin_join_room(cmd, msg)
	} else if strings.HasPrefix(cmd, "leave-room") {
		m.admin_leave_room(cmd, msg)
	} else {
		ReplyAuto(m.client, msg, "不支持的命令: "+cmd)
	}
}

func (m *Admin) admin_help(cmd string, msg xmpp.Chat) {

	help_msg := map[string]string{
		"help":                       "显示本信息",
		"restart":                    "重新载入配置文件，初始化各模块",
		"list-all-plugins":           "列出所有的模块(管理员命令)",
		"list-plugins":               "列出当前启用的模块(管理员命令)",
		"disable <Plugin>":           "禁用某模块(管理员命令)",
		"enable <Plugin>":            "启用某模块(管理员命令)",
		"status <status> [message]":  "设置机器人在线状态",
		"subscribe <jid>":            "请求加<jid>为好友",
		"unsubscribe <jid>":          "不再信认<jid>为好友",
		"list-admin":                 "列出管理员帐号",
		"add-admin <jid>":            "新增管理员帐号",
		"del-admin <jid>":            "新增管理员帐号",
		"list-options":               "列出所有模块可配置选项",
		"set-option <field> <value>": "设置模块相关选项",

		"list-rooms":                            "列出机器人当前所在的聊天室",
		"join-room <jid> <nickname> [password]": "加入聊天室",
		"leave-room <jid>":                      "离开聊天室",

		//"list-fields" : "",
		//"get <field>" : "",
		//"set <field> <value>" : "",
		//"list-friends" : "",
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
	ReplyAuto(m.client, msg, strings.Join(help_list, "\n"))
}

func (m *Admin) admin_restart(cmd string, msg xmpp.Chat) {
	m.Restart() //重启内置插件
	PluginRestart(m.client)
}

func (m *Admin) admin_list_all_plugins(cmd string, msg xmpp.Chat) {
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
	ReplyAuto(m.client, msg, txt)
}

func (m *Admin) admin_list_plugins(cmd string, msg xmpp.Chat) {
	var names []string
	for _, v := range plugins {
		names = append(names, v.GetName()+" -- "+v.GetSummary())
	}
	txt := "==运行中插件列表==\n" + strings.Join(names, "\n")
	ReplyAuto(m.client, msg, txt)
}

func (m *Admin) admin_disable_plugin(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if tokens[1] == m.Name {
		ReplyAuto(m.client, msg, m.Name+"是内置模块，不允许禁用")
	} else {
		PluginRemove(tokens[1])
	}
}

func (m *Admin) admin_enable_plugin(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	PluginAdd(tokens[1], m.client)
}

func (m *Admin) admin_status(cmd string, msg xmpp.Chat) {
	// cmd is "status chat 正在聊天中..."
	var info = ""
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) == 3 {
		info = tokens[2]
	}
	if IsValidStatus(tokens[1]) {
		SetStatus(m.client, tokens[1], info)
	} else {
		ReplyAuto(m.client, msg, "设置状态失败，有效的状态为: away, chat, dnd, xa.")
	}
}

func (m *Admin) admin_subscribe(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		m.client.RequestSubscription(tokens[1])
	}
}

func (m *Admin) admin_unsubscribe(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		if IsAdmin(tokens[1]) {
			ReplyAuto(m.client, msg, tokens[1]+"是管理员，不允许从好友中删除！")
		} else {
			m.client.RevokeSubscription(tokens[1])
		}
	}
}

func (m *Admin) admin_list_admin(cmd string, msg xmpp.Chat) {
	txt := "==管理员列表==\n" + strings.Join(m.Option["admin"].([]string), "\n")
	ReplyAuto(m.client, msg, txt)
}

func (m *Admin) admin_add_admin(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		if m.IsAdmin(tokens[1]) {
			ReplyAuto(m.client, msg, tokens[1]+" 已是管理员用户，不需再次增加！")
		} else {
			m.client.RequestSubscription(tokens[1])
			m.Option["admin"] = append(m.Option["admin"].([]string), tokens[1])
			ReplyAuto(m.client, msg, "您已添加 "+tokens[1]+"为管理员!")
			jid, _ := SplitJID(msg.Remote)
			SendAuto(m.client, tokens[1], jid+" 临时添加您为管理员!")
		}
	}
}

func (m *Admin) admin_del_admin(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	jid, _ := SplitJID(msg.Remote)
	if IsAdmin(tokens[1]) && tokens[1] != jid {
		m.Option["admin"] = ListDelete(m.Option["admin"].([]string), tokens[1])
		SendAuto(m.client, tokens[1], jid+" 临时取消了您的管理员身份!")
	} else {
		ReplyAuto(m.client, msg, "不能取消 "+tokens[1]+" 的管理员身份!")
	}
}

func (m *Admin) admin_list_options(cmd string, msg xmpp.Chat) {
	options := map[string]string{}
	for _, mod := range plugins {
		for k, v := range mod.GetOptions() {
			options[mod.GetName()+"."+k] = v
		}
	}
	keys := SortMapKeys(options)

	var opt_list []string
	for _, v := range keys {
		opt_list = append(opt_list, fmt.Sprintf("%s : %15s", v, options[v]))
	}
	txt := "==所有模块可配置选项==\n" + strings.Join(opt_list, "\n")
	ReplyAuto(m.client, msg, txt)
}

func (m *Admin) admin_set_option(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) == 3 {
		modkey := strings.SplitN(tokens[1], ".", 2)
		for _, mod := range plugins {
			if modkey[0] == mod.GetName() {
				mod.SetOption(modkey[1], tokens[2])
			}
		}
	}
}

func (m *Admin) admin_list_rooms(cmd string, msg xmpp.Chat) {
	var opt_list []string
	for k, v := range m.Rooms {
		opt_list = append(opt_list, fmt.Sprintf("%2d: %s as %s", k+1, v.JID, v.Nickname))
	}
	txt := "==聊天室列表==\n" + strings.Join(opt_list, "\n")
	ReplyAuto(m.client, msg, txt)
}

func (m *Admin) admin_join_room(cmd string, msg xmpp.Chat) {
	//"join-room <jid> <nickname> [password]"
	tokens := strings.SplitN(cmd, " ", 4)
	if len(tokens) == 4 {
		room := NewRoom(tokens[1], tokens[2], tokens[3], config.Setup.Status, config.Setup.StatusMessage)
		m.client.JoinProtectedMUC(room.JID, room.Nickname, room.Password)
		m.Rooms = append(m.Rooms, room)
		ReplyAuto(m.client, msg, "已经进入聊天室"+room.JID)
	} else if len(tokens) == 3 {
		room := NewRoom(tokens[1], tokens[2], "", config.Setup.Status, config.Setup.StatusMessage)
		m.client.JoinMUC(room.JID, room.Nickname)
		m.Rooms = append(m.Rooms, room)
		ReplyAuto(m.client, msg, "已经进入聊天室"+room.JID)
	}
}

func (m *Admin) admin_leave_room(cmd string, msg xmpp.Chat) {
	//leave-room <jid>
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		roomid := -1
		for k, room := range m.Rooms {
			if room.JID == tokens[1] {
				m.client.LeaveMUC(room.JID)
				roomid = k
			}
			fmt.Printf("[%s] Join to %s as %s\n", m.Name, room.JID, room.Nickname)
		}
		if roomid != -1 {
			m.Rooms = append(m.Rooms[:roomid], m.Rooms[roomid+1:]...)
			ReplyAuto(m.client, msg, "已经退出群聊"+tokens[1])
		}
	} else {
		ReplyAuto(m.client, msg, "命令参数或聊天室id不正确")
	}
}
