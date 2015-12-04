package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"golang.org/x/net/html"
	"strings"
)

type UrlHelper struct {
	Name   string
	client *xmpp.Client
	Option map[string]bool
}

func NewUrlHelper(name string, opt map[string]interface{}) *UrlHelper {
	return &UrlHelper{
		Name: name,
		Option: map[string]bool{
			"chat": opt["chat"].(bool),
			"room": opt["room"].(bool),
		},
	}
}

func (m *UrlHelper) GetName() string {
	return m.Name
}

func (m *UrlHelper) GetSummary() string {
	return "URL链接辅助功能"
}

func (m *UrlHelper) CheckEnv() bool {
	return true
}

func (m *UrlHelper) Begin(client *xmpp.Client) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.client = client
}

func (m *UrlHelper) End() {
	fmt.Printf("%s End\n", m.GetName())
}

func (m *UrlHelper) Restart() {
}

func (m *UrlHelper) Chat(msg xmpp.Chat) {
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

		}
	}
}

func (m *UrlHelper) Presence(pres xmpp.Presence) {
}

func (m *UrlHelper) Help() string {
	msg := []string{
		"UrlHelper模块为自动响应模块，当聊天内容中包含url时将自动激活．",
	}
	return strings.Join(msg, "\n")
}

func (m *UrlHelper) GetOptions() map[string]string {
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

func (m *UrlHelper) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		m.Option[key] = StringToBool(val)
	}
}

func get_title_content(n *html.Node) (title string) {
	if n.Type == html.ElementNode && n.Data == "title" {
		t := n.FirstChild
		if t.Type == html.TextNode {
			title = t.Data
		}
	}
	return
}

func get_html_title(str string) (title string) {
	doc, err := html.Parse(strings.NewReader(str))
	if err != nil {
		return
	}
	for a := doc.FirstChild; a != nil; a = a.NextSibling {
		if title = get_title_content(a); title != "" {
			return
		}
		for b := a.FirstChild; b != nil; b = b.NextSibling {
			if title = get_title_content(b); title != "" {
				return
			}
			for c := b.FirstChild; c != nil; c = c.NextSibling {
				if title = get_title_content(c); title != "" {
					return
				}
				for d := c.FirstChild; d != nil; d = d.NextSibling {
					if title = get_title_content(d); title != "" {
						return
					}
				}
			}
		}
	}
	return
}
