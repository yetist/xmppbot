package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/config"
	"github.com/yetist/xmppbot/core"
	"github.com/yetist/xmppbot/utils"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"
)

type Random struct {
	Name       string
	Fuck       string
	Random     string
	FuckList   []string
	RandomList []string
	bot        *core.Bot
	Option     map[string]bool
}

func NewRandom(name string, opt map[string]interface{}) *Random {
	return &Random{
		Name:   name,
		Fuck:   opt["fuck"].(string),
		Random: opt["random"].(string),
		Option: map[string]bool{
			"chat": true,
			"room": true,
		},
	}
}

func (m *Random) GetName() string {
	return m.Name
}

func (m *Random) GetSummary() string {
	return "和Bot聊天时自动回复消息"
}

func (m *Random) Help() string {
	msg := []string{
		"Random模块为自动应答模块，在以下情况下触发：和Bot聊天、在群聊时提到Bot",
		"支持以下命令：",
		m.bot.GetCmdString("fuck") + "   无聊透顶的命令，慎用",
	}
	return strings.Join(msg, "\n")
}

func (m *Random) Description() string {
	msg := []string{m.Help(),
		"本模块可配置属性:",
	}
	options := m.GetOptions()
	keys := utils.SortMapKeys(options)
	for _, v := range keys {
		msg = append(msg, fmt.Sprintf("%-20s : %s", v, options[v]))
	}
	return strings.Join(msg, "\n")
}

func (m *Random) CheckEnv() bool {
	if config.GetDataPath(AppName, AppVersion, m.Fuck) == "" {
		return false
	}
	if config.GetDataPath(AppName, AppVersion, m.Random) == "" {
		return false
	}
	return true
}

func (m *Random) Start(bot *core.Bot) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.bot = bot
	rand.Seed(time.Now().Unix())

	if data, err := ioutil.ReadFile(config.GetDataPath(AppName, AppVersion, m.Fuck)); err == nil {
		m.FuckList = strings.Split(string(data), "\n")
	}

	if data, err := ioutil.ReadFile(config.GetDataPath(AppName, AppVersion, m.Random)); err == nil {
		m.RandomList = strings.Split(string(data), "\n")
	}
}

func (m *Random) Stop() {
	fmt.Printf("[%s] Stop\n", m.GetName())
}

func (m *Random) Restart() {
}

func (m *Random) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}
	if msg.Type == "chat" {
		if m.Option["chat"] {
			if msg.Text == m.bot.GetCmdString("fuck") {
				m.bot.ReplyAuto(msg, m.FuckList[rand.Intn(len(m.FuckList))])
			} else {
				m.bot.ReplyAuto(msg, m.RandomList[rand.Intn(len(m.RandomList))])
			}
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"] {
			//忽略bot自己发送的消息
			if m.bot.SentThis(msg) || m.bot.BlockRemote(msg) {
				return
			}
			if msg.Text == m.bot.GetCmdString("fuck") {
				roomid, nick := utils.SplitJID(msg.Remote)
				m.bot.SendPub(roomid, nick+": "+m.FuckList[rand.Intn(len(m.FuckList))])
			}
			if ok, _ := m.bot.Called(msg); ok {
				roomid, _ := utils.SplitJID(msg.Remote)
				m.bot.SendPub(roomid, m.RandomList[rand.Intn(len(m.RandomList))])
			}
		}
	}
}

func (m *Random) Presence(pres xmpp.Presence) {
}

func (m *Random) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		if k == "chat" {
			opts[k] = utils.BoolToString(v) + "  #是否在好友间启用随机回复"
		} else if k == "room" {
			opts[k] = utils.BoolToString(v) + "  #是否在群聊时启用随机回复"
		}
	}
	return opts
}

func (m *Random) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		m.Option[key] = utils.StringToBool(val)
	}
}
