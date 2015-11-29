package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
)

type Chat struct {
	Name   string
	Fuck   string
	Random string
	client *xmpp.Client
}

func NewChat(name string, opt map[string]interface{}) *Chat {
	return &Chat{
		Name:   name,
		Fuck:   opt["fuck"].(string),
		Random: opt["random"].(string),
	}
}

func (m *Chat) GetName() string {
	return m.Name
}

func (m *Chat) GetSummary() string {
	return "好友模块，提供--bot开头的命令"
}

func (m *Chat) CheckEnv() bool {
	return true
}

func (m *Chat) Begin(client *xmpp.Client) {
	m.client = client
}

func (m *Chat) End() {
}

func (m *Chat) Restart() {
}

func (m *Chat) Chat(msg xmpp.Chat) {
	//if msg.Type != "chat" || len(msg.Text) == 0 {
	//	return
	//}
	if config.Bot.Debug {
		fmt.Printf("[%s] Chat:%#v\n", m.Name, msg)
	}
	//m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "You said: " + msg.Text})
}

func (m *Chat) Presence(pres xmpp.Presence) {
	//	if config.Bot.Debug {
	//		fmt.Printf("[%s] Presence:%#v\n", m.Name, pres)
	//	}
	//	//处理订阅消息
	//	if pres.Type == "subscribe" {
	//		if config.Bot.AllowFriends {
	//			m.client.ApproveSubscription(pres.From)
	//			m.client.RequestSubscription(pres.From)
	//		} else {
	//			m.client.RevokeSubscription(pres.From)
	//		}
	//	}
}

func (m *Chat) Help() {
}
