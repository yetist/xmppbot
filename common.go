package main

import (
	"bytes"
	"code.google.com/p/graphics-go/graphics"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
	"image"
	_ "image/jpeg" //必须import，否则会出现：unknown format，其余类似
	"image/png"
	"io"
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

//func ChatMsgFromBot(msg xmpp.Chat) bool {
//	if msg.Type == "chat" {
//		if id, res := SplitJID(msg.Remote); id == config.Account.Username && res == config.Account.Resource {
//			return true
//		}
//	}
//	return false
//}

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
		return ""
	} else {
		return getHtmlTitle(string(b))
	}
	return ""
}

func getBase64Image(body []byte, width, height int) string {
	src, _, err := image.Decode(strings.NewReader(string(body))) //解码图片
	if err != nil {
		return ""
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	err = graphics.Scale(dst, src) //缩小图片
	if err != nil {
		return ""
	}
	buf := bytes.NewBuffer([]byte{})
	err = png.Encode(buf, dst) //编码图片
	if err != nil {
		return ""
	}
	e64 := base64.StdEncoding
	maxEncLen := e64.EncodedLen(buf.Len())
	encBuf := make([]byte, maxEncLen)
	e64.Encode(encBuf, buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", string(encBuf))
}

func GetMd5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}
