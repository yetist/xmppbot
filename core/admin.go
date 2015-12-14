package core

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/config"
	"github.com/yetist/xmppbot/utils"
	"strings"
	"time"
)

type Admin struct {
	Name      string
	bot       *Bot
	config    config.Config
	Option    map[string]interface{}
	loginTime time.Time
	Rooms     []*Room
	Friends   []string
	admins    []string
	crons     map[string]CronEntry
}

type AdminIface interface {
	GetRooms() []*Room
	IsAdminID(jid string) bool
	IsFriendID(jid string) bool
	IsCmd(text string) bool
	IsRoomID(jid string) bool
	GetCmdString(cmd string) string
	LoginTime() time.Time
}

type CronEntry struct {
	spec string
	to   string
	text string
}

func NewAdmin(name string, config config.Config) *Admin {
	var rooms []*Room
	for _, i := range config.GetRooms() {
		password := ""
		if i["password"] != nil {
			password = i["password"].(string)
		}
		room := NewRoom(i["jid"].(string), i["nickname"].(string), password)
		rooms = append(rooms, room)
	}
	return &Admin{
		Name:   name,
		Rooms:  rooms,
		config: config,
		admins: config.GetAdmin(),
		crons:  map[string]CronEntry{},
		Option: map[string]interface{}{
			"cmd_prefix":     config.GetCmdPrefix(),
			"auto-subscribe": config.GetAutoSubscribe(),
		},
	}
}

// BotInterface
func (m *Admin) GetName() string {
	return m.Name
}

func (m *Admin) GetSummary() string {
	return "管理员模块[内置]"
}

func (m *Admin) Help() string {
	text := []string{
		m.GetSummary() + ": 提供了基础的机器人管理命令。",
		m.GetCmdString("help") + "    查看帮助命令详情",
		m.GetCmdString("admin") + "    查看管理员命令详情",
		m.GetCmdString("bot") + "    查看机器人命令详情",
		m.GetCmdString("cron") + "    查看计划任务命令详情",
		m.GetCmdString("plugin") + "    查看模块命令详情",
		m.GetCmdString("room") + "    查看聊天室命令详情",
	}
	return strings.Join(text, "\n")
}

func (m *Admin) Description() string {
	msg := []string{m.Help(),
		"本模块可配置属性:",
	}
	options := m.GetOptions()
	keys := utils.SortMapKeys(options)
	for _, v := range keys {
		msg = append(msg, fmt.Sprintf("%-20s : %s", v, options[v]))
	}
	return strings.Join(msg, "\n")
}

func (m *Admin) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		switch k {
		case "cmd_prefix":
			opts[k] = v.(string) + "  #命令前缀"
		case "auto-subscribe":
			opts[k] = utils.BoolToString(v.(bool)) + "  #是否自动完成互加好友"
		}
	}
	return opts
}

func (m *Admin) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		if key == "auto-subscribe" {
			m.Option[key] = utils.StringToBool(val)
		} else {
			m.Option[key] = val
		}
	}
}

func (m *Admin) CheckEnv() bool {
	return true
}

func (m *Admin) Start(bot *Bot) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.bot = bot
	m.loginTime = time.Now()
	m.bot.Roster()
	for _, room := range m.Rooms {
		if len(room.Password) > 0 {
			m.bot.JoinProtectedMUC(room.JID, room.Nickname, room.Password)
		} else {
			m.bot.JoinMUC(room.JID, room.Nickname)
		}
		fmt.Printf("[%s] Join to %s as %s\n", m.Name, room.JID, room.Nickname)
	}
}

func (m *Admin) Stop() {
	for _, room := range m.Rooms {
		m.bot.LeaveMUC(room.JID)
		fmt.Printf("[%s] Leave from %s\n", m.Name, room.JID)
	}
}

func (m *Admin) Restart() {
	m.Stop()
	m.Option["cmd_prefix"] = m.config.GetCmdPrefix()
	m.Option["auto-subscribe"] = m.config.GetAutoSubscribe()
	m.bot.Roster()
	m.bot.SetStatus(m.config.GetStatus(), m.config.GetStatusMessage())

	var rooms []*Room
	for _, i := range m.config.GetRooms() {
		password := ""
		if i["password"] != nil {
			password = i["password"].(string)
		}
		room := NewRoom(i["jid"].(string), i["nickname"].(string), password)
		rooms = append(rooms, room)
	}
	m.Rooms = rooms
	m.Start(m.bot)
}

