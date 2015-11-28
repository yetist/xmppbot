package main

import (
	"github.com/mattn/go-xmpp"
)

type Botter interface {
	Prep(client *xmpp.Client)
	Chat()
	Presence()
}

var plugins []Botter

func InitPlugins() {
	for m, v := range config.Plugin {
		if m == "chat" && v["enable"].(bool) {
			chat := NewChat(v)
			plugins = append(plugins, chat)
		} else if m == "muc" && v["enable"].(bool) {
			muc := NewMuc(v)
			plugins = append(plugins, muc)
		}
	}
}

func PrepPlugins(client *xmpp.Client) {
	for _, v := range plugins {
		v.Prep(client)
	}
}
