package secure

import (
	"crypto/rand"
	"math/big"
)

const (
	DefaultPasswordSymbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+,.?/:;{}[]`~"
)

// RandString randomly generates a string of length l from the given symbols.
func RandString(l int, s string) (string, error) {
	if s == "" || l <= 0 {
		return "", nil
	}
	c := []rune(s)
	m := big.NewInt(int64(len(c)))
	var r = make([]rune, l)
	for i := 0; i < l; i++ {
		if v, err := rand.Int(rand.Reader, m); err == nil {
			r[i] = c[v.Int64()]
		} else {
			return "", err
		}
	}
	return string(r), nil
}

// MaskString masks the given string.
// The middle third of the given string will be masked with asterisks.
func MaskString(origin string) string {
	runes := []rune(origin)
	l := len(runes)
	s := l / 3
	e := s * 2
	if l%3 > 0 {
		e++
	}
	for i := s; i < e && i < l; i++ {
		runes[i] = '*'
	}
	return string(runes)
}
