package plugins

import (
	"fmt"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/core"
	"github.com/yetist/xmppbot/utils"
	"net"
	"strings"
)

type About struct {
	Name string
	bot  *core.Bot
}

func NewAbout(name string, opt map[string]interface{}) *About {
	return &About{Name: name}
}

func (m *About) GetName() string {
	return m.Name
}

func (m *About) GetSummary() string {
	return "关于模块，提供Bot相关的消息。"
}

func (m *About) Help() string {
	msg := []string{
		"关于模块，提供Bot相关的消息。支持命令:",
		m.bot.GetCmdString(m.GetName()) + "    关于模块命令" + m.bot.ShowPerm(m.GetName()),
	}
	return strings.Join(msg, "\n")
}

func (m *About) Description() string {
	return "关于模块，提供Bot版本信息、Bot开发计划及Bot IP地址等消息。"
}

func (m *About) CheckEnv() bool {
	return true
}

func (m *About) Start(bot *core.Bot) {
	fmt.Printf("[%s] Starting...\n", m.GetName())
	m.bot = bot
	m.bot.SetPerm(m.GetName(), core.AllTalk)
}

func (m *About) Stop() {
	fmt.Printf("[%s] Stop\n", m.GetName())
}

func (m *About) Restart() {
	m.Stop()
	m.Start(m.bot)
}

func (m *About) Chat(msg xmpp.Chat) {
	if len(msg.Text) == 0 || !msg.Stamp.IsZero() {
		return
	}
	if strings.HasPrefix(msg.Text, m.bot.GetCmdString(m.GetName())) && m.bot.HasPerm(m.GetName(), msg) {
		cmd := strings.TrimSpace(msg.Text[len(m.bot.GetCmdString(m.GetName())):])
		m.ModCommand(cmd, msg)
	}
}

func (m *About) Presence(pres xmpp.Presence) {
}

func (m *About) GetOptions() map[string]string {
	return map[string]string{}
}

func (m *About) SetOption(key, val string) {
}

func (m *About) ModCommand(cmd string, msg xmpp.Chat) {
	if cmd == "" || cmd == "help" {
		m.cmd_mod_help(cmd, msg)
	} else if cmd == "version" {
		m.cmd_mod_version(cmd, msg)
	} else if cmd == "ip" {
		m.cmd_mod_ip(cmd, msg)
	} else if cmd == "todo" {
		m.cmd_mod_todo(cmd, msg)
	} else {
		m.bot.ReplyAuto(msg, "不支持的命令: "+cmd)
	}
}

func (m *About) cmd_mod_help(cmd string, msg xmpp.Chat) {
	help_msg := []string{"===关于命令===",
		m.bot.GetCmdString(m.Name) + " help     显示本信息",
		m.bot.GetCmdString(m.Name) + " version  显示bot版本信息",
		m.bot.GetCmdString(m.Name) + " ip       显示bot的ip地址",
		m.bot.GetCmdString(m.Name) + " todo     显示bot的开发计划",
	}
	m.bot.ReplyAuto(msg, strings.Join(help_msg, "\n"))
}

func (m *About) cmd_mod_version(cmd string, msg xmpp.Chat) {
	cfg := m.bot.GetConfig()
	m.bot.ReplyPub(msg, cfg.AppName+"-"+cfg.AppVersion)
}

func (m *About) cmd_mod_ip(cmd string, msg xmpp.Chat) {
	local_ip := getLocalIP()
	public_ip := getPubIP()
	txt := "== ip地址信息 == "
	if local_ip != "" {
		txt += "\n" + local_ip
	}
	if public_ip != "" {
		txt += "\n" + public_ip
	}
	m.bot.ReplyAuto(msg, txt)
}

func (m *About) cmd_mod_todo(cmd string, msg xmpp.Chat) {
	text := []string{
		"Bot开发计划：",
		"1. 增加gitlab支持，转发gitlab的项目提交日志",
		"2. 增加github支持，转发github的项目提交日志",
		"3. 增加小i机器人支持，提供小i机器人智能聊天功能",
		"4. 物联网功能，通过bot远程控制主机",
		"5. whois查询",
		"...",
	}
	m.bot.ReplyAuto(msg, strings.Join(text, "\n"))
}

func getLocalIP() (ip string) {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				mask := ipnet.IP.Mask(ipnet.Mask)
				ips = append(ips, "local:"+ipnet.IP.String()+", mask:"+mask.String())
			}

		}
	}
	return strings.Join(ips, "\n")
}

func getPubIP() (ip string) {
	if _, body, err := utils.HttpOpen("http://cip.cc", 5, utils.UserAgentCurl); err == nil {
		tokens := strings.SplitN(string(body), "\n", 2)
		if len(tokens) == 2 {
			ipaddr := tokens[0]
			pos := strings.LastIndex(strings.TrimSpace(ipaddr), " ")
			ip = "public: " + strings.TrimSpace(ipaddr[pos:])
			return
		}
	}
	return ""
}
