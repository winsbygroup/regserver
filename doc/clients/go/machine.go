// sample implementation, do not build or test
//go:build ignore

package main

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

func GetMachineCode() string {
	raw := strings.Join([]string{
		getMachineID(),
		getCPUID(),
		getDiskID(),
	}, "|")

	sum := sha256.Sum256([]byte(raw))
	enc := base64.RawURLEncoding.EncodeToString(sum[:])

	return enc
}
