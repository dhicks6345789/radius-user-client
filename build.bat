@echo off

set /p VERSION=<VERSION
echo The VERSION is: %VERSION%


rem CURRENTDATE=`date +"%d/%m/%Y-%H:%M"`
for /f "tokens=2 delims==" %%a in ('wmic os get localdatetime /value') do set dt=%%a
echo %dt%
set YYYY=%dt:~0,4%
set MM=%dt:~4,2%
set DD=%dt:~6,2%

set CURRENTDATE=%YYYY%-%MM%-%DD%
echo The current date is: %CURRENTDATE%


rem BUILDVERSION="$VERSION-local-$CURRENTDATE"

if not exist "..\RADIUSClient" mkdir "..\RADIUSClient"

rem Module installation.
go mod download gopkg.in/yaml.v3 2>&1

echo Building...
rem go build client.go 2>&1
rem go build -ldflags "-X main.buildVersion=$BUILDVERSION" client.go 2>&1
erase ..\RADIUSClient\client.exe 2>&1
move client.exe ..\RADIUSClient 2>&1
