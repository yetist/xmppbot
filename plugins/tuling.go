package plugins

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/robot"
	"github.com/yetist/xmppbot/utils"
	"io/ioutil"
	"net/http"
	"strings"
)

type Tuling struct {
	Name   string
	URL    string
	Key    string
	bot    *robot.Bot
	Option map[string]bool
}

func NewTuling(name string, opt map[string]interface{}) *Tuling {
	return &Tuling{
		Name: name,
		URL:  "http://www.tuling123.com/openapi/api",
		Key:  opt["key"].(string),
		Option: map[string]bool{
			"chat": true,
			"room": true,
		},
	}
}

func (m *Tuling) GetName() string {
	return m.Name
}

func (m *Tuling) GetSummary() string {
	return "图灵机器人模块"
}

func (m *Tuling) Help() string {
	msg := []string{
		m.GetSummary() + ": 在和Bot聊天及群聊时提到Bot时自动响应．",
	}
	return strings.Join(msg, "\n")
}

func (m *Tuling) Description() string {
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

func (m *Tuling) CheckEnv() bool {
	return true
}

func (m *Tuling) Start(bot *robot.Bot) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.bot = bot
}

func (m *Tuling) Stop() {
	fmt.Printf("%s End\n", m.GetName())
}

func (m *Tuling) Restart() {
}

func (m *Tuling) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}

	//忽略命令消息
	if m.bot.IsCmd(msg.Text) {
		return
	}

	if msg.Type == "chat" {
		if m.Option["chat"] {
			m.bot.ReplyAuto(msg, m.GetAnswer(msg.Text, utils.GetMd5(msg.Remote)))
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"] {
			//忽略bot自己发送的消息
			if m.bot.SentThis(msg) || m.bot.BlockRemote(msg) {
				return
			}
			if ok, message := m.bot.Called(msg); ok {
				roomid, _ := utils.SplitJID(msg.Remote)
				m.bot.SendPub(roomid, m.GetAnswer(message, utils.GetMd5(msg.Remote)))
			}
		}
	}
}

func (m *Tuling) Presence(pres xmpp.Presence) {
}

func (m *Tuling) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		if k == "chat" {
			opts[k] = utils.BoolToString(v) + "  #是否响应好友消息"
		} else if k == "room" {
			opts[k] = utils.BoolToString(v) + "  #是否响应群聊呼叫消息"
		}
	}
	return opts
}

func (m *Tuling) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		m.Option[key] = utils.StringToBool(val)
	}
}

func (m *Tuling) Request(words, uid string) (text string, err error) {

	resp, err := http.Get(fmt.Sprintf("%s?key=%s&userid=%s&loc=%s&info=%s", m.URL, m.Key, uid, "北京上地", words))
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var json *simplejson.Json
	json, err = simplejson.NewJson(body)
	if err != nil {
		return "", err
	}

	code := json.Get("code").MustInt()
	if code == 100000 {
		text = json.Get("text").MustString()
	} else if code == 200000 {
		text = fmt.Sprintf("<p>%s, 点击查看<a href='%s'>详情</a><br/></p>", json.Get("text").MustString(), json.Get("url").MustString())
	} else if code == 302000 {
		var l []string
		list := json.Get("list").MustArray()
		for _, v := range list {
			item := v.(map[string]interface{})
			l = append(l, fmt.Sprintf("%s:<a href='%s'>%s</a><br/>", item["source"].(string), item["detailurl"].(string), item["article"].(string)))
		}
		text = fmt.Sprintf("<p>%s<br/>%s<br/></p>", json.Get("text").MustString(), strings.Join(l, "\n"))
	} else if code == 308000 {
		fmt.Printf("get menu info:%s\n", json.Get("text").MustString())
		var l []string
		list := json.Get("list").MustArray()
		for _, v := range list {
			item := v.(map[string]interface{})
			l = append(l, fmt.Sprintf("<a href='%s'>%s</a>，食材:%s<br/>", item["detailurl"].(string), item["name"].(string), item["info"].(string)))
		}
		text = fmt.Sprintf("<p>%s<br/>%s<br/></p>", json.Get("text").MustString(), strings.Join(l, "\n"))
	}
	return text, nil
}

func (m *Tuling) GetAnswer(text, uid string) string {
	txt := strings.TrimSpace(text)

	if text, err := m.Request(txt, uid); err != nil {
		return "我知道了"
	} else {
		return strings.Replace(text, "图灵", "", -1)
	}
}