func (m *Admin) Chat(msg xmpp.Chat) {
	if msg.Type == "roster" {
		for _, v := range msg.Roster {
			fmt.Printf("%#v\n", v)
			if !m.IsFriendID(v.Remote) {
				m.Friends = append(m.Friends, v.Remote)
			}
			m.bot.SetRobert(v.Remote)
		}
	}

	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}

	// 仅处理好友消息
	if strings.HasPrefix(msg.Text, m.GetCmdString("help")) {
		cmd := strings.TrimSpace(msg.Text[len(m.GetCmdString("help")):])
		m.HelpCommand(cmd, msg)
	} else if strings.HasPrefix(msg.Text, m.GetCmdString("room")) {
		cmd := strings.TrimSpace(msg.Text[len(m.GetCmdString("room")):])
		m.RoomCommand(cmd, msg)
	} else if strings.HasPrefix(msg.Text, m.GetCmdString("cron")) {
		cmd := strings.TrimSpace(msg.Text[len(m.GetCmdString("cron")):])
		m.CronCommand(cmd, msg)
	} else if strings.HasPrefix(msg.Text, m.GetCmdString("bot")) {
		cmd := strings.TrimSpace(msg.Text[len(m.GetCmdString("bot")):])
		m.BotCommand(cmd, msg)
	} else if strings.HasPrefix(msg.Text, m.GetCmdString("admin")) {
		cmd := strings.TrimSpace(msg.Text[len(m.GetCmdString("admin")):])
		m.AdminCommand(cmd, msg)
	} else if strings.HasPrefix(msg.Text, m.GetCmdString("plugin")) {
		cmd := strings.TrimSpace(msg.Text[len(m.GetCmdString("plugin")):])
		m.PluginCommand(cmd, msg)
	}
}

func (m *Admin) Presence(pres xmpp.Presence) {
	if m.config.GetDebug() {
		fmt.Printf("[%s] Presence:%#v\n", m.Name, pres)
	}
	//处理订阅消息
	if pres.Type == "subscribe" {
		if m.Option["auto-subscribe"].(bool) {
			m.bot.ApproveSubscription(pres.From)
			m.bot.RequestSubscription(pres.From)
		} else {
			m.bot.RevokeSubscription(pres.From)
		}
	}
}

// AdminInterface
func (m *Admin) GetRooms() []*Room {
	return m.Rooms
}

func (m *Admin) IsAdminID(jid string) bool {
	u, _ := utils.SplitJID(jid)
	for _, admin := range m.admins {
		if u == admin {
			return true
		}
	}
	return false
}

func (m *Admin) IsSysAdminID(jid string) bool {
	u, _ := utils.SplitJID(jid)
	for _, admin := range m.config.Setup.Admin {
		if u == admin {
			return true
		}
	}
	return false
}

func (m *Admin) IsFriendID(jid string) bool {
	u, _ := utils.SplitJID(jid)
	for _, friend := range m.Friends {
		if u == friend {
			return true
		}
	}
	return false
}

func (m *Admin) LoginTime() time.Time {
	return m.loginTime.UTC()
}

// jid 是已进入的聊天室吗？
func (m *Admin) IsRoomID(jid string) bool {
	for _, v := range m.Rooms {
		if v.JID == jid {
			return true
		}
	}
	return false
}

func (m *Admin) IsCmd(text string) bool {
	return strings.HasPrefix(strings.TrimSpace(text), m.Option["cmd_prefix"].(string))
}
func (m *Admin) GetCmdString(cmd string) string {
	return m.Option["cmd_prefix"].(string) + cmd
}

/* help 命令 */
func (m *Admin) HelpCommand(cmd string, msg xmpp.Chat) {
	if cmd == "" {
		m.cmd_help_all(cmd, msg)
	} else if len(cmd) > 0 {
		m.cmd_help_plugin(cmd, msg)
	} else {
		m.bot.ReplyAuto(msg, "不支持的命令: "+cmd)
	}
}

func (m *Admin) cmd_help_all(cmd string, msg xmpp.Chat) {
	help_msg := []string{"==所有模块帮助=="}
	for _, v := range m.bot.GetPlugins() {
		help_msg = append(help_msg, "=="+v.GetName()+"模块==")
		help_msg = append(help_msg, v.Help())
	}
	m.bot.ReplyAuto(msg, strings.Join(help_msg, "\n"))
}

