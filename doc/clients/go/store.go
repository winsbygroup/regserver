// sample implementation, do not build or test
//go:build ignore

package main

// Registration file storage for offline license validation.
// Stores registration data as JSON in a well-known location.

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

const (
	companyName = "Company"
	productName = "Product"
	fileName    = "product.json"
)

// GetStorageDirectory returns the directory where registration files are stored.
// Creates the directory if it doesn't exist.
// On Windows: C:\ProgramData\Company\Product
// On Linux/macOS: /var/lib/Company/Product
func GetStorageDirectory() (string, error) {
	var basePath string

	if runtime.GOOS == "windows" {
		basePath = os.Getenv("ProgramData")
		if basePath == "" {
			basePath = `C:\ProgramData`
		}
	} else {
		basePath = "/var/lib"
	}

	path := filepath.Join(basePath, companyName, productName)

	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}

	return path, nil
}

// GetRegistrationFilePath returns the full path to the registration file.
func GetRegistrationFilePath() (string, error) {
	dir, err := GetStorageDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// SaveRegistration saves an activation response to the registration file.
func SaveRegistration(response *ActivationResponse) error {
	path, err := GetRegistrationFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// LoadRegistration loads a previously saved registration from the registration file.
// Returns nil and no error if the file doesn't exist.
func LoadRegistration() (*ActivationResponse, error) {
	path, err := GetRegistrationFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var response ActivationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// RegistrationExists checks if a registration file exists.
func RegistrationExists() bool {
	path, err := GetRegistrationFilePath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

// DeleteRegistration deletes the registration file if it exists.
func DeleteRegistration() error {
	path, err := GetRegistrationFilePath()
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
