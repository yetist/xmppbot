package core

import (
	"fmt"
	"github.com/jakecoffman/cron"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/config"
	"github.com/yetist/xmppbot/utils"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"strings"
)

type Bot struct {
	client       *xmpp.Client
	cron         *cron.Cron
	web          *WebServer
	plugins      []BotIface
	admin        AdminIface
	config       config.Config
	createPlugin NewFunc
}

type BotIface interface {
	Help() string
	GetName() string
	GetSummary() string
	CheckEnv() bool
	Start(bot *Bot)
	Stop()
	Restart()
	Chat(chat xmpp.Chat)
	Presence(pres xmpp.Presence)
	GetOptions() map[string]string
	SetOption(key, val string)
}

func NewBot(client *xmpp.Client, config config.Config, f NewFunc) *Bot {
	b := &Bot{
		client:       client,
		cron:         cron.New(),
		config:       config,
		web:          NewWebServer(config.GetWebHost(), config.GetWebPort()),
		createPlugin: f,
	}

	// 自动启用内置插件
	admin := NewAdmin("admin", config)
	b.admin = admin
	b.plugins = append(b.plugins, admin)

	for name, v := range config.GetPlugins() {
		if v["enable"].(bool) { //模块是否被启用
			plugin := f(name, v)
			if plugin != nil && plugin.CheckEnv() { //模块运行环境是否满足
				b.plugins = append(b.plugins, plugin)
			}
		}
	}
	return b
}

func (b *Bot) GetPlugins() []BotIface {
	return b.plugins
}

type NewFunc func(name string, opt map[string]interface{}) BotIface

// Interface(), 初始化并加载所有模块
func (b *Bot) Init(f NewFunc) {
	// 自动启用内置插件
	admin := NewAdmin("admin", b.config)
	b.admin = admin
	b.plugins = append(b.plugins, admin)

	for name, v := range b.config.GetPlugins() {
		if v["enable"].(bool) { //模块是否被启用
			plugin := f(name, v)
			if plugin != nil && plugin.CheckEnv() { //模块运行环境是否满足
				b.plugins = append(b.plugins, plugin)
			}
		}
	}
}

func (b *Bot) Start() {
	for _, v := range b.plugins {
		v.Start(b)
	}

	// 每分钟运行ping
	b.cron.AddFunc("0 0/1 * * * ?", func() { b.client.PingC2S(b.config.GetUsername(), b.config.GetServer()) }, "xmpp ping")
	b.cron.Start()
}

func (b *Bot) Run() {
	go func() {
		for {
			chat, err := b.client.Recv()
			if err != nil {
				log.Fatal(err)
			}
			switch v := chat.(type) {
			case xmpp.Chat:
				b.Chat(v)
			case xmpp.Presence:
				b.Presence(v)
			}
		}
	}()
	b.web.Start()
}

// Interface(), 模块收到消息时的处理
func (b *Bot) Chat(chat xmpp.Chat) {
	for _, v := range b.plugins {
		v.Chat(chat)
	}
}

// Interface(), 模块收到Presence消息时的处理
func (b *Bot) Presence(presence xmpp.Presence) {
	for _, v := range b.plugins {
		v.Presence(presence)
	}
}

// Interface(), 模块卸载时的处理函数
func (b *Bot) Stop() {
	for _, v := range b.plugins {
		v.Stop()
	}
}

// Interface(), 重新载入并初始化各模块
func (b *Bot) Restart() {
	var disable_plugins []string

	// 对正在运行中的插件，调用Restart接口重启
	for name, _ := range b.config.GetPlugins() {
		for _, v := range b.plugins {
			if name == v.GetName() {
				v.Restart()
				continue
			}
		}
		disable_plugins = append(disable_plugins, name)
	}
	// 对禁用的插件，重新启用
	for _, n := range disable_plugins {
		b.AddPlugin(n) //FIXME:
	}
}

