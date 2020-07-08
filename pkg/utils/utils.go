package dyndns

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const allowedChars = "0123456789abcdefghijklmnopqrstuvwxyz/_-."

// ContainsInvalidChars returns error if the provided string contains
// characters outside of the allowedChars character set.
func ContainsInvalidChars(s string) error {
	for i, c := range s {
		if !strings.Contains(allowedChars, strings.ToLower(string(c))) &&
			!strings.Contains(allowedChars, strings.ToUpper(string(c))) {
			return fmt.Errorf("string %s contains forbidden character %d, pos: %d", s, c, i)
		}
	}
	return nil
}

// ContainsValidCharset returns error if the provided string contains
// characters outside of the provided character set.
func ContainsValidCharset(charset, s string) error {
	for i, c := range s {
		if !strings.Contains(charset, string(c)) {
			return fmt.Errorf("string %s contains forbidden character %d, pos: %d", s, c, i)
		}
	}
	return nil
}
