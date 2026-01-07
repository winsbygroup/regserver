using System;
using System.Linq;
using System.Management;
using System.Security.Cryptography;
using System.Text;

// Usage: string machineCode = MachineCode.GetMachineCode();

public static class MachineCode
{
    public static string GetMachineCode()
    {
        var raw = string.Join("|", new[]
        {
            GetWmi("Win32_BaseBoard", "SerialNumber"),
            GetWmi("Win32_Processor", "ProcessorId"),
            GetWmi("Win32_OperatingSystem", "SerialNumber"),
            GetWmi("Win32_LogicalDisk", "VolumeSerialNumber", "DeviceID='C:'")
        }.Where(s => !string.IsNullOrWhiteSpace(s)));

        return Hash(raw);
    }

    private static string GetWmi(string className, string propertyName, string whereClause = null)
    {
        try
        {
            string query = $"SELECT {propertyName} FROM {className}";
            if (!string.IsNullOrEmpty(whereClause))
                query += $" WHERE {whereClause}";

            using var searcher = new ManagementObjectSearcher(query);
            foreach (var obj in searcher.Get())
            {
                return obj[propertyName]?.ToString()?.Trim();
            }
        }
        catch
        {
            // swallow and return null
        }
        return null;
    }

    private static string Hash(string input)
    {
        using var sha = SHA256.Create();
        var bytes = sha.ComputeHash(Encoding.UTF8.GetBytes(input));
        return Convert.ToBase64String(bytes)
            .Replace("=", "")
            .Replace("/", "")
            .Replace("+", "");
    }
}