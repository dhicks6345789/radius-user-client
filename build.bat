@echo off

echo Checking libraries are installed...
go get gopkg.in/yaml.v3

if not exist "..\RADIUSClient" mkdir "..\RADIUSClient"

echo Building...
go build client.go 2>&1
move client.exe ..\RADIUSClient
