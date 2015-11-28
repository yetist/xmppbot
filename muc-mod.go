package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"strings"
)

type RoomOption struct {
	JID      string
	Nickname string
	Password string
	RoomLog  bool
}

type Muc struct {
	Name   string
	Rooms  []RoomOption
	client *xmpp.Client
}

func NewMuc(name string, opt map[string]interface{}) *Muc {
	var rooms []RoomOption
	for _, i := range opt["rooms"].([]map[string]interface{}) {
		room := RoomOption{
			JID:      i["jid"].(string),
			Nickname: i["nickname"].(string),
			RoomLog:  i["room_log"].(bool),
		}
		if i["password"] != nil {
			room.Password = i["password"].(string)
		}
		rooms = append(rooms, room)
	}
	return &Muc{Name: name, Rooms: rooms}
}

func (m *Muc) GetName() string {
	return m.Name
}

func (m *Muc) GetSummary() string {
	return "聊天室模块，提供--room开头的命令，并自动响应聊天室消息"
}

func (m *Muc) CheckEnv() bool {
	return true
}

// 模块加载时，自动进入聊天室。
func (m *Muc) Begin(client *xmpp.Client) {
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
func (m *Muc) End() {
	for _, room := range m.Rooms {
		m.client.LeaveMUC(room.JID)
		fmt.Printf("[%s] Leave from %s\n", m.Name, room.JID)
	}
}

func (m *Muc) Chat(msg xmpp.Chat) {
	if msg.Type != "groupchat" || len(msg.Text) == 0 {
		return
	}
	//	if config.Bot.Debug {
	//		fmt.Printf("[%s] Chat:%#v\n", m.Name, msg)
	//	}
	for _, v := range m.Rooms {
		// 对bot自己发出的消息直接忽略
		if msg.Remote == v.JID+"/"+v.Nickname {
			return
		}
	}

	tokens := strings.SplitN(msg.Remote, "/", 2)
	m.client.Send(xmpp.Chat{Remote: tokens[0], Type: "groupchat", Text: tokens[1] + " said: " + msg.Text})
	//m.client.Send(xmpp.Chat{Remote: "test@groups.isoft-linux.org", Type: "groupchat", Text: "You said: " + msg.Text})
	//m.client.Send(xmpp.Chat{Remote: msg.Remote, Type: "groupchat", Text: "You said: " + msg.Text, Other: []string{"other", "ok"}})
}

func (m *Muc) Presence(pres xmpp.Presence) {
	//	if config.Bot.Debug {
	//		fmt.Printf("[%s] Presence:%#v\n", m.Name, pres)
	//	}
}

func (m *Muc) Help() {
}
