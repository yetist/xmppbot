package main

import (
	"fmt"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mattn/go-xmpp"
	"net/http"
	"strings"
	"time"
)

type Logger struct {
	Name   string
	Option map[string]bool
	bot    *Bot
	x      *xorm.Engine
}

func NewLogger(name string, opt map[string]interface{}) *Logger {
	var err error

	m := &Logger{
		Name: name,
		Option: map[string]bool{
			"chat": opt["chat"].(bool),
			"room": opt["room"].(bool),
		},
	}

	switch opt["dbtype"].(string) {
	case "sqlite3":
		m.x, err = xorm.NewEngine("sqlite3", opt["dbname"].(string))
	case "mysql":
		m.x, err = xorm.NewEngine("mysql", opt["dbuser"].(string)+" "+opt["dbpass"].(string)+" "+opt["dbname"].(string))
	}
	if err != nil {
		fmt.Printf("[%s] Database initial error: %v\n", name, err)
	}
	return m
}

func (m *Logger) GetName() string {
	return m.Name
}

func (m *Logger) GetSummary() string {
	return "记录聊天日志"
}

type ChatLogger struct {
	Id      int64
	JID     string
	Nick    string
	Text    string
	IsRoom  bool
	IsImage bool
	Created time.Time `xorm:"created index"`
	Updated time.Time `xorm:"updated index"`
}

func (m *Logger) LogInsert(msg xmpp.Chat) (err error) {
	jid, nick := SplitJID(msg.Remote)
	log := &ChatLogger{JID: jid, Nick: nick, Text: msg.Text}

	if strings.Contains(msg.Text, "<img") {
		log.IsImage = true
	}
	if msg.Type == "groupchat" {
		log.IsRoom = true
	}
	_, err = m.x.InsertOne(log)
	return
}

func (m *Logger) SelectAllText() ([]ChatLogger, error) {
	logs := make([]ChatLogger, 0)
	err := m.x.Find(&logs)
	return logs, err
}

func (m *Logger) GetRoomLogs(jid string) ([]ChatLogger, error) {
	logs := make([]ChatLogger, 0)
	//err := x.Where("j_i_d = ?", jid).Where("direct = ?", 1).Limit(n).Desc("created").Find(&nets)
	err := m.x.Where("j_i_d = ?", jid).Limit(1).Desc("created").Find(&logs)
	return logs, err
}

func (m *Logger) CheckEnv() bool {
	if m.x == nil {
		fmt.Printf("[%s] Database initial error, disable this plugin.\n", m.GetName())
		return false
	}
	//m.x.ShowDebug = true
	//m.x.ShowErr = true
	//m.x.ShowSQL = true
	m.x.SetMaxConns(10)

	cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 10000)
	m.x.SetDefaultCacher(cacher)
	m.x.Sync2(new(ChatLogger))

	return true
}

func (m *Logger) Start(bot *Bot) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.bot = bot
	r := bot.GetRouter(m.GetName())
	r.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from"))
	})
}

func (m *Logger) Stop() {
	fmt.Printf("[%s] Stop\n", m.GetName())
}

func (m *Logger) Restart() {
}

func (m *Logger) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}

	if msg.Type == "chat" {
		if m.Option["chat"] {
			m.LogInsert(msg)
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"] {
			m.LogInsert(msg)
		}
	}
}

func (m *Logger) Presence(pres xmpp.Presence) {
}

func (m *Logger) Help() string {
	msg := []string{
		"日志记录模块，当有好友或群聊消息时将自动记录日志．",
	}
	return strings.Join(msg, "\n")
}

func (m *Logger) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		if k == "chat" {
			opts[k] = BoolToString(v) + "  #是否响应好友消息"
		} else if k == "room" {
			opts[k] = BoolToString(v) + "  #是否响应群聊消息"
		}
	}
	return opts
}

func (m *Logger) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		m.Option[key] = StringToBool(val)
	}
}
