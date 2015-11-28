package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
)

type RoomOption struct {
	Jid      string
	NickName string
	RoomLog  bool
}

type Muc struct {
	Rooms []RoomOption
}

func NewMuc(opt map[string]interface{}) *Muc {
	var rooms []RoomOption
	for _, i := range opt["rooms"].([]map[string]interface{}) {
		room := RoomOption{
			Jid:      i["jid"].(string),
			NickName: i["nickname"].(string),
			RoomLog:  i["room_log"].(bool),
		}
		rooms = append(rooms, room)
	}
	return &Muc{Rooms: rooms}
}

func (m *Muc) Prep(client *xmpp.Client) {
	fmt.Print("call muc prep")
	for _, room := range m.Rooms {
		client.JoinMUC(room.Jid, room.NickName)
	}
}

func (m *Muc) Chat() {
	println("call muc muc")
}
func (m *Muc) Presence() {
	println("call muc presence")
}
