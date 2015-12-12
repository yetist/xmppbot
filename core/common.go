package core

import (
	"crypto/md5"
	"fmt"
	"io"
	"sort"
	"strings"
)

func SplitJID(jid string) (string, string) {
	if strings.Contains(jid, "/") {
		tokens := strings.SplitN(jid, "/", 2)
		return tokens[0], tokens[1]
	} else {
		return jid, ""
	}
}

func IsValidStatus(status string) bool {
	switch status {
	case
		"away",
		"chat",
		"dnd",
		"xa":
		return true
	}
	return false
}

func StringToBool(val string) bool {
	switch strings.ToLower(val) {
	case
		"1",
		"true",
		"t",
		"y",
		"yes",
		"ok":
		return true
	}
	return false
}

func BoolToString(val bool) string {
	if val {
		return "true"
	} else {
		return "false"
	}
}

func SortMapKeys(dict map[string]string) []string {
	keys := make([]string, 0, len(dict))
	for key := range dict {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func ListDelete(list []string, key string) []string {
	for k, v := range list {
		if v == key {
			list = append(list[:k], list[k+1:]...)
		}
	}
	return list
}

func GetMd5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}
