package utils

import (
	"fmt"
	"strings"
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

// MaskSecret masks secret strings
func MaskSecret(s string, j int, k int) string {
	r := ""
	mask := false
	if k == 0 {
		k = len(s)
	} else {
		k = len(s) - k + 1
	}

	for i, c := range s {
		if i >= j {
			mask = true
		}
		if i > k {
			mask = false
		}
		if mask {
			r = r + "*"
			continue
		}
		r = r + string(c)
	}
	return r
}
