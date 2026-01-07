param(
    [string]$DbFile,
    [string]$BackupDir
)

if (!(Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir | Out-Null
}

$bdate = (Get-Date).ToString("yyyy-MM-dd_HH.mm.ss")

sqlite3 $DbFile "vacuum into 'temp.db';"
sqlite3 temp.db .dump > "$BackupDir\$bdate`_regdump.sql"

Add-Content -Path "$BackupDir\$bdate`_regdump.sql" -Value "PRAGMA journal_mode=WAL"

Remove-Item temp.db
