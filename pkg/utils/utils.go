package dyndns

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const allowedChars = "0123456789abcdefghijklmnopqrstuvwxyz/_-."

func ContainsInvalidChars(s string) error {
	for i, c := range s {
		if !strings.Contains(allowedChars, strings.ToLower(string(c))) &&
			!strings.Contains(allowedChars, strings.ToUpper(string(c))) {
			return fmt.Errorf("string %s contains forbidden character %d, pos: %d", s, c, i)
		}
	}
	return nil
}

func ContainsValidCharset(charset, s string) error {
	for i, c := range s {
		if !strings.Contains(charset, string(c)) {
			return fmt.Errorf("string %s contains forbidden character %d, pos: %d", s, c, i)
		}
	}
	return nil
}
