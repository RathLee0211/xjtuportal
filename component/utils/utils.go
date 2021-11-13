package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func Sprint(e ...interface{}) string {
	var buf bytes.Buffer
	eLen := len(e)
	for i := 0; i < eLen; i++ {
		switch eType := e[i].(type) {
		case string:
			buf.WriteString(eType)
		case error:
			buf.WriteString(eType.Error())
		case []interface{}:
			for _, elem := range eType {
				e = append(e, elem)
			}
			eLen += len(eType)
		default:
			buf.WriteString(fmt.Sprint(eType))
		}
	}
	return buf.String()
}

func RemoveDuplicateStrings(s []string) (result []string, resultMap map[string]struct{}) {
	result = make([]string, 0)
	resultMap = make(map[string]struct{})
	for _, v := range s {
		if _, ok := resultMap[v]; ok { // Duplicated element found
			continue
		} else {
			resultMap[v] = struct{}{}
			result = append(result, v)
		}
	}
	return
}

func DeleteEmptyString(s []string) []string {
	r := make([]string, 0, len(s))
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func GetCurrentRunningDir() string {
	dir, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(dir)
}
