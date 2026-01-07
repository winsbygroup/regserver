// sample implementation, do not build or test
//go:build ignore

package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
)

type ActivationRequest struct {
	MachineCode string `json:"machineCode"`
	UserName    string `json:"userName"`
}

type ActivationResponse struct {
	UserName            string         `json:"UserName"`
	UserCompany         string         `json:"UserCompany"`
	ProductGUID         string         `json:"ProductGUID"`
	MachineCode         string         `json:"MachineCode"`
	ExpirationDate      string         `json:"ExpirationDate"`
	MaintExpirationDate string         `json:"MaintExpirationDate"`
	MaxProductVersion   string         `json:"MaxProductVersion"`
	LatestVersion       string         `json:"LatestVersion"`
	LicenseKey          string         `json:"LicenseKey"`
	RegistrationHash    string         `json:"RegistrationHash"`
	Features            map[string]any `json:"Features"`
}

func ActivateProduct(baseURL, licenseKey, machineCode, userName string) (*ActivationResponse, error) {
	reqBody := ActivationRequest{
		MachineCode: machineCode,
		UserName:    userName,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/api/v1/activate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-License-Key", licenseKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("activation failed: %s", resp.Status)
	}

	var result ActivationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// CalculateRegistrationHash computes the registration hash for offline license validation.
// The algorithm:
// 1. Build string: {MachineCode}|{ExpirationDate}|{MaintExpirationDate}|{MaxProductVersion}|{Feature1}={Value1}|...
// 2. Append the secret
// 3. Encode as UTF-16LE (no BOM)
// 4. Compute SHA1 hash
// 5. Base64 encode
func CalculateRegistrationHash(machineCode, expirationDate, maintExpirationDate, maxProductVersion, secret string, features map[string]any) string {
	// Step 1: Build the registration string
	regString := machineCode + "|" + expirationDate + "|" + maintExpirationDate + "|" + maxProductVersion

	// Add features (sorted alphabetically by key)
	if len(features) > 0 {
		keys := make([]string, 0, len(features))
		for k := range features {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			regString += "|" + k + "=" + fmt.Sprintf("%v", features[k])
		}
	}

	// Step 2: Append the secret
	regString += secret

	// Step 3: Encode as UTF-16LE (no BOM)
	utf16Bytes := encodeUTF16LE(regString)

	// Step 4: Compute SHA1 hash
	hash := sha1.Sum(utf16Bytes)

	// Step 5: Base64 encode
	return base64.StdEncoding.EncodeToString(hash[:])
}

// encodeUTF16LE converts a string to UTF-16 Little Endian bytes (no BOM).
func encodeUTF16LE(s string) []byte {
	runes := []rune(s)
	buf := new(bytes.Buffer)

	for _, r := range runes {
		if r <= 0xFFFF {
			binary.Write(buf, binary.LittleEndian, uint16(r))
		} else {
			// Surrogate pair for characters outside BMP
			r -= 0x10000
			high := uint16(0xD800 + (r >> 10))
			low := uint16(0xDC00 + (r & 0x3FF))
			binary.Write(buf, binary.LittleEndian, high)
			binary.Write(buf, binary.LittleEndian, low)
		}
	}

	return buf.Bytes()
}

// ValidateRegistration verifies an activation response by comparing the calculated hash
// with the server-provided hash.
func ValidateRegistration(response *ActivationResponse, secret string) bool {
	calculatedHash := CalculateRegistrationHash(
		response.MachineCode,
		response.ExpirationDate,
		response.MaintExpirationDate,
		response.MaxProductVersion,
		secret,
		response.Features,
	)

	return calculatedHash == response.RegistrationHash
}

// IsVersionAllowed checks if the installed product version is allowed based on MaxProductVersion.
// Returns true if no restriction (empty MaxProductVersion) or if installedVersion <= maxProductVersion.
func IsVersionAllowed(installedVersion, maxProductVersion string) bool {
	if maxProductVersion == "" {
		return true // No restriction
	}
	return CompareVersions(installedVersion, maxProductVersion) <= 0
}

// CompareVersions compares two semver version strings (e.g., "1.2.3").
// Returns negative if a < b, zero if a == b, positive if a > b.
func CompareVersions(a, b string) int {
	aParts := parseVersion(a)
	bParts := parseVersion(b)

	if aParts[0] != bParts[0] {
		return aParts[0] - bParts[0]
	}
	if aParts[1] != bParts[1] {
		return aParts[1] - bParts[1]
	}
	return aParts[2] - bParts[2]
}

func parseVersion(v string) [3]int {
	var parts [3]int
	fmt.Sscanf(v, "%d.%d.%d", &parts[0], &parts[1], &parts[2])
	return parts
}
