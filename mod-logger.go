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
}

func NewLogger(name string, opt map[string]interface{}) *Logger {
	return &Logger{
		Name: name,
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
	if len(msg.Text) == 0 {
		return
	}

	if msg.Type == "chat" {
		if m.Option["chat"] {
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"] {
			admin := GetAdminPlugin()
			rooms := admin.GetRooms()
			//忽略bot自己发送的消息
			if RoomsMsgFromBot(rooms, msg) || RoomsMsgBlocked(rooms, msg) {
				return
			}
			//roomid, _ := SplitJID(msg.Remote)
			//SendPub(m.client, roomid, m.RandomList[rand.Intn(len(m.RandomList))])
		}
	}
}

func (m *Logger) Presence(pres xmpp.Presence) {
}

func (m *Logger) Help() string {
	admin := GetAdminPlugin()
	msg := []string{
		"Logger模块为自动应答模块，在以下情况下触发：和Bot聊天、在群聊时提到Bot",
		"支持以下命令：",
		admin.GetCmdString("fuck") + "   无聊透顶的命令，慎用",
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
