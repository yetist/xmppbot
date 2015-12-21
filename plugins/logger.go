package plugins

import (
	"fmt"
	"github.com/go-xorm/xorm"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/robot"
	"github.com/yetist/xmppbot/utils"
	"net/http"
	"strings"
	"text/template"
	"time"
)

const (
	index_tmpl = `
{{range .}}
{{if .IsRoom}}<p>chatroom: <a href='{{.JID}}/'>{{.JID}}</a></p>{{end}}
{{end}}`
	jid_tmpl = `<a href="../">Chatroom Index</a><br/>
{{range .}}
<p>{{.Created.Format "2006-01-02"}}: <a href='{{.Created.Format "2006-01-02"}}.txt'>Text</a> <a href='{{.Created.Format "2006-01-02"}}.html'>Html</a></p>
{{end}}`
	show_text_tmpl = `{{range .}}
{{if and .IsRoom .IsImage}}[{{.Created.Format "2006-01-02 15:04:05"}}] {{.Nick|printf "%-10s"}}: ***image***{{else}}[{{.Created.Format "2006-01-02 15:04:05"}}] {{.Nick|printf "%-10s"}}: {{.Text}}{{end}}
{{end}}`
	show_html_tmpl = `<html><body><a href="../">Logs date</a><br/>
{{range .}}
{{if .IsRoom}}[{{.Created.Format "2006-01-02 15:04:05"}}] {{.Nick}}: {{.Text}}<br/>{{end}}
{{end}}</body></html>`
)

type Logger struct {
	Name   string
	Option map[string]bool
	bot    *robot.Bot
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
	return "日志记录模块"
}

func (m *Logger) Help() string {
	msg := []string{
		m.GetSummary() + ": 当有好友或群聊消息时将自动记录日志．",
	}
	return strings.Join(msg, "\n")
}

func (m *Logger) Description() string {
	msg := []string{m.Help(),
		"当有好友或群聊消息时将自动记录日志．对好友消息，将只记录好友发出的消息，不记录bot回应的消息，对群聊消息将全部记录。",
		"在本模块启用时，将同时提供一个web服务来查询所有历史聊天记录。",
		"历史记录的网址为 http://your-host-name/" + m.GetName() + "/",
		"本模块可配置属性:",
	}
	options := m.GetOptions()
	keys := utils.SortMapKeys(options)
	for _, v := range keys {
		msg = append(msg, fmt.Sprintf("%-20s : %s", v, options[v]))
	}
	return strings.Join(msg, "\n")
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
	jid, nick := utils.SplitJID(msg.Remote)
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
	err := m.x.Where("j_i_d = ?", jid).Limit(1).Desc("created").Find(&logs)
	return logs, err
}

func (m *Logger) CheckEnv() bool {
	if m.x == nil {
		fmt.Printf("[%s] Database initial error, disable this plugin.\n", m.GetName())
		return false
	}
	m.x.ShowDebug = false
	m.x.ShowErr = false
	m.x.ShowSQL = false
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
	t, _ := template.New("index").Parse(index_tmpl)
	t.Execute(w, logs)
}

func (m *Logger) JIDPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jid := vars["jid"]
	logs := make([]ChatLogger, 0)

	t, _ := template.New("jid").Parse(jid_tmpl)
	m.x.Select("distinct strftime('%Y-%m-%d', created, 'localtime') as created").Where("j_i_d = ?", jid).Find(&logs)
	t.Execute(w, logs)
}

func (m *Logger) ShowPage(w http.ResponseWriter, r *http.Request) {
	logs := make([]ChatLogger, 0)
	vars := mux.Vars(r)
	jid := vars["jid"]
	date := vars["date"]
	format := vars["format"]
	sql := fmt.Sprintf("select * from chat_logger where (j_i_d = '%s') and (strftime('%%Y-%%m-%%d', created)=strftime('%%Y-%%m-%%d', '%s'))", jid, date)
	if err := m.x.Sql(sql).Desc("created").Find(&logs); err != nil {
		w.Write([]byte("no record"))
	} else {
		var t *template.Template
		if format == "txt" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			t, _ = template.New("index").Parse(show_text_tmpl)
		} else if format == "html" {
			w.Header().Set("Content-Type", "text/html")
			t, _ = template.New("index").Parse(show_html_tmpl)
		}
		t.Execute(w, logs)
	}
}

func (m *Logger) Start(bot *robot.Bot) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.bot = bot
	m.bot.AddHandler(m.GetName(), "/", m.IndexPage, "index")
	m.bot.AddHandler(m.GetName(), "/{jid}/", m.JIDPage, "jidpage")
	m.bot.AddHandler(m.GetName(), "/{jid}/{date}.{format}", m.ShowPage, "showlog")
}

func (m *Logger) Stop() {
	fmt.Printf("[%s] Stop\n", m.GetName())
	m.bot.DelHandler(m.GetName(), "index")
	m.bot.DelHandler(m.GetName(), "jidpage")
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

func (m *Logger) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		if k == "chat" {
			opts[k] = utils.BoolToString(v) + "  #是否响应好友消息"
		} else if k == "room" {
			opts[k] = utils.BoolToString(v) + "  #是否响应群聊消息"
		}
	}
	return opts
}

func (m *Logger) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		m.Option[key] = utils.StringToBool(val)
	}
}
