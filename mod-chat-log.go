package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
)

type ChatLogger struct {
	Name   string
	Rooms  []*Room
	client *xmpp.Client
}

func NewChatLogger(name string, opt map[string]interface{}) *ChatLogger {
	var rooms []*Room
	for _, i := range opt["rooms"].([]map[string]interface{}) {
		var password string
		jid := i["jid"].(string)
		nickname := i["nickname"].(string)
		if i["password"] == nil {
			password = ""
		} else {
			password = i["password"].(string)
		}
		room := NewRoom(jid, nickname, password)
		rooms = append(rooms, room)
	}
	return &ChatLogger{Name: name, Rooms: rooms}
}

func (m *ChatLogger) GetName() string {
	return m.Name
}

func (m *ChatLogger) GetSummary() string {
	return "聊天室模块，提供--room开头的命令，并自动响应聊天室消息"
}

func (m *ChatLogger) CheckEnv() bool {
	return true
}

// 模块加载时，自动进入聊天室。
func (m *ChatLogger) Begin(client *xmpp.Client) {
	m.client = client
	for _, room := range m.Rooms {
		if len(room.Password) > 0 {
			client.JoinProtectedMUC(room.JID, room.Nickname, room.Password)
		} else {
			client.JoinMUC(room.JID, room.Nickname)
		}
		fmt.Printf("[%s] Join to %s as %s\n", m.Name, room.JID, room.Nickname)
	}
}

// 模块卸载时，自动离开聊天室。
func (m *ChatLogger) End() {
	for _, room := range m.Rooms {
		m.client.LeaveMUC(room.JID)
		fmt.Printf("[%s] Leave from %s\n", m.Name, room.JID)
	}
}

func (m *ChatLogger) Restart() {
	m.End()

	var rooms []*Room
	v := config.Plugin[m.GetName()]
	for _, i := range v["rooms"].([]map[string]interface{}) {
		var password string
		jid := i["jid"].(string)
		nickname := i["nickname"].(string)
		if i["password"] == nil {
			password = ""
		} else {
			password = i["password"].(string)
		}
		room := NewRoom(jid, nickname, password)
		rooms = append(rooms, room)
	}
	m.Rooms = rooms

	m.Begin(m.client)
}

func (m *ChatLogger) Chat(msg xmpp.Chat) {
	if msg.Type != "groupchat" || len(msg.Text) == 0 {
		return
	}
	for _, v := range m.Rooms {
		// 对bot自己发出的消息直接忽略
		if msg.Remote == v.JID+"/"+v.Nickname {
			return
		}
	}
	roomid, nick := SplitJID(msg.Remote)
	SendPub(m.client, roomid, nick+" said: "+msg.Text)
}

func (m *ChatLogger) Presence(pres xmpp.Presence) {
}

func (m *ChatLogger) Help() string {
	return "help for muc"
}

func (m *ChatLogger) GetOptions() map[string]string {
	return map[string]string{"name": "muc"}
}

func (m *ChatLogger) SetOption(key, val string) {
	println("[muc] set " + key + "=" + val)
}
