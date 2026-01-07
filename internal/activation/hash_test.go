package activation

import "testing"

func TestBuildRegistrationString(t *testing.T) {
	machineCode := "ABC123"
	expDate := "2025-12-31"
	maintDate := "2025-12-31"
	maxVersion := "4.5"
	features := map[string]string{
		"FeatureB": "Value2",
		"FeatureA": "Value1",
	}

	got := buildRegistrationString(machineCode, expDate, maintDate, maxVersion, features)
	want := "ABC123|2025-12-31|2025-12-31|4.5|FeatureA=Value1|FeatureB=Value2"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildRegistrationString_EmptyMaxVersion(t *testing.T) {
	machineCode := "ABC123"
	expDate := "2025-12-31"
	maintDate := "2025-12-31"
	maxVersion := ""
	features := map[string]string{
		"FeatureA": "Value1",
	}

	got := buildRegistrationString(machineCode, expDate, maintDate, maxVersion, features)
	want := "ABC123|2025-12-31|2025-12-31||FeatureA=Value1"

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestComputeRegistrationHash(t *testing.T) {
	machineCode := "3V6EC/qizaPlMQgIJaM1oUDRDG8=2jmj7l5rSw0yVb/vlWAYkK/YBwk="
	expDate := "2025-12-31"
	maintDate := "2025-12-31"
	maxVersion := "5.0"
	features := map[string]string{
		"Legacy":     "True",
		"PartTypes":  "999999999",
		"Structured": "True",
	}
	secret := "test-secret-key"

	regStr := buildRegistrationString(machineCode, expDate, maintDate, maxVersion, features)

	got, err := computeRegistrationHash(regStr, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == "" {
		t.Fatal("expected non-empty hash")
	}

	// Verify different secrets produce different hashes
	got2, _ := computeRegistrationHash(regStr, "different-secret")
	if got == got2 {
		t.Error("expected different hashes for different secrets")
	}

	// Verify different maxVersion produces different hash
	regStr2 := buildRegistrationString(machineCode, expDate, maintDate, "6.0", features)
	got3, _ := computeRegistrationHash(regStr2, secret)
	if got == got3 {
		t.Error("expected different hashes for different maxVersion")
	}
}

func TestComputeRegistrationHash_EmptyFeatures(t *testing.T) {
	machineCode := "MACHINE001"
	expDate := "2025-06-15"
	maintDate := "2025-06-15"
	maxVersion := ""
	features := map[string]string{}
	secret := "test-secret-key"

	regStr := buildRegistrationString(machineCode, expDate, maintDate, maxVersion, features)
	want := "MACHINE001|2025-06-15|2025-06-15|"

	if regStr != want {
		t.Errorf("regStr: got %q, want %q", regStr, want)
	}

	hash, err := computeRegistrationHash(regStr, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}
}
