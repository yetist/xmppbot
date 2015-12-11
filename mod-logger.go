package main

import (
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/gorilla/mux"
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
	m.x.ShowDebug = true
	m.x.ShowErr = true
	m.x.ShowSQL = true
	m.x.SetMaxConns(10)

	cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 10000)
	m.x.SetDefaultCacher(cacher)
	m.x.Sync2(new(ChatLogger))

	return true
}

/* web pages */
func (m *Logger) IndexPage(w http.ResponseWriter, r *http.Request) {
	logs := make([]ChatLogger, 0)
	if err := m.x.Distinct("j_i_d").Cols("j_i_d", "is_room").Find(&logs); err != nil {
		w.Write([]byte("no record"))
		return
	}
	var lst []string
	for _, v := range logs {
		var info string
		if v.IsRoom {
			info = fmt.Sprintf("<p>chatroom: <a href='%s/'>%s</a></p>", v.JID, v.JID)
			lst = append(lst, info)
		} else {
			info = fmt.Sprintf("<p>chat: <a href='%s/'>%s</a></p>", v.JID, v.JID)
		}
		//		if v.JID == "desktoprd@groups.isoft-linux.org" {
		//			lst = append(lst, info)
		//		}
	}

	w.Write([]byte(strings.Join(lst, "\n")))
}

func (m *Logger) JIDPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jid := vars["jid"]
	logs := make([]ChatLogger, 0)

	//if err := m.x.Distinct("j_i_d").Cols("j_i_d", "is_room").Find(&logs); err != nil {
	query := fmt.Sprintf("select distinct strftime('%%Y-%%m-%%d', created, 'localtime') from chat_logger where j_i_d = %s;", jid)
	if result, err := m.x.Query(query); err != nil {
		fmt.Printf("%#v\n", err)
		w.Write([]byte("no record"))
		return
	} else {
		fmt.Printf("%#v\n", result)
	}
	//	as created").Find(&logs); err != nil {
	//"<a href=%s/desktoprd@groups.isoft-linux.org>日志</a>"))
	//Distinct("j_i_d").Find(&logs); err != nil {
	// select distinct strftime('%Y-%m-%d', created, 'localtime') from chat_logger;
	//if err := m.x.Select("distinct strftime('%Y-%m-%d', created, 'localtime') as created").Find(&logs); err != nil {
	//	w.Write([]byte("no record"))
	//}
}

func (m *Logger) ShowPage(w http.ResponseWriter, r *http.Request) {
	//logs := make([]ChatLogger, 0)
	//vars := mux.Vars(r)
	//jid := vars["jid"]
	//date := vars["date"]
	//format := vars["format"]
	//if jid != "desktoprd@groups.isoft-linux.org" {
	//}

	//if format == "txt" || format == "html" {
	//}

	//// select distinct strftime('%Y-%m-%d', created, 'localtime') from chat_logger;
	////select created from chat_logger where strftime('%Y-%m-%d', created, 'localtime')=strftime('%Y-%m-%d', '2015-12-08');
	//err := m.x.Where("j_i_d = ?", jid).Limit(1).Desc("created").Find(&logs)
	//logs := make([]ChatLogger, 0)
	//if err := m.x.Distinct("j_i_d").Find(&logs); err != nil {
	//	w.Write([]byte("no record"))
	//}
	//for _, v := range logs {
	//	fmt.Printf("%#v\n", v.JID)
	//}
	w.Write([]byte("index page."))
}

func (m *Logger) Start(bot *Bot) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.bot = bot
	m.bot.AddHandler(m.GetName(), "/", m.IndexPage, "index")
	m.bot.AddHandler(m.GetName(), "/{jid}/", m.JIDPage, "jidpage")
	m.bot.AddHandler(m.GetName(), "/{jid}/{date}.{format}", m.ShowPage, "showlog")
}

func (m *Logger) Stop() {
	fmt.Printf("[%s] Stop\n", m.GetName())
	m.bot.DelHandler(m.GetName(), "index")
	m.bot.DelHandler(m.GetName(), "showlog")
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
