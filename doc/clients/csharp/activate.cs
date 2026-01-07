// Usage:
// var client = new LicenseClient("https://license.example.com", "your-license-key");
// var result = await client.ActivateAsync(machineCode, Environment.UserName);
// var isValid = client.ValidateRegistration(result, "your-secret");

using System.Net.Http.Json;
using System.Security.Cryptography;
using System.Text;

public record ActivationRequest(string MachineCode, string UserName);

public record ActivationResponse(
    string UserName,
    string UserCompany,
    string ProductGUID,
    string MachineCode,
    string ExpirationDate,
    string MaintExpirationDate,
    string MaxProductVersion,
    string LatestVersion,
    string LicenseKey,
    string RegistrationHash,
    Dictionary<string, object>? Features
);

public class LicenseClient
{
    private readonly HttpClient _client;
    private readonly string _baseUrl;

    public LicenseClient(string baseUrl, string licenseKey)
    {
        _baseUrl = baseUrl;
        _client = new HttpClient();
        _client.DefaultRequestHeaders.Add("X-License-Key", licenseKey);
    }

    public async Task<ActivationResponse> ActivateAsync(string machineCode, string userName)
    {
        var request = new ActivationRequest(machineCode, userName);

        var response = await _client.PostAsJsonAsync($"{_baseUrl}/api/v1/activate", request);
        response.EnsureSuccessStatusCode();

        return await response.Content.ReadFromJsonAsync<ActivationResponse>()
            ?? throw new InvalidOperationException("Empty response");
    }

    /// <summary>
    /// Calculates the registration hash for offline license validation.
    /// </summary>
    /// <param name="machineCode">The machine code</param>
    /// <param name="expirationDate">Expiration date in yyyy-mm-dd format</param>
    /// <param name="maintExpirationDate">Maintenance expiration date in yyyy-mm-dd format</param>
    /// <param name="maxProductVersion">Maximum allowed product version (empty if no restriction)</param>
    /// <param name="features">Feature dictionary (values will be converted to strings)</param>
    /// <param name="secret">The REGISTRATION_SECRET shared with the server</param>
    /// <returns>Base64-encoded SHA1 hash</returns>
    public static string CalculateRegistrationHash(
        string machineCode,
        string expirationDate,
        string maintExpirationDate,
        string maxProductVersion,
        Dictionary<string, object>? features,
        string secret)
    {
        // Step 1: Build the registration string
        var sb = new StringBuilder();
        sb.Append(machineCode);
        sb.Append('|');
        sb.Append(expirationDate);
        sb.Append('|');
        sb.Append(maintExpirationDate);
        sb.Append('|');
        sb.Append(maxProductVersion ?? "");

        // Add features sorted alphabetically by key
        if (features != null && features.Count > 0)
        {
            var sortedKeys = features.Keys.OrderBy(k => k, StringComparer.Ordinal).ToList();
            foreach (var key in sortedKeys)
            {
                sb.Append('|');
                sb.Append(key);
                sb.Append('=');
                sb.Append(features[key]?.ToString() ?? "");
            }
        }

        // Step 2: Append the secret
        sb.Append(secret);

        // Step 3: Encode as UTF-16LE (no BOM)
        var utf16Bytes = Encoding.Unicode.GetBytes(sb.ToString());

        // Step 4: Compute SHA1 hash
        var hashBytes = SHA1.HashData(utf16Bytes);

        // Step 5: Base64 encode
        return Convert.ToBase64String(hashBytes);
    }

    /// <summary>
    /// Validates an activation response by comparing the calculated hash with the server-provided hash.
    /// </summary>
    public static bool ValidateRegistration(ActivationResponse response, string secret)
    {
        var calculatedHash = CalculateRegistrationHash(
            response.MachineCode,
            response.ExpirationDate,
            response.MaintExpirationDate,
            response.MaxProductVersion,
            response.Features,
            secret);

        return calculatedHash == response.RegistrationHash;
    }

    /// <summary>
    /// Checks if the installed product version is allowed based on MaxProductVersion.
    /// </summary>
    /// <param name="installedVersion">The version of the running software</param>
    /// <param name="maxProductVersion">The maximum allowed version from the license (empty = no restriction)</param>
    /// <returns>True if the version is allowed, false if restricted</returns>
    public static bool IsVersionAllowed(string installedVersion, string maxProductVersion)
    {
        if (string.IsNullOrEmpty(maxProductVersion))
            return true; // No restriction

        return CompareVersions(installedVersion, maxProductVersion) <= 0;
    }

    /// <summary>
    /// Compares two semver version strings (e.g., "1.2.3").
    /// </summary>
    /// <returns>Negative if a &lt; b, zero if a == b, positive if a &gt; b</returns>
    public static int CompareVersions(string a, string b)
    {
        var aParts = ParseVersion(a);
        var bParts = ParseVersion(b);

        if (aParts.Major != bParts.Major)
            return aParts.Major - bParts.Major;

        if (aParts.Minor != bParts.Minor)
            return aParts.Minor - bParts.Minor;

        return aParts.Patch - bParts.Patch;
    }

    private static (int Major, int Minor, int Patch) ParseVersion(string version)
    {
        var parts = (version ?? "").Split('.');
        return (
            parts.Length > 0 && int.TryParse(parts[0], out var major) ? major : 0,
            parts.Length > 1 && int.TryParse(parts[1], out var minor) ? minor : 0,
            parts.Length > 2 && int.TryParse(parts[2], out var patch) ? patch : 0
        );
    }
}

