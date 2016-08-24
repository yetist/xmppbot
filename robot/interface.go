package robot

import (
	"github.com/mattn/go-xmpp"
	"time"
)

type AdminIface interface {
	GetRooms() []*Room
	IsAdminID(jid string) bool
	IsFriendID(jid string) bool
	IsCmd(text string) bool
	IsRoomID(jid string) bool
	GetCmdString(cmd string) string
	LoginTime() time.Time
	HasPerm(name string, msg xmpp.Chat) bool
	ShowPerm(name string) string
	SetPerm(name string, perm int)
}

type PluginIface interface {
	Help() string
	GetName() string
	GetSummary() string
	Description() string
	CheckEnv() bool
	Start(bot *Bot)
	Stop()
	Restart()
	Chat(chat xmpp.Chat)
	Presence(pres xmpp.Presence)
	GetOptions() map[string]string
	SetOption(key, val string)
}
