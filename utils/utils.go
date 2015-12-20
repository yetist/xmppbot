package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"sort"
	"strings"
)

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

func SortMapKeys(dict interface{}) []string {
	keys := make([]string, 0, len(dict.(map[string]string)))
	for key := range dict.(map[string]string) {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func MapDelete(d interface{}, key string) {
	switch d.(type) {
	case map[string]interface{}:
		dict := d.(map[string]interface{})
		_, ok := dict[key]
		if ok {
			delete(dict, key)
		}
	}
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
