package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/mattn/go-xmpp"
	"log"
	"os"
	"strconv"
)

const (
	AppName    = "xmppbot"
	AppVersion = "0.1"
	AppConfig  = "xmppbot.toml"
)

var config Config

func init() {
	config, _ = LoadConfig(AppName, AppVersion, AppConfig)
	flag.StringVar(&config.Account.Username, "username", config.Account.Username, "username")
	flag.StringVar(&config.Account.Password, "password", config.Account.Password, "password")
	flag.StringVar(&config.Account.Resource, "resource", config.Account.Resource, "resource")
	flag.StringVar(&config.Account.Server, "server", config.Account.Server, "server")
	flag.IntVar(&config.Account.Port, "port", config.Account.Port, "port")
	flag.BoolVar(&config.Account.NoTLS, "notls", config.Account.NoTLS, "No TLS")
	flag.BoolVar(&config.Account.Session, "session", config.Account.Session, "use server session")

	flag.BoolVar(&config.Setup.Debug, "debug", config.Setup.Debug, "debug output")
	flag.StringVar(&config.Setup.Status, "status", config.Setup.Status, "status")
	flag.StringVar(&config.Setup.StatusMessage, "status-msg", config.Setup.StatusMessage, "status message")
}

func NewClient() (client *xmpp.Client, err error) {
	options := xmpp.Options{
		Host:          config.Account.Server + ":" + strconv.Itoa(config.Account.Port),
		User:          config.Account.Username,
		Password:      config.Account.Password,
		Resource:      config.Account.Resource,
		NoTLS:         config.Account.NoTLS,
		Session:       config.Account.Session,
		Debug:         config.Setup.Debug,
		Status:        config.Setup.Status,
		StatusMessage: config.Setup.StatusMessage,
	}

	if config.Account.SelfSign || !config.Account.NoTLS {
		options.TLSConfig = &tls.Config{
			ServerName:         config.Account.Server,
			InsecureSkipVerify: config.Account.SelfSign, //如果没有tls，就跳过检查。
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

	if config.Account.Username == "" || config.Account.Password == "" {
		if config.Setup.Debug && config.Account.Username == "" && config.Account.Password == "" {
			fmt.Fprintf(os.Stderr, "no Username or Password were given; attempting ANONYMOUS auth\n")
		} else if config.Account.Username != "" || config.Account.Password != "" {
			flag.Usage()
		}
	}
	if !IsValidStatus(config.Setup.Status) {
		fmt.Fprintf(os.Stderr, "invalid status setup, allowed are: away, chat, dnd, xa.\n")
		os.Exit(1)
	}
}

// 新增模块在这里注册
func CreatePlugin(name string, opt map[string]interface{}) BotIface {
	var plugin BotIface
	if name == "auto-reply" {
		plugin = NewAutoReply(name, opt)
	} else if name == "url-helper" {
		plugin = NewUrlHelper(name, opt)
	} else if name == "tuling" {
		plugin = NewTuling(name, opt)
	} else if name == "logger" {
		plugin = NewLogger(name, opt)
	}
	return plugin
}

func main() {
	var client *xmpp.Client
	var err error
	parseArgs()

	if client, err = NewClient(); err != nil {
		log.Fatal(err)
	}

	bot := NewBot(client, config, CreatePlugin)
	//bot.Init()
	bot.Start()
	bot.Run()
}
