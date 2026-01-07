// Usage:
// var client = new UpdateCheckClient("https://license.example.com");
// var updateInfo = await client.CheckForUpdateAsync(licenseKey, machineCode, "1.0.0");
// if (updateInfo.UpdateAvailable)
//     Console.WriteLine($"Update available: {updateInfo.LatestVersion} - {updateInfo.DownloadURL}");

using System.Net.Http.Json;

public record LicenseInfoRequest(string MachineCode, string InstalledVersion);

public record LicenseInfoResponse(
    string CustomerName,
    string ProductGUID,
    string ProductName,
    int LicenseCount,
    int LicensesAvailable,
    string ExpirationDate,
    string MaintExpirationDate,
    string MaxProductVersion,
    string LatestVersion,
    Dictionary<string, object>? Features
);

public record ProductVersionResponse(
    string ProductGUID,
    string LatestVersion,
    string DownloadURL
);

public record UpdateInfo(
    bool UpdateAvailable,
    string CurrentVersion,
    string LatestVersion,
    string? DownloadURL
);

public class UpdateCheckClient
{
    private readonly HttpClient _client;
    private readonly string _baseUrl;

    public UpdateCheckClient(string baseUrl)
    {
        _baseUrl = baseUrl;
        _client = new HttpClient();
    }

    /// <summary>
    /// Reports the installed version to the server and returns license info.
    /// </summary>
    public async Task<LicenseInfoResponse> UpdateInstalledVersionAsync(
        string licenseKey,
        string machineCode,
        string installedVersion)
    {
        var request = new LicenseInfoRequest(machineCode, installedVersion);

        var response = await _client.PutAsJsonAsync($"{_baseUrl}/api/v1/license/{licenseKey}", request);
        response.EnsureSuccessStatusCode();

        return await response.Content.ReadFromJsonAsync<LicenseInfoResponse>()
            ?? throw new InvalidOperationException("Empty response");
    }

    /// <summary>
    /// Retrieves the latest version and download URL for a product.
    /// </summary>
    public async Task<ProductVersionResponse> GetProductVersionAsync(string productGUID)
    {
        var response = await _client.GetAsync($"{_baseUrl}/api/v1/productver/{productGUID}");
        response.EnsureSuccessStatusCode();

        return await response.Content.ReadFromJsonAsync<ProductVersionResponse>()
            ?? throw new InvalidOperationException("Empty response");
    }

    /// <summary>
    /// Reports the installed version and checks if an update is available.
    /// Returns update info including the download URL if an update is available.
    /// </summary>
    public async Task<UpdateInfo> CheckForUpdateAsync(
        string licenseKey,
        string machineCode,
        string installedVersion)
    {
        // Step 1: Report installed version and get license info
        var licenseInfo = await UpdateInstalledVersionAsync(licenseKey, machineCode, installedVersion);

        // Step 2: Check if update is available
        if (LicenseClient.CompareVersions(installedVersion, licenseInfo.LatestVersion) >= 0)
        {
            return new UpdateInfo(
                UpdateAvailable: false,
                CurrentVersion: installedVersion,
                LatestVersion: licenseInfo.LatestVersion,
                DownloadURL: null
            );
        }

        // Step 3: Get download URL
        var productInfo = await GetProductVersionAsync(licenseInfo.ProductGUID);

        return new UpdateInfo(
            UpdateAvailable: true,
            CurrentVersion: installedVersion,
            LatestVersion: licenseInfo.LatestVersion,
            DownloadURL: productInfo.DownloadURL
        );
    }
}
