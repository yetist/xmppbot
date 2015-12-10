package main

import (
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

func (r *Room) GetJID() string {
	return r.JID
}

func (r *Room) SetNick(nick string) {
	r.Nickname = nick
}
