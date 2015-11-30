package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"io/ioutil"
	"math/rand"
	"strings"
)

type AutoReply struct {
	Name       string
	Fuck       string
	Random     string
	FuckList   []string
	RandomList []string
	client     *xmpp.Client
	Option     map[string]bool
}

func NewAutoReply(name string, opt map[string]interface{}) *AutoReply {
	return &AutoReply{
		Name:   name,
		Fuck:   opt["fuck"].(string),
		Random: opt["random"].(string),
		Option: map[string]bool{
			"chat": true,
			"room": true,
		},
	}
}

func (m *AutoReply) GetName() string {
	return m.Name
}

func (m *AutoReply) GetSummary() string {
	return "和Bot聊天时自动回复消息"
}

func (m *AutoReply) CheckEnv() bool {
	if GetDataPath(m.Fuck) == "" {
		return false
	}
	if GetDataPath(m.Random) == "" {
		return false
	}
	return true
}

func (m *AutoReply) Begin(client *xmpp.Client) {
	m.client = client
	rand.Seed(42)

	if data, err := ioutil.ReadFile(GetDataPath(m.Fuck)); err == nil {
		m.FuckList = strings.Split(string(data), "\n")
	}

	if data, err := ioutil.ReadFile(GetDataPath(m.Random)); err == nil {
		m.RandomList = strings.Split(string(data), "\n")
	}
}

func (m *AutoReply) End() {
	fmt.Printf("%s End\n", m.GetName())
}

func (m *AutoReply) Restart() {
}

func (m *AutoReply) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 {
		return
	}
	if msg.Type == "chat" {
		if m.Option["chat"] {
			ReplyAuto(m.client, msg, m.RandomList[rand.Intn(len(m.RandomList))])
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"] {
			admin := GetAdminPlugin()
			//忽略bot自己发送的消息
			if admin.IsBotSend(msg) {
				return
			}
			if !admin.IsNotifyBot(msg) {
				return
			}
			roomid, _ := SplitJID(msg.Remote)
			SendPub(m.client, roomid, m.RandomList[rand.Intn(len(m.RandomList))])
		}
	}
}

func (m *AutoReply) Presence(pres xmpp.Presence) {
}

func (m *AutoReply) Help() {
}

func (m *AutoReply) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		if v {
			opts[k] = "true"
		} else {
			opts[k] = "false"
		}
		if k == "chat" {
			opts[k] = opts[k] + "  (是否在好友间启用随机回复)"
		} else if k == "room" {
			opts[k] = opts[k] + "  (是否在群聊时启用随机回复)"
		}
	}
	return opts
}

func (m *AutoReply) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		m.Option[key] = StringToBool(val)
	}
}
