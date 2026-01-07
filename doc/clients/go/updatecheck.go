// sample implementation, do not build or test
//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type LicenseInfoRequest struct {
	MachineCode      string `json:"machineCode"`
	InstalledVersion string `json:"installedVersion"`
}

type LicenseInfoResponse struct {
	CustomerName        string         `json:"CustomerName"`
	ProductGUID         string         `json:"ProductGUID"`
	ProductName         string         `json:"ProductName"`
	LicenseCount        int            `json:"LicenseCount"`
	LicensesAvailable   int            `json:"LicensesAvailable"`
	ExpirationDate      string         `json:"ExpirationDate"`
	MaintExpirationDate string         `json:"MaintExpirationDate"`
	MaxProductVersion   string         `json:"MaxProductVersion"`
	LatestVersion       string         `json:"LatestVersion"`
	Features            map[string]any `json:"Features"`
}

type ProductVersionResponse struct {
	ProductGUID   string `json:"ProductGUID"`
	LatestVersion string `json:"LatestVersion"`
	DownloadURL   string `json:"DownloadURL"`
}

type UpdateInfo struct {
	UpdateAvailable  bool
	CurrentVersion   string
	LatestVersion    string
	DownloadURL      string
}

// UpdateInstalledVersion reports the installed version to the server and returns license info.
func UpdateInstalledVersion(baseURL, licenseKey, machineCode, installedVersion string) (*LicenseInfoResponse, error) {
	reqBody := LicenseInfoRequest{
		MachineCode:      machineCode,
		InstalledVersion: installedVersion,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("PUT", baseURL+"/api/v1/license/"+licenseKey, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update failed: %s", resp.Status)
	}

	var result LicenseInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// GetProductVersion retrieves the latest version and download URL for a product.
func GetProductVersion(baseURL, productGUID string) (*ProductVersionResponse, error) {
	resp, err := http.Get(baseURL + "/api/v1/productver/" + productGUID)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get product version failed: %s", resp.Status)
	}

	var result ProductVersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// CheckForUpdate reports the installed version and checks if an update is available.
// Returns update info including the download URL if an update is available.
func CheckForUpdate(baseURL, licenseKey, machineCode, installedVersion string) (*UpdateInfo, error) {
	// Step 1: Report installed version and get license info
	licenseInfo, err := UpdateInstalledVersion(baseURL, licenseKey, machineCode, installedVersion)
	if err != nil {
		return nil, fmt.Errorf("update installed version: %w", err)
	}

	info := &UpdateInfo{
		CurrentVersion: installedVersion,
		LatestVersion:  licenseInfo.LatestVersion,
	}

	// Step 2: Check if update is available
	if CompareVersions(installedVersion, licenseInfo.LatestVersion) >= 0 {
		info.UpdateAvailable = false
		return info, nil
	}

	// Step 3: Get download URL
	productInfo, err := GetProductVersion(baseURL, licenseInfo.ProductGUID)
	if err != nil {
		return nil, fmt.Errorf("get product version: %w", err)
	}

	info.UpdateAvailable = true
	info.DownloadURL = productInfo.DownloadURL

	return info, nil
}