//获取模块
func (b *Bot) GetPluginByName(name string) BotIface {
	for _, v := range b.plugins {
		if name == v.GetName() {
			return v
		}
	}
	return nil
}

// 按名称卸载某个模块
func (b *Bot) RemovePlugin(name string) {
	id := -1
	for k, v := range b.plugins {
		if name == v.GetName() {
			v.Stop()
			id = k
		}
	}
	if id > -1 {
		b.plugins = append(b.plugins[:id], b.plugins[id+1:]...)
	}
}

// 按名称加载某个模块
func (b *Bot) AddPlugin(name string) {
	for _, v := range b.plugins {
		if name == v.GetName() {
			return
		}
	}
	for n, v := range b.config.GetPlugins() {
		if n == name && v["enable"].(bool) {
			plugin := b.createPlugin(name, v)
			if plugin != nil && plugin.CheckEnv() { //模块运行环境是否满足
				plugin.Start(b)
				b.plugins = append(b.plugins, plugin)
			}
		}
	}
}

func (b *Bot) JoinMUC(jid, nickname string) {
	b.client.JoinMUC(jid, nickname)
}

func (b *Bot) JoinProtectedMUC(jid, nickname, password string) {
	b.client.JoinProtectedMUC(jid, nickname, password)
}

func (b *Bot) LeaveMUC(jid string) {
	b.client.LeaveMUC(jid)
}

func (b *Bot) Roster() error {
	return b.client.Roster()
}

func (b *Bot) ApproveSubscription(jid string) {
	b.client.ApproveSubscription(jid)
}

func (b *Bot) RevokeSubscription(jid string) {
	b.client.RevokeSubscription(jid)
}

func (b *Bot) RequestSubscription(jid string) {
	b.client.RequestSubscription(jid)
}

// 设置状态消息
func (b *Bot) SetStatus(status, info string) (n int, err error) {
	return b.client.SendOrg(fmt.Sprintf("<presence xml:lang='en'><show>%s</show><status>%s</status></presence>", status, info))
}

func (b *Bot) InviteToMUC(jid, roomid, reason string) {
	if b.IsRoomID(roomid) {
		var nick, password string
		for _, v := range b.admin.GetRooms() {
			if roomid == v.JID {
				nick = v.GetNick()
				password = v.GetPassword()
			}
		}
		b.client.InviteToMUC(b.config.GetUsername(), nick, jid, roomid, password, reason)
	}
}

func (b *Bot) SendHtml(chat xmpp.Chat) {
	text := strings.Replace(chat.Text, "&", "&amp;", -1)
	org := fmt.Sprintf("<message to='%s' type='%s' xml:lang='en'><body>%s</body>"+
		"<html xmlns='http://jabber.org/protocol/xhtml-im'><body xmlns='http://www.w3.org/1999/xhtml'>%s</body></html></message>",
		html.EscapeString(chat.Remote), html.EscapeString(chat.Type), html.EscapeString(chat.Text), text)
	b.client.SendOrg(org)
}

// 回复好友消息，或聊天室私聊消息
func (b *Bot) ReplyAuto(recv xmpp.Chat, text string) {
	if strings.Contains(text, "<a href") || strings.Contains(text, "<img") {
		b.SendHtml(xmpp.Chat{Remote: recv.Remote, Type: "chat", Text: text})
	} else {
		b.client.Send(xmpp.Chat{Remote: recv.Remote, Type: "chat", Text: text})
	}
}

// 回复好友消息，或聊天室公共消息
func (b *Bot) ReplyPub(recv xmpp.Chat, text string) {
	if recv.Type == "groupchat" {
		roomid, _ := utils.SplitJID(recv.Remote)
		if strings.Contains(text, "<a href") || strings.Contains(text, "<img") {
			b.SendHtml(xmpp.Chat{Remote: roomid, Type: recv.Type, Text: text})
		} else {
			b.client.Send(xmpp.Chat{Remote: roomid, Type: recv.Type, Text: text})
		}
	} else {
		b.ReplyAuto(recv, text)
	}
}

