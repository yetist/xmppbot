package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/mattn/go-xmpp"
	"log"
	"os"
	"strconv"
	"strings"
)

func init() {
	flag.StringVar(&config.Account.Username, "username", config.Account.Username, "username")
	flag.StringVar(&config.Account.Password, "password", config.Account.Password, "password")
	flag.StringVar(&config.Account.Domain, "domain", config.Account.Domain, "server")
	flag.StringVar(&config.Account.Server, "server", config.Account.Server, "server")
	flag.IntVar(&config.Account.Port, "port", config.Account.Port, "port")
	flag.StringVar(&config.Account.Status, "status", config.Account.Status, "status")
	flag.StringVar(&config.Account.StatusMessage, "status-msg", config.Account.StatusMessage, "status message")
	flag.BoolVar(&config.Bot.NoTLS, "notls", config.Bot.NoTLS, "No TLS")
	flag.BoolVar(&config.Bot.Debug, "debug", config.Bot.Debug, "debug output")
	flag.BoolVar(&config.Bot.Session, "session", config.Bot.Session, "use server session")
}

func NewClient() (talk *xmpp.Client, err error) {
	options := xmpp.Options{
		Host:          config.Account.Server + ":" + strconv.Itoa(config.Account.Port),
		User:          config.Account.Username + "@" + config.Account.Domain,
		Password:      config.Account.Password,
		Resource:      config.Account.Resource,
		NoTLS:         config.Bot.NoTLS,
		Debug:         config.Bot.Debug,
		Session:       config.Bot.Session,
		Status:        config.Account.Status,
		StatusMessage: config.Account.StatusMessage,
	}

	talk, err = options.NewClient()
	return
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: xmppbot [options]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	flag.Parse()

	if config.Account.Username == "" || config.Account.Password == "" {
		if config.Bot.Debug && config.Account.Username == "" && config.Account.Password == "" {
			fmt.Fprintf(os.Stderr, "no Username or Password were given; attempting ANONYMOUS auth\n")
		} else if config.Account.Username != "" || config.Account.Password != "" {
			flag.Usage()
		}
	}
	if !IsValidStatus(config.Account.Status) {
		fmt.Fprintf(os.Stderr, "invalid status setup, allowed are: away, chat, dnd, xa.\n")
		os.Exit(1)
	}

	PluginInit()
	xmpp.DefaultConfig = tls.Config{
		ServerName:         config.Account.Server,
		InsecureSkipVerify: config.Bot.NoTLS, //如果没有tls，就跳过检查。
	}

	talk, err := NewClient()

	if err != nil {
		log.Fatal(err)
	}

	PluginBegin(talk)

	go func() {
		for {
			chat, err := talk.Recv()
			if err != nil {
				log.Fatal(err)
			}
			switch v := chat.(type) {
			case xmpp.Chat:
				PluginChat(v)
			case xmpp.Presence:
				PluginPresence(v)
			}
		}
	}()
	for {
		in := bufio.NewReader(os.Stdin)
		line, err := in.ReadString('\n')
		if err != nil {
			continue
		}
		line = strings.TrimRight(line, "\n")

		tokens := strings.SplitN(line, " ", 2)
		if len(tokens) == 2 {
			talk.Send(xmpp.Chat{Remote: tokens[0], Type: "chat", Text: tokens[1]})
		}
	}
}
