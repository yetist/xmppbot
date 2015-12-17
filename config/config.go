package config

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

func getExecPath() (string, error) {
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
func getExecDir() (string, error) {
	execPath, err := getExecPath()
	return path.Dir(strings.Replace(execPath, "\\", "/", -1)), err
}

// isFile returns true if given path is a file,
// or returns false when it's a directory or does not exist.
func isFile(filePath string) bool {
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
func expandUser(path string) string {
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
	if dir, err := getExecDir(); err != nil || strings.HasSuffix(dir, "_obj/exe") {
		wd, _ := os.Getwd()
		return wd
	} else {
		return dir
	}
}

func userConfigDir(name, version string) (pth string) {
	if pth = os.Getenv("XDG_CONFIG_HOME"); pth == "" {
		pth = expandUser("~/.config")
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
		pth = expandUser(filepath.SplitList(pth)[0])
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
	defer func() {
		config.AppName = name
		config.AppVersion = version
		config.AppConfig = cfgname
	}()
	if isFile(cwdconf) {
		if _, err = toml.DecodeFile(cwdconf, &config); err != nil {
			return
		}
	} else if isFile(selfconf) {
		if _, err = toml.DecodeFile(selfconf, &config); err != nil {
			return
		}
	} else if isFile(userconf) {
		if _, err = toml.DecodeFile(userconf, &config); err != nil {
			return
		}
	} else if isFile(sysconf) {
		if _, err = toml.DecodeFile(sysconf, &config); err != nil {
			return
		}
	} else {
		fmt.Printf("\n*** 无法找到配置文件，有效的配置文件路径列表为(按顺序查找)***\n\n1. %s\n2. %s\n3. %s\n", selfconf, userconf, sysconf)
	}
	return
}
