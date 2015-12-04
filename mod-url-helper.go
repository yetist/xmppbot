package main

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
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
			m.DoHttpHelper(msg)
		}
	} else if msg.Type == "groupchat" {
		if m.Option["room"] {
			admin := GetAdminPlugin()
			rooms := admin.GetRooms()
			//忽略bot自己发送的消息
			if RoomsMsgFromBot(rooms, msg) || RoomsMsgBlocked(rooms, msg) {
				return
			}
			m.DoHttpHelper(msg)
		}
	}
}

func (m *UrlHelper) DoHttpHelper(msg xmpp.Chat) {
	if strings.Contains(msg.Text, "http://") || strings.Contains(msg.Text, "https://") {
		for k, url := range get_urls(msg.Text) {
			if url != "" {
				status, size, mime := get_url_info(url)
				if status != http.StatusOK {
					println(k, url, "访问出错了")
					return
				}
				if strings.HasPrefix(mime, "text/html") {
					println(k, url, "发了一个网页", size)
					//get_html_title(str string) (title string) {
				} else if strings.HasPrefix(mime, "image/") {
					println(k, url, "发了一个图片")
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

func get_url_info(url string) (status int, size int64, mime string) {
	response, err := http.Head(url)
	if err != nil {
		return
	}

	if status = response.StatusCode; status != http.StatusOK {
		return
	}
	length, _ := strconv.Atoi(response.Header.Get("Content-Length"))
	size = int64(length)
	mime = response.Header.Get("Content-Type")
	return
}

func get_urls(source string) []string {
	pattern := `https?://[\w\-./%?=&]+[\w\-./%?=&]*`
	reg := regexp.MustCompile(pattern)
	return reg.FindAllString(source, -1)
}

func get_url_contents(url string) (cont []byte, err error) {
	var resp *http.Response
	resp, err = http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	cont, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return []byte{}, err
	}
	return
}

// reqType is one of HTTP request strings (GET, POST, PUT, DELETE, etc.)
func DoRequest(reqType string, url string, headers map[string]string, data []byte, timeoutSeconds int) (int, []byte, map[string][]string, error) {
	var reader io.Reader
	if data != nil && len(data) > 0 {
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(reqType, url, reader)
	if err != nil {
		return 0, nil, nil, err
	}

	// I strongly advise setting user agent as some servers ignore request without it
	req.Header.Set("User-Agent", "YourUserAgentString")
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	var (
		statusCode int
		body       []byte
		timeout    time.Duration
		ctx        context.Context
		cancel     context.CancelFunc
		header     map[string][]string
	)
	timeout = time.Duration(time.Duration(timeoutSeconds) * time.Second)
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err = httpDo(ctx, req, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}

		defer resp.Body.Close()
		body, _ = ioutil.ReadAll(resp.Body)
		statusCode = resp.StatusCode
		header = resp.Header

		return nil
	})
	return statusCode, body, header, err
}
