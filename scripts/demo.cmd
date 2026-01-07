@echo off
del /q testdata\demo.db 2>nul
set ADMIN_API_KEY=demo
set REGISTRATION_SECRET=demo-secret
set DB_PATH=./testdata/demo.db
call dist\regserver.exe -demo && exit 0 || exit 1
