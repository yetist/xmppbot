package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

type Config struct {
	Account struct {
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
		WebHost       string
		WebPort       int
		Rooms         []map[string]interface{}
	}
	Plugin map[string]map[string]interface{}
}

func (c *Config) GetWebHost() string {
	return c.Setup.WebHost
}

func (c *Config) GetWebPort() int {
	return c.Setup.WebPort
}

func (c *Config) GetPlugins() map[string]map[string]interface{} {
	return c.Plugin
}

func (c *Config) GetUsername() string {
	return c.Account.Username
}

func (c *Config) GetResource() string {
	return c.Account.Resource
}

func (c *Config) GetServer() string {
	return c.Account.Server
}

func (c *Config) GetAdmin() []string {
	return c.Setup.Admin
}

func (c *Config) GetRooms() []map[string]interface{} {
	return c.Setup.Rooms
}

func (c *Config) GetCmdPrefix() string {
	return c.Setup.CmdPrefix
}
func (c *Config) GetAutoSubscribe() bool {
	return c.Setup.AutoSubscribe
}
func (c *Config) GetStatus() string {
	return c.Setup.Status
}
func (c *Config) GetStatusMessage() string {
	return c.Setup.StatusMessage
}
func (c *Config) GetDebug() bool {
	return c.Setup.Debug
}

//func init() {
//	LoadConfig(AppName, AppVersion, AppConfig)
//}

func ExecPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	p, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	return p, nil
}

// WorkDir returns absolute path of work directory.
func ExecDir() (string, error) {
	execPath, err := ExecPath()
	return path.Dir(strings.Replace(execPath, "\\", "/", -1)), err
}

// IsFile returns true if given path is a file,
// or returns false when it's a directory or does not exist.
func IsFile(filePath string) bool {
	f, e := os.Stat(filePath)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

// ExpandUser is a helper function that expands the first '~' it finds in the
// passed path with the home directory of the current user.
//
// Note: This only works on environments similar to bash.
func ExpandUser(path string) string {
	if u, err := user.Current(); err == nil {
		return strings.Replace(path, "~", u.HomeDir, -1)
	}
	return path
}

func cwdDir() string {
	cwd, _ := os.Getwd()
	return cwd
}

func selfConfigDir() string {
	if dir, err := ExecDir(); err != nil || strings.HasSuffix(dir, "_obj/exe") {
		wd, _ := os.Getwd()
		return wd
	} else {
		return dir
	}
}

func userConfigDir(name, version string) (pth string) {
	if pth = os.Getenv("XDG_CONFIG_HOME"); pth == "" {
		pth = ExpandUser("~/.config")
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
		pth = ExpandUser(filepath.SplitList(pth)[0])
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
	cwdconf := path.Join(cwdDir(), cfgname)
	if IsFile(cwdconf) {
		if _, err = toml.DecodeFile(cwdconf, &config); err != nil {
			return
		}
	} else if IsFile(selfconf) {
		if _, err = toml.DecodeFile(selfconf, &config); err != nil {
			return
		}
	} else if IsFile(userconf) {
		if _, err = toml.DecodeFile(userconf, &config); err != nil {
			return
		}
	} else if IsFile(sysconf) {
		if _, err = toml.DecodeFile(sysconf, &config); err != nil {
			return
		}
	} else {
		fmt.Printf("\n*** 无法找到配置文件，有效的配置文件路径列表为(按顺序查找)***\n\n1. %s\n2. %s\n3. %s\n", selfconf, userconf, sysconf)
	}
	return
}

func GetDataPath(datafile string) string {
	syspath := path.Join(sysConfigDir(AppName, AppVersion), "data", datafile)
	userpath := path.Join(userConfigDir(AppName, AppVersion), "data", datafile)
	selfpath := path.Join(selfConfigDir(), "data", datafile)
	cwdpath := path.Join(cwdDir(), "data", datafile)
	if IsFile(cwdpath) {
		return cwdpath
	} else if IsFile(selfpath) {
		return selfpath
	} else if IsFile(userpath) {
		return userpath
	} else if IsFile(syspath) {
		return syspath
	}
	return ""
}
