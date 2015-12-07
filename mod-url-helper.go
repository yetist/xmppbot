package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type UrlHelper struct {
	Name   string
	client *xmpp.Client
	Option map[string]interface{}
}

func NewUrlHelper(name string, opt map[string]interface{}) *UrlHelper {
	return &UrlHelper{
		Name: name,
		Option: map[string]interface{}{
			"chat":    opt["chat"].(bool),
			"room":    opt["room"].(bool),
			"timeout": opt["timeout"].(int64),
			"width":   100,
			"height":  100,
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
	opt := config.Plugin[m.GetName()]
	m.Option["chat"] = opt["chat"].(bool)
	m.Option["room"] = opt["room"].(bool)
	m.Option["timeout"] = opt["timeout"].(int64)
	m.Option["width"] = 100
	m.Option["height"] = 100
}

func (m *UrlHelper) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}

	if msg.Type == "chat" {
		if m.Option["chat"].(bool) {
			if ChatMsgFromBot(msg) {
				fmt.Printf("[%s] *** 忽略由bot发送的消息: %s\n", m.GetName(), msg.Text)
				return
			}
			m.DoHttpHelper(msg)
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"].(bool) {
			admin := GetAdminPlugin()
			rooms := admin.GetRooms()
			//忽略bot自己发送的消息
			if RoomsMsgFromBot(rooms, msg) || RoomsMsgBlocked(rooms, msg) {
				fmt.Printf("[%s] *** 忽略由bot发送的消息: %s\n", m.GetName(), msg.Text)
				return
			}
			m.DoHttpHelper(msg)
		}
	}
}

func (m *UrlHelper) SendHtml(msg xmpp.Chat, info string) {
	if msg.Type == "groupchat" {
		roomid, nick := SplitJID(msg.Remote)
		text := fmt.Sprintf("<p>%s %s</p>", nick, info)
		SendHtml(m.client, xmpp.Chat{Remote: roomid, Type: "groupchat", Text: text})
	} else {
		text := fmt.Sprintf("<p>%s</p>", info)
		SendHtml(m.client, xmpp.Chat{Remote: msg.Remote, Type: "chat", Text: text})
	}
}

func (m *UrlHelper) DoHttpHelper(msg xmpp.Chat) {
	if strings.Contains(msg.Text, "http://") || strings.Contains(msg.Text, "https://") {
		for k, url := range GetUrls(msg.Text) {
			if url != "" {
				timeout := time.Duration(m.Option["timeout"].(int64))
				res, body, err := HttpOpen(url, timeout)
				if err != nil || res.StatusCode != http.StatusOK {
					m.SendHtml(msg, fmt.Sprintf("对不起，无法打开此<a href='%s'>链接</a>", url))
					return
				}
				if strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
					title := getUTF8HtmlTitle(string(body))
					if title == "" {
						m.SendHtml(msg, fmt.Sprintf("报歉，无法得到<a href='%s'>链接</a>标题", url))

					} else {
						m.SendHtml(msg, fmt.Sprintf("发链接了，标题是[<a href='%s'>%s</a>]", url, title))
					}
				} else if strings.HasPrefix(res.Header.Get("Content-Type"), "image/") {
					img := getBase64Image(body, m.Option["width"].(int), m.Option["height"].(int))
					m.SendHtml(msg, fmt.Sprintf("发<a href='%s'>图片</a>了:<br/><img alt='点击看大图' src='%s'/>", url, img))
				} else {
					println(k, url, "发了其它类型文件")
				}
			}
		}
	}
}

func (m *UrlHelper) Presence(pres xmpp.Presence) {
}

func (m *UrlHelper) Help() string {
	msg := []string{
		"UrlHelper模块自动为聊天消息提供url标题或显示图片缩略图。",
	}
	return strings.Join(msg, "\n")
}

func (m *UrlHelper) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		if k == "chat" {
			opts[k] = BoolToString(v.(bool)) + "  #是否记录朋友发送的日志"
		} else if k == "room" {
			opts[k] = BoolToString(v.(bool)) + "  #是否记录群聊日志"
		} else if k == "timeout" {
			opts[k] = strconv.FormatInt(v.(int64), 10) + "  #访问链接超时时间"
		} else if k == "width" {
			opts[k] = strconv.Itoa(v.(int)) + "  #预览图片宽度"
		} else if k == "height" {
			opts[k] = strconv.Itoa(v.(int)) + "  #预览图片高度"
		}
	}
	return opts

}

func (m *UrlHelper) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		if key == "chat" || key == "room" {
			m.Option[key] = StringToBool(val)
		} else if key == "timeout" {
			if i, e := strconv.ParseInt(val, 10, 0); e == nil {
				m.Option[key] = i
			}
		} else if key == "width" {
			if i, e := strconv.Atoi(val); e == nil {
				m.Option[key] = i
			}
		} else if key == "height" {
			if i, e := strconv.Atoi(val); e == nil {
				m.Option[key] = i
			}
		}
	}
}

func GetUrls(source string) []string {
	pattern := `https?://[\w\-./%?=&]+[\w\-./%?=&]*`
	reg := regexp.MustCompile(pattern)
	return reg.FindAllString(source, -1)
}
