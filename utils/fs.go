package utils

import (
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

func GetExecPath() (string, error) {
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
func GetExecDir() (string, error) {
	execPath, err := GetExecPath()
	return path.Dir(strings.Replace(execPath, "\\", "/", -1)), err
}

// isFile returns true if given path is a file,
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

func CwdDir() string {
	cwd, _ := os.Getwd()
	return cwd
}
