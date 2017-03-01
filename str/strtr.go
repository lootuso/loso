package str

import "strings"

func Replace(str string, pairs map[string]string) string  {
	for k, v := range pairs {
		str = strings.Replace(str, k, v, 1)
	}
	return str
}

