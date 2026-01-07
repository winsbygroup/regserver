// Registration file storage for offline license validation.
// Stores registration data as JSON in a well-known location.

using System.Text.Json;

public static class RegistrationStore
{
    private const string CompanyName = "Company";
    private const string ProductName = "Product";
    private const string FileName = "product.json";

    /// <summary>
    /// Gets the directory where registration files are stored.
    /// Creates the directory if it doesn't exist.
    /// </summary>
    public static string GetStorageDirectory()
    {
        var path = Path.Combine(
            Environment.GetFolderPath(Environment.SpecialFolder.CommonApplicationData),
            CompanyName,
            ProductName);

        if (!Directory.Exists(path))
        {
            Directory.CreateDirectory(path);
        }

        return path;
    }

    /// <summary>
    /// Gets the full path to the registration file.
    /// </summary>
    public static string GetRegistrationFilePath()
    {
        return Path.Combine(GetStorageDirectory(), FileName);
    }

    /// <summary>
    /// Saves an activation response to the registration file.
    /// </summary>
    public static void SaveRegistration(ActivationResponse response)
    {
        var json = JsonSerializer.Serialize(response, new JsonSerializerOptions
        {
            WriteIndented = true
        });

        File.WriteAllText(GetRegistrationFilePath(), json);
    }

    /// <summary>
    /// Loads a previously saved registration from the registration file.
    /// Returns null if no registration file exists.
    /// </summary>
    public static ActivationResponse? LoadRegistration()
    {
        var path = GetRegistrationFilePath();

        if (!File.Exists(path))
        {
            return null;
        }

        var json = File.ReadAllText(path);
        return JsonSerializer.Deserialize<ActivationResponse>(json);
    }

    /// <summary>
    /// Checks if a registration file exists.
    /// </summary>
    public static bool RegistrationExists()
    {
        return File.Exists(GetRegistrationFilePath());
    }

    /// <summary>
    /// Deletes the registration file if it exists.
    /// </summary>
    public static void DeleteRegistration()
    {
        var path = GetRegistrationFilePath();

        if (File.Exists(path))
        {
            File.Delete(path);
        }
    }
}
