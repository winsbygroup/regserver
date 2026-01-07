unit activate;

interface

uses
  System.Net.HttpClient, System.JSON, System.SysUtils, System.Classes,
  System.Hash, System.NetEncoding, System.Generics.Collections,
  System.Generics.Defaults;

type
  TFeatures = TDictionary<string, string>;

  TActivationResponse = record
    UserName: string;
    UserCompany: string;
    ProductGUID: string;
    MachineCode: string;
    ExpirationDate: string;
    MaintExpirationDate: string;
    MaxProductVersion: string;
    LatestVersion: string;
    LicenseKey: string;
    RegistrationHash: string;
    Features: TFeatures;
  end;

  TVersionParts = record
    Major, Minor, Patch: Integer;
  end;

  function ActivateProduct(const BaseURL, LicenseKey, MachineCode, UserName: string): TActivationResponse;

  function CalculateRegistrationHash(const MachineCode, ExpirationDate, MaintExpirationDate,
      MaxProductVersion, Secret: string; Features: TFeatures): string;

  function ValidateRegistration(const Response: TActivationResponse; const Secret: string): Boolean;

  function IsVersionAllowed(const InstalledVersion, MaxProductVersion: string): Boolean;
  
implementation

function ActivateProduct(const BaseURL, LicenseKey, MachineCode, UserName: string): TActivationResponse;
var
  Response: IHTTPResponse;
begin
  var Client := THTTPClient.Create;
  var RequestBody := TJSONObject.Create;
  try
    RequestBody.AddPair('machineCode', MachineCode);
    RequestBody.AddPair('userName', UserName);

    var Content := TStringStream.Create(RequestBody.ToJSON, TEncoding.UTF8);
    try
      Client.CustomHeaders['X-License-Key'] := LicenseKey;
      Client.ContentType := 'application/json';

      Response := Client.Post(BaseURL + '/api/v1/activate', Content);

      if Response.StatusCode <> 200 then
        raise Exception.CreateFmt('Activation failed: %d %s', [Response.StatusCode, Response.StatusText]);

      var ResponseBody := TJSONObject.ParseJSONValue(Response.ContentAsString) as TJSONObject;
      try
        Result.UserName := ResponseBody.GetValue<string>('UserName');
        Result.UserCompany := ResponseBody.GetValue<string>('UserCompany');
        Result.ProductGUID := ResponseBody.GetValue<string>('ProductGUID');
        Result.MachineCode := ResponseBody.GetValue<string>('MachineCode');
        Result.ExpirationDate := ResponseBody.GetValue<string>('ExpirationDate');
        Result.MaintExpirationDate := ResponseBody.GetValue<string>('MaintExpirationDate');
        Result.MaxProductVersion := ResponseBody.GetValue<string>('MaxProductVersion');
        Result.LatestVersion := ResponseBody.GetValue<string>('LatestVersion');
        Result.LicenseKey := ResponseBody.GetValue<string>('LicenseKey');
        Result.RegistrationHash := ResponseBody.GetValue<string>('RegistrationHash');
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

function CalculateRegistrationHash(const MachineCode, ExpirationDate, MaintExpirationDate,
    MaxProductVersion, Secret: string; Features: TFeatures): string;
var
  UTF16Bytes: TBytes;
  HashBytes: TBytes;
begin
  // Step 1: Build the base registration string
  var RegString := MachineCode + '|' + ExpirationDate + '|' + MaintExpirationDate + '|' + MaxProductVersion;

  // Add features sorted alphabetically by key
  if Assigned(Features) and (Features.Count > 0) then
  begin
    var SortedKeys := TList<string>.Create;
    try
      for var key in Features.Keys do
        SortedKeys.Add(key);

      SortedKeys.Sort(TComparer<string>.Construct(
        function(const Left, Right: string): Integer
        begin
          Result := CompareStr(Left, Right);
        end));

      for var key in SortedKeys do
        RegString := RegString + '|' + Key + '=' + Features[Key];
    finally
      SortedKeys.Free;
    end;
  end;

  // Step 2: Append the secret
  RegString := RegString + Secret;

  // Step 3: Encode as UTF-16LE (no BOM)
  UTF16Bytes := TEncoding.Unicode.GetBytes(RegString);

  // Step 4: Compute SHA1 hash
  HashBytes := THashSHA1.GetHashBytes(UTF16Bytes);

  // Step 5: Base64 encode
  Result := TNetEncoding.Base64.EncodeBytesToString(HashBytes);
end;

function ValidateRegistration(const Response: TActivationResponse; const Secret: string): Boolean;
begin
  var CalculatedHash := CalculateRegistrationHash(
    Response.MachineCode,
    Response.ExpirationDate,
    Response.MaintExpirationDate,
    Response.MaxProductVersion,
    Secret,
    Response.Features);

  Result := CalculatedHash = Response.RegistrationHash;
end;

function ParseVersion(const S: string): TVersionParts;
var
  Parts: TArray<string>;
begin
  Parts := S.Split(['.']);
  Result.Major := StrToIntDef(Parts[0], 0);
  if Length(Parts) > 1 then
    Result.Minor := StrToIntDef(Parts[1], 0)
  else
    Result.Minor := 0;
  if Length(Parts) > 2 then
    Result.Patch := StrToIntDef(Parts[2], 0)
  else
    Result.Patch := 0;
end;

function CompareVersions(const A, B: string): Integer;
var
  V1, V2: TVersionParts;
begin
  V1 := ParseVersion(A);
  V2 := ParseVersion(B);

  if V1.Major <> V2.Major then
    Exit(V1.Major - V2.Major);

  if V1.Minor <> V2.Minor then
    Exit(V1.Minor - V2.Minor);

  Result := V1.Patch - V2.Patch;
end;

function IsVersionAllowed(const InstalledVersion, MaxProductVersion: string): Boolean;
begin
  if MaxProductVersion = '' then
    Result := True  // No restriction
  else
    Result := CompareVersions(InstalledVersion, MaxProductVersion) <= 0;
end;

end.
