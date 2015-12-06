package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"strings"
)

type Logger struct {
	Name   string
	client *xmpp.Client
	Option map[string]bool
	DBPath string
}

func NewLogger(name string, opt map[string]interface{}) *Logger {
	return &Logger{
		Name:   name,
		DBPath: opt["logger"].(string),
		Option: map[string]bool{
			"chat": opt["chat"].(bool),
			"room": opt["room"].(bool),
		},
	}
}

func (m *Logger) GetName() string {
	return m.Name
}

func (m *Logger) GetSummary() string {
	return "聊天日志记录"
}

func (m *Logger) CheckEnv() bool {
	return true
}

func (m *Logger) Begin(client *xmpp.Client) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.client = client
}

func (m *Logger) End() {
	fmt.Printf("%s End\n", m.GetName())
}

func (m *Logger) Restart() {
}

func (m *Logger) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}

	if msg.Type == "chat" {
		if m.Option["chat"] {
			fmt.Printf("(%s), %s, %s, %s\n", m.Name, msg.Remote, msg.Text, msg.Stamp.Local())
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"] {
			fmt.Printf("(%s), %s, %s, %s\n", m.Name, msg.Remote, msg.Text, msg.Stamp.Local())
		}
	}
}

func (m *Logger) Presence(pres xmpp.Presence) {
}

func (m *Logger) Help() string {
	//admin := GetAdminPlugin()
	msg := []string{
		"Logger模块为自动响应模块，当有好友或群聊消息时将自动记录",
	}
	return strings.Join(msg, "\n")
}

func (m *Logger) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		if k == "chat" {
			opts[k] = BoolToString(v) + "  #是否记录朋友发送的日志"
		} else if k == "room" {
			opts[k] = BoolToString(v) + "  #是否记录群聊日志"
		}
	}
	return opts
}

func (m *Logger) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		m.Option[key] = StringToBool(val)
	}
}
