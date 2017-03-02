package strings

import "strings"

func ReplaceAll(str string, pairs map[string]string) string  {
	for k, v := range pairs {
		str = strings.Replace(str, k, v, 1)
	}
	return str
}
