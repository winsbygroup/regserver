# Client Implementation

Software registrations are designed to be validated locally (offline) unless required to activate a new or expired license. 

A local registration file contains a hash calculated from personalized registration information. This hash should be 
re-calculated each time the program runs and checked against the one in the local registration file (to avoid tampering).

### Registration Hash Calculation

**Step 1: Build the registration string**

```
{MachineCode}|{ExpirationDate}|{MaintExpirationDate}|{MaxProductVersion}|{Feature1}={Value1}|{Feature2}={Value2}|...
```

Rules:
- Dates are in `yyyy-mm-dd` format
- `MaxProductVersion` may be empty (but the separator is still included)
- Features are sorted alphabetically by name (key)
- All Feature values are strings (without quotes)
- Vertical bars are used as separators not terminators

**Example (with MaxProductVersion):**
```
3V6EC/qizaPlMQgIJaM1oUDRDG8=2jmj7l5rSw0yVb/vlWAYkK/YBwk=|2025-12-31|2025-12-31|4.5|Legacy=True|PartTypes=999999999|Structured=True
```

**Example (no version restriction):**
```
3V6EC/qizaPlMQgIJaM1oUDRDG8=2jmj7l5rSw0yVb/vlWAYkK/YBwk=|2025-12-31|2025-12-31||Legacy=True|PartTypes=999999999|Structured=True
```

**Step 2: Append the registration secret**

The registration hash is computed by appending a secret key to the registration string before hashing. 
This prevents users from tampering with the registration file. This secret must also be used by the regserver API (to calculate 
the hash) and is referred to there by the env variable `REGISTRATION_SECRET` .

**Step 3: Encode as UTF-16LE**

Convert the combined string to UTF-16 Little Endian bytes (no BOM).

**Step 4: Compute SHA1 hash**

Hash the UTF-16LE bytes using SHA1 (produces 20 bytes).

**Step 5: Base64 encode**

Base64 encode the 20-byte hash to produce the final registration hash.

**Important:** The client software must use the same `REGISTRATION_SECRET` value used by the registration service to validate 
registration files offline. This secret should be embedded in the client application (obfuscated if possible).

For an invalid license, the process would be as follows:

1. Calculate a code that is unique to that machine (platform-specific functions are readily available)
2. Prompt for a UserName
3. Prompt for the LicenseKey provided to them
4. Submit that information to the Activation endpoint
5. Save the results to a local file for future (off-line) license validation

---

## Registration Storage

Each implementation includes a `store` module that handles saving and loading registration data to/from JSON files in 
a well-known location for the product application being registered:

| Platform | Storage Location                                    |
|----------|-----------------------------------------------------|
| Windows | `C:\ProgramData\(company)\(product)\(product).json` |
| Linux/macOS | `/var/lib/(company)/(product)/(product).json`    |

Update the `CompanyName` and `ProductName` constants in the `store` module for your application.

### Storage Functions

| Function | Description |
|----------|-------------|
| `SaveRegistration(response)` | Saves activation response to JSON file |
| `LoadRegistration()` | Loads saved registration (returns null/nil if not found) |
| `RegistrationExists()` | Checks if a registration file exists |
| `DeleteRegistration()` | Deletes the registration file |
| `GetStorageDirectory()` | Returns the storage directory path |
| `GetRegistrationFilePath()` | Returns the full path to the registration file |

---

## Checking License Availability

Before attempting activation, clients can check license availability using the `/license/:license_key` endpoint:

```
GET /api/v1/license/{license_key}
```

**Response:**
```json
{
  "CustomerName": "Acme Corp",
  "ProductGUID": "5177851a-33d6-422f-96df-9ad6b7ff4611",
  "ProductName": "AceMapper",
  "LicenseCount": 5,
  "LicensesAvailable": 3,
  "ExpirationDate": "2025-12-31",
  "MaintExpirationDate": "2025-12-31",
  "MaxProductVersion": "4.5",
  "LatestVersion": "5.5.1.0",
  "Features": {}
}
```

| Field | Description |
|-------|-------------|
| `CustomerName` | Name of the customer who owns this license |
| `LicenseCount` | Total number of licenses purchased |
| `LicensesAvailable` | Remaining licenses available for activation |
| `MaxProductVersion` | Maximum product version allowed (empty = no restriction) |
| `LatestVersion` | Latest available product version |

**Note:** `LicensesAvailable` only counts non-expired machine registrations as "in use". Expired registrations do not reduce the available count.

This endpoint is useful for:
- Displaying license status to users before activation
- Checking if licenses are available before prompting for activation
- Administrative tools that need to show license usage

---

## Updating Installed Version

Clients can report their installed version using the PUT `/license/:license_key` endpoint. This is useful for tracking deployments and prompting users when updates are available.

See implementation examples:
- **Go**: [updatecheck.go](go/updatecheck.go)
- **Delphi**: [updatecheck.pas](delphi/updatecheck.pas)
- **C#**: [updatecheck.cs](csharp/updatecheck.cs)

```
PUT /api/v1/license/{license_key}
Content-Type: application/json

{
  "machineCode": "5mToXAaMQRRXOG58VT2oRKBgD8c=nWxB5pHxLwJx/LbewudPWXecK3c=",
  "installedVersion": "5.5.0"
}
```

**Response:** Same as GET `/license/:license_key`

| Field | Required | Description |
|-------|----------|-------------|
| `machineCode` | Yes | The machine code identifying this installation |
| `installedVersion` | No | The version currently installed |