func (m *Admin) cmd_help_plugin(cmd string, msg xmpp.Chat) {
	var helps []string
	names := strings.Split(cmd, " ")
	for _, name := range names {
		if v := m.bot.GetPluginByName(name); v != nil {
			helps = append(helps, "=="+v.GetName()+"帮助==")
			helps = append(helps, v.Help())
		}
	}
	m.bot.ReplyAuto(msg, strings.Join(helps, "\n"))
}

/* room 命令 */
func (m *Admin) RoomCommand(cmd string, msg xmpp.Chat) {
	if cmd == "" || cmd == "help" {
		m.room_help(cmd, msg)
	} else if strings.HasPrefix(cmd, "send ") {
		m.room_send(cmd, msg)
	} else if strings.HasPrefix(cmd, "nick ") {
		m.room_nick(cmd, msg)
	} else if strings.HasPrefix(cmd, "invite ") {
		m.room_invite(cmd, msg)
	} else if strings.HasPrefix(cmd, "list-blocks ") {
		m.room_list_blocks(cmd, msg)
	} else if strings.HasPrefix(cmd, "block ") {
		m.room_block(cmd, msg)
	} else if strings.HasPrefix(cmd, "unblock ") {
		m.room_unblock(cmd, msg)
	} else if cmd == "list" {
		m.room_list(cmd, msg)
	} else if strings.HasPrefix(cmd, "join ") {
		m.room_join(cmd, msg)
	} else if strings.HasPrefix(cmd, "leave ") {
		m.room_leave(cmd, msg)
	} else {
		m.bot.ReplyAuto(msg, "不支持的命令: "+cmd)
	}
}

func (m *Admin) room_help(cmd string, msg xmpp.Chat) {

	help_msg := []string{"==聊天室命令==",
		m.GetCmdString("room") + " help                         显示本信息",
		m.GetCmdString("room") + " send <Rid|all> <Message>     让机器人在聊天室中发送消息msg",
		m.GetCmdString("room") + " nick <Rid|all> <NickName>    修改机器人在聊天室的昵称为NickName",
		m.GetCmdString("room") + " invite <jid> <Rid> [Reason]  邀请某人进入聊天室", "",

		m.GetCmdString("room") + " list-blocks <Rid|all>    查看聊天室屏蔽列表",
		m.GetCmdString("room") + " block <Rid|all> <Who>    屏蔽who，对who发送的消息不响应",
		m.GetCmdString("room") + " unblock <Rid|all> <Who>  重新对who发送的消息进行响应", "",

		m.GetCmdString("room") + " list                          列出机器人当前所在的聊天室",
		m.GetCmdString("room") + " join <Rid> <Nick> [Password]  加入聊天室",
		m.GetCmdString("room") + " leave <Rid>                   离开聊天室", "",
		"注: Rid 请使用聊天室jid, all表示所有的聊天室。",
	}
	m.bot.ReplyAuto(msg, strings.Join(help_msg, "\n"))
}

func (m *Admin) room_send(cmd string, msg xmpp.Chat) {
	//"send <jid|all> <msg>":       "让机器人在聊天室中发送消息msg",
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) != 3 {
		return
	}
	if tokens[1] == "all" {
		for _, v := range m.Rooms {
			m.bot.SendPub(v.JID, tokens[2])
		}
	} else {
		for _, v := range m.Rooms {
			if v.JID == tokens[1] {
				m.bot.SendPub(tokens[1], tokens[2])
				return
			}
		}
	}
}

// 修改bot在聊天室中的昵称．
func (m *Admin) room_nick(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) != 3 {
		return
	}
	if tokens[1] == "all" {
		for _, v := range m.Rooms {
			m.bot.SetRoomNick(v, tokens[2])
		}
	} else {
		for _, v := range m.Rooms {
			if v.JID == tokens[1] {
				m.bot.SetRoomNick(v, tokens[2])
				return
			}
		}
	}
}

