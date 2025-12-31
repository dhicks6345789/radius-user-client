@echo off

rem Get the version number from the repository...
set /p VERSION=<VERSION
rem ...figure out the current date and time...
for /f "tokens=2 delims==" %%a in ('wmic os get localdatetime /value') do set dt=%%a
set YYYY=%dt:~0,4%
set MM=%dt:~4,2%
set DD=%dt:~6,2%
set HH=%dt:~8,2%
set NN=%dt:~10,2%
set CURRENTDATE=%DD%/%MM%/%YYYY%-%HH%:%NN%
rem ... and combine those to give a build version value.
set BUILDVERSION=%VERSION%-local-%CURRENTDATE%

if not exist "..\RADIUSClient" mkdir "..\RADIUSClient"

rem Module installation.
go mod download gopkg.in/yaml.v3 2>&1
go mod download layeh.com/radius 2>&1

echo Building version: %BUILDVERSION%...
rem go build client.go 2>&1
go build -ldflags "-X main.buildVersion=%BUILDVERSION%" client.go 2>&1
erase ..\RADIUSClient\client.exe 2>&1
move client.exe ..\RADIUSClient 2>&1