// 发送到好友消息，或聊天室私聊消息
func (b *Bot) SendAuto(to, text string) {
	if strings.Contains(text, "<a href") || strings.Contains(text, "<img") {
		b.SendHtml(xmpp.Chat{Remote: to, Type: "chat", Text: text})
	} else {
		b.client.Send(xmpp.Chat{Remote: to, Type: "chat", Text: text})
	}
}

// 发送聊天室公共消息
func (b *Bot) SendPub(to, text string) {
	if strings.Contains(text, "<a href") || strings.Contains(text, "<img") {
		b.SendHtml(xmpp.Chat{Remote: to, Type: "groupchat", Text: text})
	} else {
		b.client.Send(xmpp.Chat{Remote: to, Type: "groupchat", Text: text})
	}
}

func (b *Bot) IsAdminID(jid string) bool {
	return b.admin.IsAdminID(jid)
}

func (b *Bot) IsRoomID(jid string) bool {
	roomid, _ := utils.SplitJID(jid)
	for _, v := range b.admin.GetRooms() {
		if roomid == v.JID {
			return true
		}
	}
	return false
}

// 消息是由bot自己发出的吗？
func (b *Bot) SentThis(msg xmpp.Chat) bool {

	if msg.Type == "chat" {
		if id, res := utils.SplitJID(msg.Remote); id == b.config.GetUsername() && res == b.config.GetResource() {
			return true
		}
	} else if msg.Type == "groupchat" {
		for _, v := range b.admin.GetRooms() {
			if msg.Remote == v.JID+"/"+v.Nickname {
				return true
			}
		}
	}
	return false
}

// 此人在聊天中被忽略了吗?
func (b *Bot) BlockRemote(msg xmpp.Chat) bool {
	if msg.Type == "groupchat" {
		for _, v := range b.admin.GetRooms() {
			roomid, nick := utils.SplitJID(msg.Remote)
			if roomid == v.JID && v.IsBlocked(nick) {
				return true
			}
		}
	}
	return false
}

// bot 在群里被点名了吗？
func (b *Bot) Called(msg xmpp.Chat) (ok bool, text string) {
	if msg.Type == "groupchat" {
		for _, v := range b.admin.GetRooms() {
			if strings.Contains(msg.Text, v.Nickname) {
				if strings.HasPrefix(msg.Text, v.Nickname+":") {
					return true, msg.Text[len(v.Nickname)+1:]
				} else {
					return true, strings.Replace(msg.Text, v.Nickname, "", -1)
				}
			}
		}
	}
	return false, msg.Text
}

func (b *Bot) SetRoomNick(r *Room, nick string) (n int, err error) {
	msg := fmt.Sprintf("<presence from='%s/%s' to='%s/%s'/>",
		b.config.GetUsername(), b.config.GetResource(), r.GetJID(), nick)
	if n, err = b.client.SendOrg(msg); err == nil {
		r.SetNick(nick)
	}
	return
}

func (b *Bot) SetRobert(jid string) (n int, err error) {
	msg := fmt.Sprintf("<presence from='%s/%s' to='%s'><caps:c node='http://talk.google.com/xmpp/bot/caps' ver='1.0' xmlns:caps='http://jabber.org/protocol/caps'/></presence>", b.config.GetUsername(), b.config.GetResource(), jid)
	return b.client.SendOrg(msg)
}

func (b *Bot) GetCmdString(cmd string) string {
	return b.admin.GetCmdString(cmd)
}

func (b *Bot) IsCmd(text string) bool {
	return b.admin.IsCmd(text)
}

func (b *Bot) GetCron() *cron.Cron {
	return b.cron
}

func (b *Bot) AddHandler(mod, path string, handler http.HandlerFunc, name string) {
	b.web.Handler("/"+mod+path, handler, utils.GetMd5(mod+name))
}

func (b *Bot) DelHandler(mod, name string) {
	b.web.Destroy(utils.GetMd5(mod + name))
}
