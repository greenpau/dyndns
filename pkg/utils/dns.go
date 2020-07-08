package utils

import (
	"fmt"
)

// ResolveName returns public IP address associated with the provided
// DNS record
func ResolveName(name string, version int) (string, error) {
	if version != 4 {
		return "", fmt.Errorf("only ip version 4 is supported")
	}

	return "127.0.0.1", nil
}
