package main

import (
	"github.com/mattn/go-xmpp"
)

type Chat struct {
	Fuck   string
	Random string
}

func NewChat(opt map[string]interface{}) *Chat {
	return &Chat{
		Fuck:   opt["fuck"].(string),
		Random: opt["random"].(string),
	}
}

func (c *Chat) Prep(client *xmpp.Client) {
	println("call chat prep")
}

func (c *Chat) Chat() {
	println("call chat chat")
}

func (c *Chat) Presence() {
	println("call chat presence")
}
