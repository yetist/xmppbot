package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
)

type Muc struct {
	Name   string
	Rooms  []RoomInfo
	client *xmpp.Client
}

func NewMuc(name string, opt map[string]interface{}) *Muc {
	var rooms []RoomInfo
	for _, i := range opt["rooms"].([]map[string]interface{}) {
		room := RoomInfo{
			JID:      i["jid"].(string),
			Nickname: i["nickname"].(string),
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

func (m *Muc) Restart() {
	m.End()

	var rooms []RoomInfo
	v := config.Plugin[m.GetName()]
	for _, i := range v["rooms"].([]map[string]interface{}) {
		room := RoomInfo{
			JID:      i["jid"].(string),
			Nickname: i["nickname"].(string),
		}
		if i["password"] != nil {
			room.Password = i["password"].(string)
		}
		rooms = append(rooms, room)
	}
	m.Rooms = rooms

	m.Begin(m.client)
}

func (m *Muc) Chat(msg xmpp.Chat) {
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

func (m *Muc) Presence(pres xmpp.Presence) {
}

func (m *Muc) Help() string {
	return "help for muc"
}

func (m *Muc) GetOptions() map[string]string {
	return map[string]string{"name": "muc"}
}

func (m *Muc) SetOption(key, val string) {
	println("[muc] set " + key + "=" + val)
}
