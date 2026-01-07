unit updatecheck;

interface

uses
  System.Net.HttpClient, System.JSON, System.SysUtils, System.Classes;

type
  TLicenseInfoResponse = record
    CustomerName: string;
    ProductGUID: string;
    ProductName: string;
    LicenseCount: Integer;
    LicensesAvailable: Integer;
    ExpirationDate: string;
    MaintExpirationDate: string;
    MaxProductVersion: string;
    LatestVersion: string;
  end;

  TProductVersionResponse = record
    ProductGUID: string;
    LatestVersion: string;
    DownloadURL: string;
  end;

  TUpdateInfo = record
    UpdateAvailable: Boolean;
    CurrentVersion: string;
    LatestVersion: string;
    DownloadURL: string;
  end;

  function UpdateInstalledVersion(const BaseURL, LicenseKey, MachineCode, InstalledVersion: string): TLicenseInfoResponse;
  function GetProductVersion(const BaseURL, ProductGUID: string): TProductVersionResponse;
  function CheckForUpdate(const BaseURL, LicenseKey, MachineCode, InstalledVersion: string): TUpdateInfo;

implementation

uses
  activate; // For CompareVersions

function UpdateInstalledVersion(const BaseURL, LicenseKey, MachineCode, InstalledVersion: string): TLicenseInfoResponse;
var
  Response: IHTTPResponse;
begin
  var Client := THTTPClient.Create;
  var RequestBody := TJSONObject.Create;
  try
    RequestBody.AddPair('machineCode', MachineCode);
    RequestBody.AddPair('installedVersion', InstalledVersion);

    var Content := TStringStream.Create(RequestBody.ToJSON, TEncoding.UTF8);
    try
      Client.ContentType := 'application/json';

      Response := Client.Put(BaseURL + '/api/v1/license/' + LicenseKey, Content);

      if Response.StatusCode <> 200 then
        raise Exception.CreateFmt('Update failed: %d %s', [Response.StatusCode, Response.StatusText]);

      var ResponseBody := TJSONObject.ParseJSONValue(Response.ContentAsString) as TJSONObject;
      try
        Result.CustomerName := ResponseBody.GetValue<string>('CustomerName');
        Result.ProductGUID := ResponseBody.GetValue<string>('ProductGUID');
        Result.ProductName := ResponseBody.GetValue<string>('ProductName');
        Result.LicenseCount := ResponseBody.GetValue<Integer>('LicenseCount');
        Result.LicensesAvailable := ResponseBody.GetValue<Integer>('LicensesAvailable');
        Result.ExpirationDate := ResponseBody.GetValue<string>('ExpirationDate');
        Result.MaintExpirationDate := ResponseBody.GetValue<string>('MaintExpirationDate');
        Result.MaxProductVersion := ResponseBody.GetValue<string>('MaxProductVersion');
        Result.LatestVersion := ResponseBody.GetValue<string>('LatestVersion');
      finally
        ResponseBody.Free;
      end;
    finally
      Content.Free;
    end;
  finally
    RequestBody.Free;
    Client.Free;
  end;
end;

function GetProductVersion(const BaseURL, ProductGUID: string): TProductVersionResponse;
var
  Response: IHTTPResponse;
begin
  var Client := THTTPClient.Create;
  try
    Response := Client.Get(BaseURL + '/api/v1/productver/' + ProductGUID);

    if Response.StatusCode <> 200 then
      raise Exception.CreateFmt('Get product version failed: %d %s', [Response.StatusCode, Response.StatusText]);

    var ResponseBody := TJSONObject.ParseJSONValue(Response.ContentAsString) as TJSONObject;
    try
      Result.ProductGUID := ResponseBody.GetValue<string>('ProductGUID');
      Result.LatestVersion := ResponseBody.GetValue<string>('LatestVersion');
      Result.DownloadURL := ResponseBody.GetValue<string>('DownloadURL');
    finally
      ResponseBody.Free;
    end;
  finally
    Client.Free;
  end;
end;

function CheckForUpdate(const BaseURL, LicenseKey, MachineCode, InstalledVersion: string): TUpdateInfo;
var
  LicenseInfo: TLicenseInfoResponse;
  ProductInfo: TProductVersionResponse;
begin
  // Step 1: Report installed version and get license info
  LicenseInfo := UpdateInstalledVersion(BaseURL, LicenseKey, MachineCode, InstalledVersion);

  Result.CurrentVersion := InstalledVersion;
  Result.LatestVersion := LicenseInfo.LatestVersion;
  Result.UpdateAvailable := False;
  Result.DownloadURL := '';

  // Step 2: Check if update is available
  if CompareVersions(InstalledVersion, LicenseInfo.LatestVersion) >= 0 then
    Exit;

  // Step 3: Get download URL
  ProductInfo := GetProductVersion(BaseURL, LicenseInfo.ProductGUID);

  Result.UpdateAvailable := True;
  Result.DownloadURL := ProductInfo.DownloadURL;
end;

end.
