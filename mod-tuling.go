package main

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/mattn/go-xmpp"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

type Tuling struct {
	Name   string
	URL    string
	Key    string
	client *xmpp.Client
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
	return "图灵机器人插件"
}

func (m *Tuling) CheckEnv() bool {
	return true
}

func (m *Tuling) Begin(client *xmpp.Client) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.client = client
}

func (m *Tuling) End() {
	fmt.Printf("%s End\n", m.GetName())
}

func (m *Tuling) Restart() {
}

func (m *Tuling) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}

	if msg.Type == "chat" {
		if m.Option["chat"] {
			ReplyAuto(m.client, msg, m.GetAnswer(msg.Text))
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"] {
			admin := GetAdminPlugin()
			rooms := admin.GetRooms()
			//忽略bot自己发送的消息
			if RoomsMsgFromBot(rooms, msg) || RoomsMsgBlocked(rooms, msg) {
				return
			}
			if ok, message := RoomsMsgCallBot(rooms, msg); ok {
				roomid, _ := SplitJID(msg.Remote)
				SendPub(m.client, roomid, m.GetAnswer(message))
			}
		}
	}
}

func (m *Tuling) Presence(pres xmpp.Presence) {
}

func (m *Tuling) Help() string {
	msg := []string{
		"Tuling模块为图灵机器人模块，在和Bot聊天及群聊时提到Bot时自动响应．",
	}
	return strings.Join(msg, "\n")
}

func (m *Tuling) GetOptions() map[string]string {
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

func (m *Tuling) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		m.Option[key] = StringToBool(val)
	}
}

func (m *Tuling) Request(words string) (string, error) {

	resp, err := http.Get(fmt.Sprintf("%s?key=%s&loc=%s&info=%s", m.URL, m.Key, "北京上地", words))
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	json, err := simplejson.NewJson(body)
	fmt.Printf("%#v\n", string(body))
	if err != nil {
		return "", err
	}

	result, err := json.Map()
	fmt.Printf("%#v\n", result)
	if err != nil {
		return "", err
	}

	textValue := result["text"]
	text := reflect.ValueOf(textValue).Interface().(string)

	return text, nil
}

func (m *Tuling) GetAnswer(text string) string {
	txt := strings.TrimSpace(text)

	if text, err := m.Request(txt); err != nil {
		return "我知道了"
	} else {
		return strings.Replace(text, "图灵", "", -1)
	}
}
