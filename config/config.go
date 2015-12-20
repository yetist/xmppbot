package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/yetist/xmppbot/utils"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Config struct {
	AppName    string
	AppVersion string
	AppConfig  string
	Account    struct {
		Username string
		Password string
		Resource string
		Server   string
		Port     int
		NoTLS    bool
		SelfSign bool `toml:"self_signed"`
		Session  bool
	}
	Setup struct {
		Admin         []string
		Debug         bool
		AutoSubscribe bool   `toml:"auto_subscribe"`
		CmdPrefix     string `toml:"cmd_prefix"`
		Status        string
		StatusMessage string `toml:"status_message"`
		WebHost       string `toml:"web_host"`
		WebPort       int    `toml:"web_port"`
		Rooms         []map[string]interface{}
	}
	Plugin map[string]map[string]interface{}
}

func selfConfigDir() string {
	if dir, err := utils.GetExecDir(); err != nil || strings.HasSuffix(dir, "_obj/exe") {
		wd, _ := os.Getwd()
		return wd
	} else {
		return dir
	}
}

func userConfigDir(name, version string) (pth string) {
	if pth = os.Getenv("XDG_CONFIG_HOME"); pth == "" {
		pth = utils.ExpandUser("~/.config")
	}

	if name != "" {
		pth = filepath.Join(pth, name)
	}

	if version != "" {
		pth = filepath.Join(pth, version)
	}

	return pth
}

func sysConfigDir(name, version string) (pth string) {
	if pth = os.Getenv("XDG_CONFIG_DIRS"); pth == "" {
		pth = "/etc/xdg"
	} else {
		pth = utils.ExpandUser(filepath.SplitList(pth)[0])
	}
	if name != "" {
		pth = filepath.Join(pth, name)
	}

	if version != "" {
		pth = filepath.Join(pth, version)
	}
	return pth
}

func LoadConfig(name, version, cfgname string) (config Config, err error) {
	sysconf := path.Join(sysConfigDir(name, version), cfgname)
	userconf := path.Join(userConfigDir(name, version), cfgname)
	selfconf := path.Join(selfConfigDir(), cfgname)
	cwdconf := path.Join(utils.CwdDir(), cfgname)
	defer func() {
		config.AppName = name
		config.AppVersion = version
		config.AppConfig = cfgname
	}()
	if utils.IsFile(cwdconf) {
		if _, err = toml.DecodeFile(cwdconf, &config); err != nil {
			return
		}
	} else if utils.IsFile(selfconf) {
		if _, err = toml.DecodeFile(selfconf, &config); err != nil {
			return
		}
	} else if utils.IsFile(userconf) {
		if _, err = toml.DecodeFile(userconf, &config); err != nil {
			return
		}
	} else if utils.IsFile(sysconf) {
		if _, err = toml.DecodeFile(sysconf, &config); err != nil {
			return
		}
	} else {
		fmt.Printf("\n*** 无法找到配置文件，有效的配置文件路径列表为(按顺序查找)***\n\n1. %s\n2. %s\n3. %s\n", selfconf, userconf, sysconf)
	}
	return
}
