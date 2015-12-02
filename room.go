package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"strings"
)

type Room struct {
	JID           string
	Nickname      string
	Password      string
	Block         []string
	Status        string
	StatusMessage string
}

func NewRoom(jid, nickname, password, status, status_message string) *Room {
	return &Room{
		JID:           jid,
		Nickname:      nickname,
		Password:      password,
		Status:        status,
		StatusMessage: status_message,
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

func (r *Room) SetNick(client *xmpp.Client, nick string) {
	msg := fmt.Sprintf("<presence from='%s/%s' to='%s/%s'/>",
		config.Account.Username, config.Account.Resource, r.JID, nick)
	client.SendOrg(msg)
	r.Nickname = nick
}

func (r *Room) SetStatus(client *xmpp.Client, status, info string) {
	msg := fmt.Sprintf("<presence from='%s/%s' to='%s/%s'>\n"+
		"<show>%s</show>\n"+
		"<status>%s</status>\n"+
		"</presence>`\n",
		config.Account.Username, config.Account.Resource, r.JID, r.Nickname, status, info)
	client.SendOrg(msg)
	r.Status = status
	r.StatusMessage = info
}
