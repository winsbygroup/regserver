unit machine;

interface

// note: this routine uses com and so coinitialize must have been called beforehand.

uses
  System.SysUtils,
  System.Hash,
  System.NetEncoding,
  Winapi.ActiveX,
  System.Win.ComObj;
  
function GetMachineCode: string;  
  
implementation

function WmiQuery(const WmiClass, WmiProperty: string; const Where: string = ''): string;
var
  Locator: OleVariant;
  Services: OleVariant;
  ObjSet: OleVariant;
  Obj: OleVariant;
begin
  Result := '';
  try
    Locator := CreateOleObject('WbemScripting.SWbemLocator');
    Services := Locator.ConnectServer('.', 'root\cimv2');

    var Query := Format('SELECT %s FROM %s', [WmiProperty, WmiClass]);
    if Where <> '' then
      Query := Query + ' WHERE ' + Where;

    ObjSet := Services.ExecQuery(Query);
    for Obj in ObjSet do
    begin
      Result := VarToStr(Obj.Properties_.Item(WmiProperty).Value);
      Exit;
    end;
  except
    // swallow and return empty
  end;
end;

function GetMachineCode: string;
var
  Raw: string;
  HashBytes: TBytes;
  HashBase64: string;
begin
  Raw := WmiQuery('Win32_BaseBoard', 'SerialNumber') + '|' +
         WmiQuery('Win32_Processor', 'ProcessorId') + '|' +
         WmiQuery('Win32_OperatingSystem', 'SerialNumber') + '|' +
         WmiQuery('Win32_LogicalDisk', 'VolumeSerialNumber', 'DeviceID="C:""');

  HashBytes := THashSHA2.GetHashBytes(Raw);
  HashBase64 := TNetEncoding.Base64.EncodeBytesToString(HashBytes);

  // sanitize for URLs, DB keys, etc.
  HashBase64 := HashBase64.Replace('=', '').Replace('+', '').Replace('/', '');

  Result := HashBase64;
end;

end.
