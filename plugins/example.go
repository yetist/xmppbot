package plugins

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/robot"
	"strings"
)

type Example struct {
	Name   string
	Option map[string]bool
	bot    *robot.Bot
}

func NewExample(name string, opt map[string]interface{}) *Example {
	return &Example{
		Name:   name,
		Option: map[string]bool{},
	}
}

func (m *Example) GetName() string {
	return m.Name
}

func (m *Example) GetSummary() string {
	return "示例模块"
}

func (m *Example) Help() string {
	msg := []string{
		m.GetSummary() + ": 回显消息．",
	}
	return strings.Join(msg, "\n")
}

func (m *Example) Description() string {
	msg := []string{m.Help(),
		"当有好友或群聊消息时将自动回复原内容。",
		"本模块可配置属性:",
	}
	return strings.Join(msg, "\n")
}

func (m *Example) CheckEnv() bool {
	return true
}

func (m *Example) Start(bot *robot.Bot) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.bot = bot
}

func (m *Example) Stop() {
	fmt.Printf("[%s] Stop\n", m.GetName())
}

func (m *Example) Restart() {
}

func (m *Example) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}
	//fmt.Printf("%#v\n", msg)
	m.bot.ReplyAuto(msg, msg.Text)
}

func (m *Example) Presence(pres xmpp.Presence) {
}

func (m *Example) GetOptions() map[string]string {
	opts := map[string]string{}
	return opts
}

func (m *Example) SetOption(key, val string) {
}
