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
	if config.Setup.Debug {
		fmt.Printf("[%s] Chat:%#v\n", m.Name, msg)
	}
	//m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: "You said: " + msg.Text})
}

func (m *Chat) Presence(pres xmpp.Presence) {
}

func (m *Chat) Help() {
}

func (m *Chat) GetOptions() map[string]string {
	return map[string]string{"name": "chat"}
}

func (m *Chat) SetOption(key, val string) {
	println("[muc] set " + key + "=" + val)
}
