@echo off

if not exist "..\RADIUSClient" mkdir "..\RADIUSClient"

echo Building...
go build client.go 2>&1
move client.exe ..\RADIUSClient
