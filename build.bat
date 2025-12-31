@echo off

if not exist "..\RADIUSClient" mkdir "..\RADIUSClient"

rem Module installation.
go mod download gopkg.in/yaml.v3

echo Building...
go build client.go 2>&1
move client.exe ..\RADIUSClient
