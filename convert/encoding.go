package convert

import "github.com/axgle/mahonia"

const (
	CHARSET_GBK  = "GBK"
	CHARSET_UTF8 = "UTF8"
)

// convert specify encode to utf8
func ToUtf8(encode, source string) string {
	e := mahonia.NewDecoder(encode)
	return e.ConvertString(source)
}
