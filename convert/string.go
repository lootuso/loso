package convert

import "strconv"

//parse string value to int
func ParseInt(s string) int64 {
	if v, err := strconv.ParseInt(s, 10, 0); err == nil {
		return v
	}
	return 0
}


// parse string value to float
func ParseFloat(s string) float64 {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return 0
}