**Example workflow - Check for Updates:**

1. Call PUT `/license/{license_key}` with current `machineCode` and `installedVersion`
2. Compare `installedVersion` with `LatestVersion` in response
3. If `LatestVersion` is newer:
   - Call GET `/productver/{ProductGUID}` to get the `DownloadURL`
   - Prompt user to update with download link

```
GET /api/v1/productver/5177851a-33d6-422f-96df-9ad6b7ff4611

Response:
{
  "ProductGUID": "5177851a-33d6-422f-96df-9ad6b7ff4611",
  "LatestVersion": "5.5.1",
  "DownloadURL": "https://example.com/downloads/product-5.5.1.zip"
}
```

---

## Usage Examples

### C#

```csharp
// Activate and validate in one step
var client = new LicenseClient("https://license.example.com", licenseKey);
var response = await client.ActivateAsync(machineCode, Environment.UserName);

if (LicenseClient.ValidateRegistration(response, "your-registration-secret"))
{
    // Save response to local file for offline validation
    RegistrationStore.SaveRegistration(response);
}

// Later: Offline validation from saved registration file
var saved = RegistrationStore.LoadRegistration();
if (saved != null)
{
    var calculatedHash = LicenseClient.CalculateRegistrationHash(
        saved.MachineCode,
        saved.ExpirationDate,
        saved.MaintExpirationDate,
        saved.MaxProductVersion,
        saved.Features,
        "your-registration-secret");

    if (calculatedHash == saved.RegistrationHash)
    {
        // License is valid - check expiration dates and version
        if (DateTime.Parse(saved.ExpirationDate) >= DateTime.Today)
        {
            // Check version restriction
            if (LicenseClient.IsVersionAllowed(MyApp.Version, saved.MaxProductVersion))
            {
                // License valid and version allowed
            }
        }
    }
}
```

### Delphi

```pascal
uses
  activate, store;

var
  Response: TActivationResponse;
  SavedReg: TActivationResponse;
  CalculatedHash: string;
begin
  // Activate and validate in one step
  Response := ActivateProduct('https://license.example.com', LicenseKey, MachineCode, UserName);

  if ValidateRegistration(Response, 'your-registration-secret') then
  begin
    // Save response to local file for offline validation
    SaveRegistration(Response);
  end;

  // Later: Offline validation from saved registration file
  if LoadRegistration(SavedReg) then
  begin
    CalculatedHash := CalculateRegistrationHash(
      SavedReg.MachineCode,
      SavedReg.ExpirationDate,
      SavedReg.MaintExpirationDate,
      SavedReg.MaxProductVersion,
      'your-registration-secret',
      SavedReg.Features);

    if CalculatedHash = SavedReg.RegistrationHash then
    begin
      // License is valid - check expiration dates and version
      if StrToDate(SavedReg.ExpirationDate) >= Date then
      begin
        // Check version restriction
        if IsVersionAllowed(MyAppVersion, SavedReg.MaxProductVersion) then
        begin
          // License valid and version allowed
        end;
      end;
    end;
  end;
end;
```

### Go

```go
// Activate and validate in one step
response, err := ActivateProduct(baseURL, licenseKey, machineCode, userName)
if err != nil {
    log.Fatal(err)
}

if ValidateRegistration(response, "your-registration-secret") {
    // Save response to local file for offline validation
    if err := SaveRegistration(response); err != nil {
        log.Printf("Failed to save registration: %v", err)
    }
}

// Later: Offline validation from saved registration file
saved, err := LoadRegistration()
if err != nil {
    log.Fatal(err)
}

if saved != nil {
    calculatedHash := CalculateRegistrationHash(
        saved.MachineCode,
        saved.ExpirationDate,
        saved.MaintExpirationDate,
        saved.MaxProductVersion,
        "your-registration-secret",
        saved.Features,
    )

    if calculatedHash == saved.RegistrationHash {
        // License is valid - check expiration dates and version
        expDate, _ := time.Parse("2006-01-02", saved.ExpirationDate)
        if !expDate.Before(time.Now().Truncate(24 * time.Hour)) {
            // Check version restriction
            if IsVersionAllowed(myAppVersion, saved.MaxProductVersion) {
                // License valid and version allowed
            }
        }
    }
}
```

---

## Update Check Examples

### C#

```csharp
var client = new UpdateCheckClient("https://license.example.com");
var updateInfo = await client.CheckForUpdateAsync(licenseKey, machineCode, "1.0.0");

if (updateInfo.UpdateAvailable)
{
    Console.WriteLine($"Update available: {updateInfo.LatestVersion}");
    Console.WriteLine($"Download: {updateInfo.DownloadURL}");
}
```

### Delphi

```pascal
uses
  updatecheck;

var
  Info: TUpdateInfo;
begin
  Info := CheckForUpdate('https://license.example.com', LicenseKey, MachineCode, '1.0.0');

  if Info.UpdateAvailable then
  begin
    ShowMessage(Format('Update available: %s' + sLineBreak + 'Download: %s',
      [Info.LatestVersion, Info.DownloadURL]));
  end;
end;
```

### Go

```go
info, err := CheckForUpdate(baseURL, licenseKey, machineCode, "1.0.0")
if err != nil {
    log.Fatal(err)
}

if info.UpdateAvailable {
    fmt.Printf("Update available: %s\n", info.LatestVersion)
    fmt.Printf("Download: %s\n", info.DownloadURL)
}
```
