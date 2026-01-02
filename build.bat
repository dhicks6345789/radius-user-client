@echo off

set BUILDNAME=%1
if "%1"=="" (
  set BUILDNAME=test
)

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
set BUILDVERSION=%VERSION%-%BUILDNAME%-%CURRENTDATE%

if not exist "..\RADIUSClient" mkdir "..\RADIUSClient"

rem The Go YAML library.
go get -u gopkg.in/yaml.v3 2>&1
rem The Go RADIUS library.
go get -u layeh.com/radius 2>&1

rem Clear out previous builds.
erase ..\RADIUSClient\client.exe 2>&1
erase ..\RADIUSClient.zip 2>&1

echo Building version: %BUILDVERSION%...
go build -ldflags "-X main.buildVersion=%BUILDVERSION%" client.go 2>&1

if exist client.exe (
  echo Build succesful - creating Zip archive...
  move client.exe ..\RADIUSClient 2>&1
  tar -a -c -f ..\RADIUSClient.zip ..\RADIUSClient 2>&1
)
