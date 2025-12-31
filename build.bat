@echo off

rem source VERSION
rem CURRENTDATE=`date +"%d/%m/%Y-%H:%M"`
rem BUILDVERSION="$VERSION-local-$CURRENTDATE"

for /f "tokens=2 delims==" %%a in ('findstr "version" VERSION') do set VERSION=%%a
echo The VERSION is: %VERSION%

if not exist "..\RADIUSClient" mkdir "..\RADIUSClient"

rem Module installation.
go mod download gopkg.in/yaml.v3

echo Building...
rem go build client.go 2>&1
go build -ldflags "-X main.buildVersion=$BUILDVERSION" client.go 2>&1
move client.exe ..\RADIUSClient