func (m *Admin) room_invite(cmd string, msg xmpp.Chat) {
	//invite <jid> <roomid> [reason] 修改机器人在聊天室的昵称为NickName", "",
	var jid, roomid, reason string
	if tokens := strings.SplitN(cmd, " ", 4); len(tokens) == 4 {
		jid = tokens[1]
		roomid = tokens[2]
		reason = tokens[3]
	} else if tokens := strings.SplitN(cmd, " ", 3); len(tokens) == 3 {
		jid = tokens[1]
		roomid = tokens[2]
	} else {
		return
	}
	if !m.IsFriendID(jid) {
		return
	}
	m.bot.InviteToMUC(jid, roomid, reason)
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
				m.bot.ReplyAuto(msg, "Bot未进入此聊天室")
			}
		}
	}
	m.bot.ReplyAuto(msg, strings.Join(blocks, "\n"))
}

func (m *Admin) room_block(cmd string, msg xmpp.Chat) {
	//"block <jid|all> <who>":     "屏蔽who，对who发送的消息不响应",
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) != 3 {
		return
	}
	if tokens[1] == "all" {
		for _, v := range m.Rooms {
			m.bot.SendPub(v.JID, "/me 忽略了 "+tokens[2]+" 的消息")
			v.BlockOne(tokens[2])
		}
	} else {
		for _, v := range m.Rooms {
			if v.JID == tokens[1] {
				m.bot.SendPub(v.JID, "/me 忽略了 "+tokens[2]+" 的消息")
				v.BlockOne(tokens[2])
			} else {
				m.bot.ReplyAuto(msg, "Bot未进入此聊天室")
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
			m.bot.SendPub(v.JID, "/me 开始关注 "+tokens[2]+" 的消息")
			v.UnBlockOne(tokens[2])
		}
	} else {
		for _, v := range m.Rooms {
			if v.JID == tokens[1] {
				m.bot.SendPub(v.JID, "/me 开始关注 "+tokens[2]+" 的消息")
				v.UnBlockOne(tokens[2])
			} else {
				m.bot.ReplyAuto(msg, "Bot未进入此聊天室")
			}
		}
	}
}

func (m *Admin) room_list(cmd string, msg xmpp.Chat) {
	//list
	var opt_list []string
	for k, v := range m.Rooms {
		opt_list = append(opt_list, fmt.Sprintf("%2d: %s as %s", k+1, v.JID, v.Nickname))
	}
	txt := "==聊天室列表==\n" + strings.Join(opt_list, "\n")
	m.bot.ReplyAuto(msg, txt)
}

func (m *Admin) room_join(cmd string, msg xmpp.Chat) {
	//join <jid> <nickname> [password]"
	tokens := strings.SplitN(cmd, " ", 4)
	if len(tokens) == 4 {
		room := NewRoom(tokens[1], tokens[2], tokens[3])
		m.bot.JoinProtectedMUC(room.JID, room.Nickname, room.Password)
		m.Rooms = append(m.Rooms, room)
		m.bot.ReplyAuto(msg, "已经进入聊天室"+room.JID)
	} else if len(tokens) == 3 {
		room := NewRoom(tokens[1], tokens[2], "")
		m.bot.JoinMUC(room.JID, room.Nickname)
		m.Rooms = append(m.Rooms, room)
		m.bot.ReplyAuto(msg, "已经进入聊天室"+room.JID)
	}
}

func (m *Admin) room_leave(cmd string, msg xmpp.Chat) {
	//leave <jid>
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		roomid := -1
		for k, room := range m.Rooms {
			if room.JID == tokens[1] {
				m.bot.LeaveMUC(room.JID)
				roomid = k
			}
			fmt.Printf("[%s] Join to %s as %s\n", m.Name, room.JID, room.Nickname)
		}
		if roomid != -1 {
			m.Rooms = append(m.Rooms[:roomid], m.Rooms[roomid+1:]...)
			m.bot.ReplyAuto(msg, "已经退出群聊"+tokens[1])
		}
	} else {
		m.bot.ReplyAuto(msg, "命令参数或聊天室id不正确")
	}
}

/* cron 命令处理 */
func (m *Admin) CronCommand(cmd string, msg xmpp.Chat) {
	if cmd == "" || cmd == "help" {
		m.cron_help(cmd, msg)
	} else if cmd == "list" {
		m.cron_list(cmd, msg)
	} else if strings.HasPrefix(cmd, "add ") {
		m.cron_add(cmd, msg)
	} else if strings.HasPrefix(cmd, "del ") {
		m.cron_del(cmd, msg)
	} else {
		m.bot.ReplyAuto(msg, "不支持的命令: "+cmd)
	}
}

