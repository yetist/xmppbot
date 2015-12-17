package plugins

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/robot"
	"github.com/yetist/xmppbot/utils"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Url struct {
	Name   string
	bot    *robot.Bot
	Option map[string]interface{}
}

func NewUrl(name string, opt map[string]interface{}) *Url {
	return &Url{
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

func (m *Url) GetName() string {
	return m.Name
}

func (m *Url) GetSummary() string {
	return "URL链接辅助模块"
}

func (m *Url) Help() string {
	msg := []string{
		m.GetSummary() + ": 可以自动为聊天消息提供url标题或显示图片缩略图。",
	}
	return strings.Join(msg, "\n")
}

func (m *Url) Description() string {
	msg := []string{m.Help(),
		"当用户在聊天过程中输入url时，机器人将自动去打开此url，并显示网页标题(html类型)或者显示一个缩略图(image类型).",
		"本模块可配置属性:",
	}
	options := m.GetOptions()
	keys := utils.SortMapKeys(options)
	for _, v := range keys {
		msg = append(msg, fmt.Sprintf("%-20s : %s", v, options[v]))
	}
	return strings.Join(msg, "\n")
}

func (m *Url) CheckEnv() bool {
	return true
}

func (m *Url) Start(bot *robot.Bot) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.bot = bot
}

func (m *Url) Stop() {
	fmt.Printf("[%s] Stop\n", m.GetName())
}

func (m *Url) Restart() {
	opt := m.bot.GetPluginOption(m.GetName())
	m.Option["chat"] = opt["chat"].(bool)
	m.Option["room"] = opt["room"].(bool)
	m.Option["timeout"] = opt["timeout"].(int64)
	m.Option["width"] = 100
	m.Option["height"] = 100
}

func (m *Url) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}

	if msg.Type == "chat" {
		if m.Option["chat"].(bool) {
			if m.bot.SentThis(msg) {
				return
			}
			text := m.GetHelper(msg.Text)
			if text != "" {
				m.bot.ReplyAuto(msg, fmt.Sprintf("<p>%s</p>", text))
			}
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"].(bool) {
			//忽略bot自己发送的消息
			if m.bot.SentThis(msg) || m.bot.BlockRemote(msg) {
				return
			}
			text := m.GetHelper(msg.Text)
			if text != "" {
				roomid, nick := utils.SplitJID(msg.Remote)
				m.bot.SendPub(roomid, fmt.Sprintf("<p>%s %s</p>", nick, text))
			}
		}
	}
}

func (m *Url) GetHelper(text string) string {
	if strings.Contains(text, "http://") || strings.Contains(text, "https://") {
		for k, url := range GetUrls(text) {
			if url != "" {
				timeout := time.Duration(m.Option["timeout"].(int64))
				res, body, err := utils.HttpOpen(url, timeout, "")
				if err != nil || res.StatusCode != http.StatusOK {
					return fmt.Sprintf("对不起，无法打开此<a href='%s'>链接</a>", url)
				}
				if strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
					title := utils.GetUTF8HtmlTitle(string(body))
					if title == "" {
						return fmt.Sprintf("报歉，无法得到<a href='%s'>链接</a>标题", url)
					} else {
						return fmt.Sprintf("发链接了，标题是[<a href='%s'>%s</a>]", url, title)
					}
				} else if strings.HasPrefix(res.Header.Get("Content-Type"), "image/") {
					img := utils.GetBase64Image(body, m.Option["width"].(int), m.Option["height"].(int))
					return fmt.Sprintf("发<a href='%s'>图片</a>了:<br/><img alt='点击看大图' src='%s'/>", url, img)
				} else {
					println(k, url, "发了其它类型文件")
				}
			}
		}
	}
	return ""
}

func (m *Url) Presence(pres xmpp.Presence) {
}

func (m *Url) GetOptions() map[string]string {
	opts := map[string]string{}
	for k, v := range m.Option {
		if k == "chat" {
			opts[k] = utils.BoolToString(v.(bool)) + "  #是否响应好友消息"
		} else if k == "room" {
			opts[k] = utils.BoolToString(v.(bool)) + "  #是否响应群聊消息"
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

func (m *Url) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		if key == "chat" || key == "room" {
			m.Option[key] = utils.StringToBool(val)
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
