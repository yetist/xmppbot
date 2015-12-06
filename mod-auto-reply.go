package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
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
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.client = client
	rand.Seed(time.Now().Unix())

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
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}
	admin := GetAdminPlugin()

	if msg.Type == "chat" {
		if m.Option["chat"] {
			if msg.Text == admin.GetCmdString("fuck") {
				ReplyAuto(m.client, msg, m.FuckList[rand.Intn(len(m.FuckList))])
			} else {
				ReplyAuto(m.client, msg, m.RandomList[rand.Intn(len(m.RandomList))])
			}
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"] {
			admin := GetAdminPlugin()
			rooms := admin.GetRooms()
			//忽略bot自己发送的消息
			if RoomsMsgFromBot(rooms, msg) || RoomsMsgBlocked(rooms, msg) {
				return
			}
			if msg.Text == admin.GetCmdString("fuck") {
				roomid, nick := SplitJID(msg.Remote)
				SendPub(m.client, roomid, nick+": "+m.FuckList[rand.Intn(len(m.FuckList))])
			}
			if !RoomsMsgCallBot(rooms, msg) {
				return
			}
			roomid, _ := SplitJID(msg.Remote)
			SendPub(m.client, roomid, m.RandomList[rand.Intn(len(m.RandomList))])
		}
	}
}

func (m *AutoReply) Presence(pres xmpp.Presence) {
}

func (m *AutoReply) Help() string {
	admin := GetAdminPlugin()
	msg := []string{
		"AutoReply模块为自动应答模块，在以下情况下触发：和Bot聊天、在群聊时提到Bot",
		"支持以下命令：",
		admin.GetCmdString("fuck") + "   无聊透顶的命令，慎用",
	}
	return strings.Join(msg, "\n")
}

func (m *AutoReply) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		if k == "chat" {
			opts[k] = BoolToString(v) + "  #是否在好友间启用随机回复"
		} else if k == "room" {
			opts[k] = BoolToString(v) + "  #是否在群聊时启用随机回复"
		}
	}
	return opts
}

func (m *AutoReply) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		m.Option[key] = StringToBool(val)
	}
}