func (m *Admin) cron_help(cmd string, msg xmpp.Chat) {
	help_msg := []string{"==计划任务命令==",
		m.GetCmdString("cron") + " help                      显示本信息",
		m.GetCmdString("cron") + " list                      列出所有的计划任务详情",
		m.GetCmdString("cron") + " add <Spec> <jid> <msg>    添加计划任务",
		"  Spec: Seconds Minutes Hours DayofMonth Month DayofWeek",
		m.GetCmdString("cron") + " del <taskid>              删除计划任务",
	}
	m.bot.ReplyAuto(msg, strings.Join(help_msg, "\n"))
}

func (m *Admin) cron_list(cmd string, msg xmpp.Chat) {
	names := []string{"==所有计划任务列表=="}
	for k, c := range m.crons {
		names = append(names, fmt.Sprintf("TaskID: %s , [%s] => [%s] : %s", k, c.spec, c.to, c.text))
	}
	m.bot.ReplyAuto(msg, strings.Join(names, "\n"))
}

func (m *Admin) cron_add(cmd string, msg xmpp.Chat) {
	//add <spec6> <jid> <msg>
	tokens := strings.SplitN(cmd, " ", 9)
	if len(tokens) != 9 {
		m.bot.ReplyAuto(msg, "添加新任务失败，请检查消息格式是否正确．")
	}
	cron := m.bot.GetCron()
	to := tokens[7]
	message := tokens[8]
	spec := strings.Join(tokens[1:7], " ")
	id := utils.GetMd5(cmd)
	if m.IsRoomID(to) {
		cron.AddFunc(spec, func() { m.bot.SendPub(to, message) }, id)
		m.crons[id] = CronEntry{spec: spec, to: to, text: message}
	} else {
		cron.AddFunc(spec, func() { m.bot.SendAuto(to, message) }, id)
		m.crons[id] = CronEntry{spec: spec, to: to, text: message}
	}
}

func (m *Admin) cron_del(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	key := tokens[1]
	cron := m.bot.GetCron()
	for k, _ := range m.crons {
		if key == k {
			cron.RemoveJob(key)
			delete(m.crons, key)
			return
		}
	}
}

/* bot 命令 */
func (m *Admin) BotCommand(cmd string, msg xmpp.Chat) {
	if !m.IsAdminID(msg.Remote) {
		m.bot.ReplyAuto(msg, "请确认您是管理员，并且通过好友消息发送了此命令。")
		return
	}
	if cmd == "" || cmd == "help" {
		m.bot_help(cmd, msg)
	} else if cmd == "restart" {
		m.bot_restart(cmd, msg)
	} else if strings.HasPrefix(cmd, "status ") {
		m.bot_status(cmd, msg)
	} else if strings.HasPrefix(cmd, "send ") {
		m.bot_send(cmd, msg)
	} else if cmd == "friends" {
		m.bot_friends(cmd, msg)
	} else if strings.HasPrefix(cmd, "subscribe ") {
		m.bot_subscribe(cmd, msg)
	} else if strings.HasPrefix(cmd, "unsubscribe ") {
		m.bot_unsubscribe(cmd, msg)
	} else {
		m.bot.ReplyAuto(msg, "不支持的命令: "+cmd)
	}
}

func (m *Admin) bot_help(cmd string, msg xmpp.Chat) {
	help_msg := []string{"==管理员命令==",
		m.GetCmdString("bot") + " help                      显示本信息",
		m.GetCmdString("bot") + " restart                   重新载入配置文件，初始化各模块",
		m.GetCmdString("bot") + " status <status> [message] 设置机器人在线状态",
		m.GetCmdString("bot") + " send <jid> [message]      给好友发送消息", "",

		m.GetCmdString("bot") + " friends           列出好友帐号",
		m.GetCmdString("bot") + " subscribe <jid>   新增好友帐号",
		m.GetCmdString("bot") + " unsubscribe <jid> 删除好友帐号",
	}
	m.bot.ReplyAuto(msg, strings.Join(help_msg, "\n"))
}

func (m *Admin) bot_restart(cmd string, msg xmpp.Chat) {
	m.Restart() //重启内置插件
	m.bot.Restart()
}

func (m *Admin) bot_send(cmd string, msg xmpp.Chat) {
	//"send <jid> <msg>":       "让机器人给好友发送消息msg",
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) != 3 {
		return
	}
	if m.IsFriendID(tokens[1]) {
		m.bot.SendAuto(tokens[1], tokens[2])
	}
}

