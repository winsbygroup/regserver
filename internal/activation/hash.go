package activation

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// buildRegistrationString builds the pipe-delimited string used for hash computation.
// Format: machineCode|expDate|maintDate|maxVersion|Feature1=Value1|Feature2=Value2|...
// Features are sorted alphabetically by name.
func buildRegistrationString(
	machineCode string,
	expDate string,
	maintDate string,
	maxVersion string,
	features map[string]string,
) string {
	var b strings.Builder

	b.WriteString(machineCode)
	b.WriteString("|")
	b.WriteString(expDate)
	b.WriteString("|")
	b.WriteString(maintDate)
	b.WriteString("|")
	b.WriteString(maxVersion)

	// Sort feature keys alphabetically
	keys := make([]string, 0, len(features))
	for k := range features {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		b.WriteString("|")
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(features[k])
	}

	return b.String()
}

// computeRegistrationHash computes SHA1 hash of UTF-16LE encoded string with appended secret.
// The secret is appended to prevent users from forging valid hashes.
// Format: SHA1(UTF16LE(regStr + secret)) -> Base64
func computeRegistrationHash(regStr, secret string) (string, error) {
	// Append secret to registration string before encoding
	combined := regStr + secret

	enc := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()

	utf16Str, _, err := transform.String(enc, combined)
	if err != nil {
		return "", fmt.Errorf("encode UTF-16LE: %w", err)
	}

	h := sha1.New()
	h.Write([]byte(utf16Str))
	digest := h.Sum(nil) // 20 bytes

	return base64.StdEncoding.EncodeToString(digest), nil
}
