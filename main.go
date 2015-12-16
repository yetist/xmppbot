package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/config"
	"github.com/yetist/xmppbot/core"
	"github.com/yetist/xmppbot/plugins"
	"github.com/yetist/xmppbot/utils"
	"log"
	"os"
	"strconv"
)

const (
	AppName    = "xmppbot"
	AppVersion = "0.1"
	AppConfig  = "xmppbot.toml"
)

var cfg config.Config

func init() {
	cfg, _ = config.LoadConfig(AppName, AppVersion, AppConfig)
	flag.StringVar(&cfg.Account.Username, "username", cfg.Account.Username, "username")
	flag.StringVar(&cfg.Account.Password, "password", cfg.Account.Password, "password")
	flag.StringVar(&cfg.Account.Resource, "resource", cfg.Account.Resource, "resource")
	flag.StringVar(&cfg.Account.Server, "server", cfg.Account.Server, "server")
	flag.IntVar(&cfg.Account.Port, "port", cfg.Account.Port, "port")
	flag.BoolVar(&cfg.Account.NoTLS, "notls", cfg.Account.NoTLS, "No TLS")
	flag.BoolVar(&cfg.Account.Session, "session", cfg.Account.Session, "use server session")

	flag.BoolVar(&cfg.Setup.Debug, "debug", cfg.Setup.Debug, "debug output")
	flag.StringVar(&cfg.Setup.Status, "status", cfg.Setup.Status, "status")
	flag.StringVar(&cfg.Setup.StatusMessage, "status-msg", cfg.Setup.StatusMessage, "status message")
}

func NewClient() (client *xmpp.Client, err error) {
	options := xmpp.Options{
		Host:          cfg.Account.Server + ":" + strconv.Itoa(cfg.Account.Port),
		User:          cfg.Account.Username,
		Password:      cfg.Account.Password,
		Resource:      cfg.Account.Resource,
		NoTLS:         cfg.Account.NoTLS,
		Session:       cfg.Account.Session,
		Debug:         cfg.Setup.Debug,
		Status:        cfg.Setup.Status,
		StatusMessage: cfg.Setup.StatusMessage,
	}

	if cfg.Account.SelfSign || !cfg.Account.NoTLS {
		options.TLSConfig = &tls.Config{
			ServerName:         cfg.Account.Server,
			InsecureSkipVerify: cfg.Account.SelfSign, //如果没有tls，就跳过检查。
		}
	}
	client, err = options.NewClient()
	return
}

func parseArgs() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: xmppbot [options]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	flag.Parse()

	if cfg.Account.Username == "" || cfg.Account.Password == "" {
		if cfg.Setup.Debug && cfg.Account.Username == "" && cfg.Account.Password == "" {
			fmt.Fprintf(os.Stderr, "no Username or Password were given; attempting ANONYMOUS auth\n")
		} else if cfg.Account.Username != "" || cfg.Account.Password != "" {
			flag.Usage()
		}
	}
	if !utils.IsValidStatus(cfg.Setup.Status) {
		fmt.Fprintf(os.Stderr, "invalid status setup, allowed are: away, chat, dnd, xa.\n")
		os.Exit(1)
	}
}

// 新增模块在这里注册
func CreatePlugin(name string, opt map[string]interface{}) core.PluginIface {
	var plugin core.PluginIface
	switch name {
	case "random":
		plugin = plugins.NewRandom(name, opt)
	case "url":
		plugin = plugins.NewUrl(name, opt)
	case "tuling":
		plugin = plugins.NewTuling(name, opt)
	case "logger":
		plugin = plugins.NewLogger(name, opt)
	case "notify":
		plugin = plugins.NewNotify(name, opt)
	case "about":
		plugin = plugins.NewAbout(name, opt)
	}
	return plugin
}

func main() {
	var client *xmpp.Client
	var err error
	/*
		var newplugin = map[string]func(name string, opt map[string]interface{}) core.BotIface{
			"random": func(name string, opt map[string]interface{}) core.BotIface { return NewRandom(name, opt) },
			"url":    func(name string, opt map[string]interface{}) core.BotIface { return NewUrl(name, opt) },
			"tuling": func(name string, opt map[string]interface{}) core.BotIface { return NewTuling(name, opt) },
			"logger": func(name string, opt map[string]interface{}) core.BotIface { return NewLogger(name, opt) },
			"notify": func(name string, opt map[string]interface{}) core.BotIface { return NewNotify(name, opt) },
			"about":  func(name string, opt map[string]interface{}) core.BotIface { return NewAbout(name, opt) },
		}
	*/

	parseArgs()

	if client, err = NewClient(); err != nil {
		log.Fatal(err)
	}

	bot := core.NewBot(client, cfg, CreatePlugin)
	bot.Start()
	bot.Run()
}