func (m *Admin) bot_status(cmd string, msg xmpp.Chat) {
	// cmd is "status chat 正在聊天中..."
	var info = ""
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) == 3 {
		info = tokens[2]
	}
	if utils.IsValidStatus(tokens[1]) {
		m.bot.SetStatus(tokens[1], info)
	} else {
		m.bot.ReplyAuto(msg, "设置状态失败，有效的状态为: away, chat, dnd, xa.")
	}
}

func (m *Admin) bot_friends(cmd string, msg xmpp.Chat) {
	txt := "==好友列表==\n" + strings.Join(m.Friends, "\n")
	m.bot.ReplyAuto(msg, txt)
}

func (m *Admin) bot_subscribe(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		if !m.IsFriendID(tokens[1]) {
			m.bot.RequestSubscription(tokens[1])
		} else {
			m.bot.ReplyAuto(msg, tokens[1]+"已经是好友，不需要多次增加！")
		}
	}
}

func (m *Admin) bot_unsubscribe(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		if !m.IsFriendID(tokens[1]) {
			m.bot.ReplyAuto(msg, tokens[1]+"不是好友，不需要删除！")
			return
		}
		jid, _ := utils.SplitJID(msg.Remote)
		if tokens[1] == jid {
			m.bot.ReplyAuto(msg, tokens[1]+"是你的id, 不支持这个操作！")
			return
		}

		if m.IsSysAdminID(tokens[1]) {
			m.bot.ReplyAuto(msg, "不允许删除超级管理员帐号 "+tokens[1]+"！")
			return
		}
		m.Friends = utils.ListDelete(m.Friends, tokens[1])
		m.bot.RevokeSubscription(tokens[1])
		if m.IsAdminID(tokens[1]) {
			m.admins = utils.ListDelete(m.admins, tokens[1])
			m.bot.ReplyAuto(msg, "将管理员帐号 "+tokens[1]+" 从好友中删除！")
		} else {
			m.bot.ReplyAuto(msg, "将帐号 "+tokens[1]+" 从好友中删除！")
		}
	}
}

/* admin 命令 */
func (m *Admin) AdminCommand(cmd string, msg xmpp.Chat) {
	if !m.IsAdminID(msg.Remote) {
		m.bot.ReplyAuto(msg, "请确认您是管理员，并且通过好友消息发送了此命令。")
		return
	}
	if cmd == "" || cmd == "help" {
		m.admin_help(cmd, msg)
	} else if cmd == "list" {
		m.admin_list(cmd, msg)
	} else if strings.HasPrefix(cmd, "add ") {
		m.admin_add(cmd, msg)
	} else if strings.HasPrefix(cmd, "del ") {
		m.admin_del(cmd, msg)
	} else {
		m.bot.ReplyAuto(msg, "不支持的命令: "+cmd)
	}
}

func (m *Admin) admin_help(cmd string, msg xmpp.Chat) {
	help_msg := []string{"==管理员命令==",
		m.GetCmdString("admin") + " list       列出管理员帐号",
		m.GetCmdString("admin") + " add <jid>  新增管理员帐号",
		m.GetCmdString("admin") + " del <jid>  删除管理员帐号", "",
	}
	m.bot.ReplyAuto(msg, strings.Join(help_msg, "\n"))
}

func (m *Admin) admin_list(cmd string, msg xmpp.Chat) {
	txt := "==管理员列表==\n" + strings.Join(m.admins, "\n")
	m.bot.ReplyAuto(msg, txt)
}

func (m *Admin) admin_add(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 && strings.Contains(tokens[1], "@") {
		if m.IsAdminID(tokens[1]) {
			m.bot.ReplyAuto(msg, tokens[1]+" 已是管理员用户，不需再次增加！")
		} else {
			if !m.IsFriendID(tokens[1]) {
				m.bot.RequestSubscription(tokens[1])
			}
			m.admins = append(m.admins, tokens[1])
			m.bot.ReplyAuto(msg, "您已添加 "+tokens[1]+"为管理员!")
			jid, _ := utils.SplitJID(msg.Remote)
			m.bot.SendAuto(tokens[1], jid+" 添加您为临时管理员!")
		}
	}
}

func (m *Admin) admin_del(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	jid, _ := utils.SplitJID(msg.Remote)
	if m.IsAdminID(tokens[1]) && tokens[1] != jid {
		m.admins = utils.ListDelete(m.admins, tokens[1])
		m.bot.SendAuto(tokens[1], jid+" 临时取消了您的管理员身份!")
	} else {
		m.bot.ReplyAuto(msg, "不能取消 "+tokens[1]+" 的管理员身份!")
	}
}

