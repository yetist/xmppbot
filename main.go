package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/mattn/go-xmpp"
	"github.com/yetist/xmppbot/config"
	"github.com/yetist/xmppbot/plugins"
	"github.com/yetist/xmppbot/robot"
	"github.com/yetist/xmppbot/utils"
	"log"
	"os"
	"strconv"
	"time"
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
func CreatePlugin(name string, opt map[string]interface{}) robot.PluginIface {
	var plugin robot.PluginIface
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

	parseArgs()
	quit := make(chan bool)

	for {
		if client, err = NewClient(); err != nil {
			log.Fatal(err)
		}

		bot := robot.NewBot(client, cfg, CreatePlugin)
		bot.Start()
		bot.Run(quit)

		select {
		case <-quit:
			fmt.Printf("Got Stop\n")
		default:
			fmt.Printf("Got some thing\n")
		}
		bot.Stop()
		time.Sleep(10 * time.Second)
	}
}
