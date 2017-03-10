package random

import (
	"math/rand"
	"time"
)
// generate random int
func Int(a, b int) int {
	if a > b {
		a, b = b, a
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(b-a+1) + a
}
// generate random int
func Int64(a, b int64) int64 {
	if a > b {
		a, b = b, a
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(b-a+1) + a
}

// generate random time
func Time(a, b time.Time) time.Time {
	rand.Seed(time.Now().UnixNano())
	return time.Unix(Int64(a.Unix(), b.Unix()), 0)
}

// generate random string
func String(length int)string  {
	buffer := make([]byte, length)
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := range buffer {
		buffer[i] = charset[rand.Int63()%int64(len(charset))]
	}
	return string(buffer)
}