/* plugin 命令 */
func (m *Admin) PluginCommand(cmd string, msg xmpp.Chat) {
	if !m.IsAdminID(msg.Remote) {
		m.bot.ReplyAuto(msg, "请确认您是管理员，并且通过好友消息发送了此命令。")
		return
	}
	if cmd == "" || cmd == "help" {
		m.plugin_help(cmd, msg)
	} else if cmd == "all" {
		m.plugin_all(cmd, msg)
	} else if cmd == "list" {
		m.plugin_list(cmd, msg)
	} else if strings.HasPrefix(cmd, "disable ") {
		m.plugin_disable(cmd, msg)
	} else if strings.HasPrefix(cmd, "enable ") {
		m.plugin_enable(cmd, msg)
	} else if strings.HasPrefix(cmd, "get") {
		m.plugin_get(cmd, msg)
	} else if strings.HasPrefix(cmd, "set ") {
		m.plugin_set(cmd, msg)
	} else {
		m.bot.ReplyAuto(msg, "不支持的命令: "+cmd)
	}
}

func (m *Admin) plugin_help(cmd string, msg xmpp.Chat) {
	help_msg := []string{"==插件命令==",
		m.GetCmdString("plugin") + " help                 显示本信息",
		m.GetCmdString("plugin") + " all                  列出所有的模块",
		m.GetCmdString("plugin") + " list                 列出当前启用的模块",
		m.GetCmdString("plugin") + " disable <Plugin>     禁用模块",
		m.GetCmdString("plugin") + " enable <Plugin>      启用模块",
		m.GetCmdString("plugin") + " get [Plugin]         列出模块属性",
		m.GetCmdString("plugin") + " set <field> <value>  设置模块属性",
	}
	m.bot.ReplyAuto(msg, strings.Join(help_msg, "\n"))
}

func (m *Admin) plugin_all(cmd string, msg xmpp.Chat) {
	names := []string{"==所有插件列表=="}

	names = append(names, m.Name+"[内置]")

	for name, v := range m.config.GetPlugins() {
		if v["enable"].(bool) {
			names = append(names, name+"[启用]")
		} else {
			names = append(names, name+"[禁用]")
		}
	}
	m.bot.ReplyAuto(msg, strings.Join(names, "\n"))
}

func (m *Admin) plugin_list(cmd string, msg xmpp.Chat) {
	names := []string{"==运行中插件列表=="}

	for _, v := range m.bot.GetPlugins() {
		names = append(names, v.GetName()+" -- "+v.GetSummary())
	}
	m.bot.ReplyAuto(msg, strings.Join(names, "\n"))
}

func (m *Admin) plugin_disable(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if tokens[1] == m.Name {
		m.bot.ReplyAuto(msg, m.Name+"是内置模块，不允许禁用")
	} else {
		m.bot.RemovePlugin(tokens[1])
	}
}

func (m *Admin) plugin_enable(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	m.bot.AddPlugin(tokens[1])
}

func (m *Admin) plugin_get(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	options := map[string]string{}
	var opt_list []string
	if len(tokens) == 2 {
		if mod := m.bot.GetPluginByName(tokens[1]); mod != nil {
			opt_list = append(opt_list, "=="+mod.GetName()+"模块属性==")
			for k, v := range mod.GetOptions() {
				options[mod.GetName()+"."+k] = v
			}
		}
	} else {
		opt_list = append(opt_list, "==所有模块属性==")
		for _, mod := range m.bot.GetPlugins() {
			for k, v := range mod.GetOptions() {
				options[mod.GetName()+"."+k] = v
			}
		}
	}

	keys := utils.SortMapKeys(options)
	for _, v := range keys {
		opt_list = append(opt_list, fmt.Sprintf("%-20s : %s", v, options[v]))
	}
	m.bot.ReplyAuto(msg, strings.Join(opt_list, "\n"))
}

func (m *Admin) plugin_set(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 3)
	if len(tokens) == 3 {
		modkey := strings.SplitN(tokens[1], ".", 2)
		for _, mod := range m.bot.GetPlugins() {
			if modkey[0] == mod.GetName() {
				mod.SetOption(modkey[1], tokens[2])
			}
		}
	}
}
