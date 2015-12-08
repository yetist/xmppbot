package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"strings"
)

type Room struct {
	JID      string
	Nickname string
	Password string
	Block    []string
}

func NewRoom(jid, nickname, password string) *Room {
	return &Room{
		JID:      jid,
		Nickname: nickname,
		Password: password,
	}
}

func (r *Room) ListBlocks() string {
	return "== Block of " + r.JID + " ==\n" + strings.Join(r.Block, "\n")
}

func (r *Room) BlockOne(nick string) {
	for _, v := range r.Block {
		if nick == v {
			return
		}
	}
	r.Block = append(r.Block, nick)
}

func (r *Room) UnBlockOne(nick string) {
	r.Block = ListDelete(r.Block, nick)
}

func (r *Room) IsBlocked(nick string) bool {
	for _, v := range r.Block {
		if nick == v {
			return true
		}
	}
	return false
}

func (r *Room) SetNick(client *xmpp.Client, nick string) {
	msg := fmt.Sprintf("<presence from='%s/%s' to='%s/%s'/>",
		config.Account.Username, config.Account.Resource, r.JID, nick)
	client.SendOrg(msg)
	r.Nickname = nick
}

// 群聊消息是由bot自己发出的吗？
func RoomsMsgFromBot(rooms []*Room, msg xmpp.Chat) bool {
	if msg.Type == "groupchat" {
		for _, v := range rooms {
			if msg.Remote == v.JID+"/"+v.Nickname {
				return true
			}
		}
	}
	return false
}

// bot 在群里被点名了吗？
func RoomsMsgCallBot(rooms []*Room, msg xmpp.Chat) (ok bool, text string) {
	if msg.Type == "groupchat" {
		for _, v := range rooms {
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

// 此人在聊天中被忽略了吗?
func RoomsMsgBlocked(rooms []*Room, msg xmpp.Chat) bool {
	if msg.Type == "groupchat" {
		for _, v := range rooms {
			roomid, nick := SplitJID(msg.Remote)
			if roomid == v.JID && v.IsBlocked(nick) {
				return true
			}
		}
	}
	return false
}
