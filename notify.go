package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/core"
	"github.com/yetist/xmppbot/utils"
	"net"
	"net/http"
	"strings"
)

type Notify struct {
	Name   string
	Allows []string
	Option map[string]string
	bot    *core.Bot
}

func NewNotify(name string, opt map[string]interface{}) *Notify {
	var allows []string
	for _, i := range opt["allows"].([]interface{}) {
		allows = append(allows, i.(string))
	}
	return &Notify{
		Name: name,
		Option: map[string]string{
			"authuser": opt["authuser"].(string),
			"authpass": opt["authpass"].(string),
		},
		Allows: allows,
	}
}

func (m *Notify) GetName() string {
	return m.Name
}

func (m *Notify) GetSummary() string {
	return "通知转发模块，自动转发通过http协议接收到的消息。"
}

func (m *Notify) Description() string {
	return "通知转发模块，可将通过http协议接收到的消息转发给好友或聊天室。"
}

func (m *Notify) Help() string {
	msg := []string{
		"通知转发模块，可将通过http协议接收到的消息转发给好友或聊天室．",
		"支持以下命令:",
		m.bot.GetCmdString(m.GetName()) + "    通知模块命令",
	}
	return strings.Join(msg, "\n")
}

func (m *Notify) CheckEnv() bool {
	return true
}

/* web pages */
func (m *Notify) JIDPage(w http.ResponseWriter, r *http.Request) {
	if !m.isIpAllowed(r) {
		http.NotFound(w, r)
		return
	}
	if username, password, ok := r.BasicAuth(); !ok {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"xmppbot\"")
		http.Error(w, http.StatusText(401), 401)
		return
	} else if !(m.Option["authuser"] == username && m.Option["authpass"] == password) {
		w.Header().Set("WWW-Authenticate", "Basic realm=\"xmppbot\"")
		http.Error(w, http.StatusText(401), 401)
		return
	}

	if strings.ToLower(r.Method) != "post" {
		http.NotFound(w, r)
		return
	}
	vars := mux.Vars(r)
	jid := vars["jid"]
	if m.bot.IsRoomID(jid) {
		m.bot.SendPub(jid, fmt.Sprintf("通知：%s\n%s", r.FormValue("subject"), r.FormValue("body")))
	} else {
		m.bot.SendAuto(jid, fmt.Sprintf("通知：%s\n%s", r.FormValue("subject"), r.FormValue("body")))
	}
	w.Write([]byte("notify sent to " + jid + "\n"))
}

func (m *Notify) Start(bot *core.Bot) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.bot = bot
	m.bot.AddHandler(m.GetName(), "/{jid}/", m.JIDPage, "jidpage")
}

func (m *Notify) Stop() {
	fmt.Printf("[%s] Stop\n", m.GetName())
	m.bot.DelHandler(m.GetName(), "jidpage")
}

func (m *Notify) Restart() {
	m.Stop()
	m.Start(m.bot)
}

func (m *Notify) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}
	if msg.Type == "groupchat" {
		return
	}
	if strings.HasPrefix(msg.Text, m.bot.GetCmdString(m.GetName())) {
		cmd := strings.TrimSpace(msg.Text[len(m.bot.GetCmdString(m.GetName())):])
		m.ModCommand(cmd, msg)
	}
}

func (m *Notify) Presence(pres xmpp.Presence) {
}

func (m *Notify) GetOptions() map[string]string {
	opts := make(map[string]string, 0)
	for k, v := range m.Option {
		if k == "authuser" {
			opts[k] = v + "  #认证用户名"
		} else if k == "authpass" {
			opts[k] = v + "  #认证密码"
		}
	}
	return opts
}

func (m *Notify) SetOption(key, val string) {
	if _, ok := m.Option[key]; ok {
		m.Option[key] = val
	}
}

func (m *Notify) ModCommand(cmd string, msg xmpp.Chat) {
	if cmd == "" || cmd == "help" {
		m.cmd_mod_help(cmd, msg)
	} else if cmd == "list-allows" {
		m.cmd_mod_list_allows(cmd, msg)
	} else if strings.HasPrefix(cmd, "add-allow ") {
		m.cmd_mod_add_allow(cmd, msg)
	} else if strings.HasPrefix(cmd, "del-allow ") {
		m.cmd_mod_del_allow(cmd, msg)
	} else {
		m.bot.ReplyAuto(msg, "不支持的命令: "+cmd)
	}
}

func (m *Notify) cmd_mod_help(cmd string, msg xmpp.Chat) {
	help_msg := []string{"==通知转发命令==",
		m.bot.GetCmdString(m.Name) + " help            显示本信息",
		m.bot.GetCmdString(m.Name) + " list-allows     列出允许访问的ip地址",
		m.bot.GetCmdString(m.Name) + " add-allow <ip>  添加新的ip地址到可允许访问列表",
		m.bot.GetCmdString(m.Name) + " del-allow <ip>  从允许访问列表中删除一个ip地址",
	}
	m.bot.ReplyAuto(msg, strings.Join(help_msg, "\n"))
}

func (m *Notify) cmd_mod_list_allows(cmd string, msg xmpp.Chat) {
	var allows_list []string
	for k, v := range m.Allows {
		allows_list = append(allows_list, fmt.Sprintf("%2d: %s", k+1, v))
	}
	txt := "==允许的ip地址列表==\n" + strings.Join(allows_list, "\n")
	m.bot.ReplyAuto(msg, txt)
}

func (m *Notify) IsAllowed(ip string) bool {
	for _, i := range m.Allows {
		if i == ip {
			return true
		}
	}
	return false
}

func (m *Notify) cmd_mod_add_allow(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if len(tokens) == 2 {
		if m.IsAllowed(tokens[1]) {
			m.bot.ReplyAuto(msg, tokens[1]+" 已经存在于ip地址列表中，不需再次增加！")
		} else {
			m.Allows = append(m.Allows, tokens[1])
			m.bot.ReplyAuto(msg, "您已添加 "+tokens[1]+"到ip地址列表!")
		}
	}
}

func (m *Notify) cmd_mod_del_allow(cmd string, msg xmpp.Chat) {
	tokens := strings.SplitN(cmd, " ", 2)
	if m.IsAllowed(tokens[1]) {
		m.Allows = utils.ListDelete(m.Allows, tokens[1])
		m.bot.ReplyAuto(msg, "禁用了ip地址："+tokens[1])
	} else {
		m.bot.ReplyAuto(msg, "ip地址 "+tokens[1]+" 不在列表中!")
	}
}

func (m *Notify) isIpAllowed(r *http.Request) (authorized bool) {
	var ip string
	if ipProxy := r.Header.Get("X-Real-IP"); len(ipProxy) > 0 {
		ip = ipProxy
	} else if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
		ip = ipProxy
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	rAddr := net.ParseIP(ip)
	for v := range m.Allows {
		_, ipNet, err := net.ParseCIDR(m.Allows[v])
		if err != nil {
			ipHost := net.ParseIP(m.Allows[v])
			if ipHost != nil {
				if ipHost.Equal(rAddr) {
					authorized = true
				}
			}
		} else {
			if ipNet.Contains(rAddr) {
				authorized = true
			}
		}
	}
	fmt.Printf("ip: %s, authorized: %v\n", ip, authorized)
	return
}
