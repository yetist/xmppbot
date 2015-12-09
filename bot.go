package main

import (
	"fmt"
	"github.com/jakecoffman/cron"
	"github.com/mattn/go-xmpp"
	"golang.org/x/net/html"
	"log"
	"strings"
)

type Bot struct {
	client       *xmpp.Client
	cron         *cron.Cron
	web          *WebServer
	plugins      []BotInterface
	admin_plugin AdminInterface
}

type BotInterface interface {
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

func NewBot() (bot *Bot, err error) {
	var client *xmpp.Client

	if client, err = NewClient(); err != nil {
		return
	}
	return &Bot{
		client: client,
		cron:   cron.New(),
		web:    NewWebServer(3000),
	}, nil
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

func (b *Bot) GetPlugins() []BotInterface {
	return b.plugins
}

// Interface(), 初始化并加载所有模块
func (b *Bot) Init() {
	// 自动启用内置插件
	admin := NewAdmin("admin")
	b.plugins = append(b.plugins, admin)
	b.admin_plugin = admin

	for name, v := range config.Plugin {
		if v["enable"].(bool) { //模块是否被启用
			plugin := CreatePlugin(name, v)
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
	b.cron.AddFunc("* */1 * * * ?", func() { b.client.PingC2S(config.Account.Username, config.Account.Server) }, "xmpp ping")
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
	b.web.InitDispatch()
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
	for name, _ := range config.Plugin {
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

//获取管理员模块
func (b *Bot) GetAdminPlugin() AdminInterface {
	return b.admin_plugin
}

//获取模块
func (b *Bot) GetPluginByName(name string) BotInterface {
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
	for n, v := range config.Plugin {
		if n == name && v["enable"].(bool) {
			plugin := CreatePlugin(name, v)
			if plugin != nil && plugin.CheckEnv() { //模块运行环境是否满足
				plugin.Start(b)
				b.plugins = append(b.plugins, plugin)
			}
		}
	}
}

//TODO: delete the function
func (b *Bot) GetClient() *xmpp.Client {
	return b.client
}

// 设置状态消息
func (b *Bot) SetStatus(status, info string) (n int, err error) {
	return b.client.SendOrg(fmt.Sprintf("<presence xml:lang='en'><show>%s</show><status>%s</status></presence>", status, info))
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
		roomid, _ := SplitJID(recv.Remote)
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
	u, _ := SplitJID(jid)
	for _, admin := range config.Setup.Admin {
		if u == admin {
			return true
		}
	}
	return false
}

// 消息是由bot自己发出的吗？
func (b *Bot) SendThis(msg xmpp.Chat) bool {

	if msg.Type == "chat" {
		if id, res := SplitJID(msg.Remote); id == config.Account.Username && res == config.Account.Resource {
			return true
		}
	} else if msg.Type == "groupchat" {
		for _, v := range b.admin_plugin.GetRooms() {
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
		for _, v := range b.admin_plugin.GetRooms() {
			roomid, nick := SplitJID(msg.Remote)
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
		for _, v := range b.admin_plugin.GetRooms() {
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
