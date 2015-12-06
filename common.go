package main

import (
	"errors"
	"fmt"
	"github.com/mattn/go-xmpp"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

func SplitJID(jid string) (string, string) {
	if strings.Contains(jid, "/") {
		tokens := strings.SplitN(jid, "/", 2)
		return tokens[0], tokens[1]
	} else {
		return jid, ""
	}
}

func IsAdmin(jid string) bool {
	u, _ := SplitJID(jid)
	for _, admin := range config.Setup.Admin {
		if u == admin {
			return true
		}
	}
	return false
}

func IsValidStatus(status string) bool {
	switch status {
	case
		"away",
		"chat",
		"dnd",
		"xa":
		return true
	}
	return false
}

func StringToBool(val string) bool {
	switch strings.ToLower(val) {
	case
		"1",
		"true",
		"t",
		"y",
		"yes",
		"ok":
		return true
	}
	return false
}

func BoolToString(val bool) string {
	if val {
		return "true"
	} else {
		return "false"
	}
}

// 设置状态消息
func SetStatus(client *xmpp.Client, status, info string) {
	client.SendOrg(fmt.Sprintf("<presence xml:lang='en'><show>%s</show><status>%s</status></presence>", status, info))
}

func SendHtml(client *xmpp.Client, chat xmpp.Chat) {
	client.SendOrg(fmt.Sprintf("<message to='%s' type='%s' xml:lang='en'>"+
		"<body>%s</body>"+
		"<html xmlns='http://jabber.org/protocol/xhtml-im'><body xmlns='http://www.w3.org/1999/xhtml'>%s</body></html></message>",
		html.EscapeString(chat.Remote), html.EscapeString(chat.Type), html.EscapeString("hello"), chat.Text))
}

func MapDelete(dict map[string]interface{}, key string) {
	_, ok := dict[key]
	if ok {
		delete(dict, key)
	}
}

func SortMapKeys(dict map[string]string) []string {
	keys := make([]string, 0, len(dict))
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func ListDelete(list []string, key string) []string {
	for k, v := range list {
		if v == key {
			list = append(list[:k], list[k+1:]...)
		}
	}
	return list
}

// 回复好友消息，或聊天室私聊消息
func ReplyAuto(client *xmpp.Client, recv xmpp.Chat, text string) {
	client.Send(xmpp.Chat{Remote: recv.Remote, Type: "chat", Text: text})
}

// 回复好友消息，或聊天室公共消息
func ReplyPub(client *xmpp.Client, recv xmpp.Chat, text string) {
	if recv.Type == "groupchat" {
		roomid, _ := SplitJID(recv.Remote)
		client.Send(xmpp.Chat{Remote: roomid, Type: recv.Type, Text: text})
	} else {
		ReplyAuto(client, recv, text)
	}
}

// 发送到好友消息，或聊天室私聊消息
func SendAuto(client *xmpp.Client, to, text string) {
	client.Send(xmpp.Chat{Remote: to, Type: "chat", Text: text})
}

// 发送聊天室公共消息
func SendPub(client *xmpp.Client, to, text string) {
	client.Send(xmpp.Chat{Remote: to, Type: "groupchat", Text: text})
}

func ChatMsgFromBot(msg xmpp.Chat) bool {
	if msg.Type == "chat" {
		if id, res := SplitJID(msg.Remote); id == config.Account.Username && res == config.Account.Resource {
			return true
		}
	}
	return false
}

func HttpOpen(url string, n time.Duration) (res *http.Response, body []byte, err error) {
	timeout := n * time.Second
	failTime := timeout * 2
	c := &http.Client{
		Timeout: timeout,
	}

	if res, err = c.Get(url); err != nil {
		return
	}

	if res.StatusCode != http.StatusOK {
		return
	}

	errc := make(chan error, 1)
	go func() {
		body, err = ioutil.ReadAll(res.Body)
		errc <- err
		res.Body.Close()
	}()

	select {
	case err = <-errc:
		if err != nil {
			return
		}
	case <-time.After(failTime):
		err = errors.New("open timeout")
		return
	}
	return
}

func getNodeData(n *html.Node, node string) (data string) {
	if n.Type == html.ElementNode && n.Data == node {
		t := n.FirstChild
		if t.Type == html.TextNode {
			data = t.Data
		}
	}
	return
}

func getHtmlTitle(str string) (title string) {
	doc, err := html.Parse(strings.NewReader(str))
	if err != nil {
		return
	}
	for a := doc.FirstChild; a != nil; a = a.NextSibling {
		if title = getNodeData(a, "title"); title != "" {
			return
		}
		for b := a.FirstChild; b != nil; b = b.NextSibling {
			if title = getNodeData(b, "title"); title != "" {
				return
			}
			for c := b.FirstChild; c != nil; c = c.NextSibling {
				if title = getNodeData(c, "title"); title != "" {
					return
				}
				for d := c.FirstChild; d != nil; d = d.NextSibling {
					if title = getNodeData(d, "title"); title != "" {
						return
					}
				}
			}
		}
	}
	return
}

func getUTF8HtmlTitle(str string) string {
	var e encoding.Encoding
	var name string

	e, name, _ = charset.DetermineEncoding([]byte(str), "text/html")
	if name == "windows-1252" {
		e, name, _ = charset.DetermineEncoding([]byte(str), "text/html;charset=gbk")
	}
	r := transform.NewReader(strings.NewReader(str), e.NewDecoder())
	if b, err := ioutil.ReadAll(r); err != nil {
		println(err)
		return ""
	} else {
		return getHtmlTitle(string(b))
	}
	return ""
}
