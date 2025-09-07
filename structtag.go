package inject

import (
	"errors"
	"strconv"
)

var errInvalidTag = errors.New("invalid tag")

// Extract 提取给定名称的引用值，如果找到则返回它。
// found布尔值有助于区分默认空字符串的"空且找到"与"空且未找到"的性质。
func Extract(name, tag string) (found bool, value string, err error) {
	for tag != "" {
		// 跳过前导空格
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// 扫描到冒号。
		// 空格或引号是语法错误
		i = 0
		for i < len(tag) && tag[i] != ' ' && tag[i] != ':' && tag[i] != '"' {
			i++
		}
		if i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			return false, "", errInvalidTag
		}
		foundName := tag[:i]
		tag = tag[i+1:]

		// 扫描引用字符串以查找值
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			return false, "", errInvalidTag
		}
		qValue := tag[:i+1]
		tag = tag[i+1:]

		if foundName == name {
			value, err = strconv.Unquote(qValue)
			if err != nil {
				return false, "", err
			}
			return true, value, nil
		}
	}
	return false, "", nil
}
