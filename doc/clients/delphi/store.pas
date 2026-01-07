unit store;

// Registration file storage for offline license validation.
// Stores registration data as JSON in a well-known location.

interface

uses
  System.SysUtils, System.IOUtils, System.JSON, System.Generics.Collections,
  activate;

const
  CompanyName = 'Company';
  ProductName = 'Product';
  FileName = 'product.json';

function GetStorageDirectory: string;
function GetRegistrationFilePath: string;
procedure SaveRegistration(const Response: TActivationResponse);
function LoadRegistration(out Response: TActivationResponse): Boolean;
function RegistrationExists: Boolean;
procedure DeleteRegistration;

implementation

function GetStorageDirectory: string;
begin
  Result := TPath.Combine(TPath.GetPublicPath, CompanyName);
  Result := TPath.Combine(Result, ProductName);

  if not TDirectory.Exists(Result) then
    TDirectory.CreateDirectory(Result);
end;

function GetRegistrationFilePath: string;
begin
  Result := TPath.Combine(GetStorageDirectory, FileName);
end;

procedure SaveRegistration(const Response: TActivationResponse);
var
  JSON: TJSONObject;
  FeaturesJSON: TJSONObject;
  Pair: TPair<string, string>;
begin
  JSON := TJSONObject.Create;
  try
    JSON.AddPair('UserName', Response.UserName);
    JSON.AddPair('UserCompany', Response.UserCompany);
    JSON.AddPair('ProductGUID', Response.ProductGUID);
    JSON.AddPair('MachineCode', Response.MachineCode);
    JSON.AddPair('ExpirationDate', Response.ExpirationDate);
    JSON.AddPair('MaintExpirationDate', Response.MaintExpirationDate);
    JSON.AddPair('CurrentVersion', Response.CurrentVersion);
    JSON.AddPair('LicenseKey', Response.LicenseKey);
    JSON.AddPair('RegistrationHash', Response.RegistrationHash);

    if Assigned(Response.Features) and (Response.Features.Count > 0) then
    begin
      FeaturesJSON := TJSONObject.Create;
      for Pair in Response.Features do
        FeaturesJSON.AddPair(Pair.Key, Pair.Value);
      JSON.AddPair('Features', FeaturesJSON);
    end;

    TFile.WriteAllText(GetRegistrationFilePath, JSON.Format);
  finally
    JSON.Free;
  end;
end;

function LoadRegistration(out Response: TActivationResponse): Boolean;
var
  JSON: TJSONObject;
  FeaturesJSON: TJSONObject;
  Pair: TJSONPair;
  FilePath: string;
begin
  Result := False;
  FilePath := GetRegistrationFilePath;

  if not TFile.Exists(FilePath) then
    Exit;

  JSON := TJSONObject.ParseJSONValue(TFile.ReadAllText(FilePath)) as TJSONObject;
  if not Assigned(JSON) then
    Exit;

  try
    Response.UserName := JSON.GetValue<string>('UserName', '');
    Response.UserCompany := JSON.GetValue<string>('UserCompany', '');
    Response.ProductGUID := JSON.GetValue<string>('ProductGUID', '');
    Response.MachineCode := JSON.GetValue<string>('MachineCode', '');
    Response.ExpirationDate := JSON.GetValue<string>('ExpirationDate', '');
    Response.MaintExpirationDate := JSON.GetValue<string>('MaintExpirationDate', '');
    Response.CurrentVersion := JSON.GetValue<string>('CurrentVersion', '');
    Response.LicenseKey := JSON.GetValue<string>('LicenseKey', '');
    Response.RegistrationHash := JSON.GetValue<string>('RegistrationHash', '');

    Response.Features := TFeatures.Create;
    if JSON.TryGetValue<TJSONObject>('Features', FeaturesJSON) then
    begin
      for Pair in FeaturesJSON do
        Response.Features.Add(Pair.JsonString.Value, Pair.JsonValue.Value);
    end;

    Result := True;
  finally
    JSON.Free;
  end;
end;

function RegistrationExists: Boolean;
begin
  Result := TFile.Exists(GetRegistrationFilePath);
end;

procedure DeleteRegistration;
var
  FilePath: string;
begin
  FilePath := GetRegistrationFilePath;

  if TFile.Exists(FilePath) then
    TFile.Delete(FilePath);
end;

end.
