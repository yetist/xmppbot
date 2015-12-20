package utils

import (
	"errors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	UserAgentCurl    = "curl/7.45.0"
	UserAgentFirefox = "Mozilla/5.0 (X11; Linux x86_64; rv:42.0) Gecko/20100101 Firefox/42.0"
	UserAgentChrome  = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.73 Safari/537.36"
)

func HttpOpen(url string, n time.Duration, agent string) (res *http.Response, body []byte, err error) {
	timeout := n * time.Second
	failTime := timeout * 2
	c := &http.Client{
		Timeout: timeout,
	}
	var req *http.Request

	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return
	}
	if agent == "" {
		req.Header.Set("User-Agent", UserAgentFirefox)
	} else {
		req.Header.Set("User-Agent", agent)
	}
	if res, err = c.Do(req); err != nil {
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

func GetUTF8HtmlTitle(str string) string {
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
