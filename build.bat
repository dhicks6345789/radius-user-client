@echo off
setlocal EnableDelayedExpansion

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

mkdir "..\RADIUSUserClient" >nul 2>&1
mkdir "..\RADIUSUserClient\NSSM" >nul 2>&1
mkdir "..\RADIUSUserClient\NSSM\2.24" >nul 2>&1
mkdir "..\RADIUSUserClient\NSSM\2.24\win32" >nul 2>&1
mkdir "..\RADIUSUserClient\NSSM\2.24\win64" >nul 2>&1

rem The Go YAML library.
go get -u gopkg.in/yaml.v3 2>&1
rem The Go RADIUS library.
go get -u layeh.com/radius 2>&1

rem Clear out previous builds.
erase ..\RADIUSUserClient\*.exe >nul 2>&1
erase ..\*.zip >nul 2>&1
if "%BUILDNAME%"=="main" (
  erase ..\RADIUSUserClient\config.txt >nul 2>&1
)

rem Copy over an example config file, but only if the user hasn't already provided their own.
if not exist ..\RADIUSUserClient\config.txt (
  echo Existing config file not found - copying default example version over.
  copy config-example.txt ..\RADIUSUserClient\config.txt
)

rem Copy over the install scripts, including NSSM for Windows Service creation.
copy install.bat ..\RADIUSUserClient >nul 2>&1
copy NSSM\2.24\win32\nssm.exe ..\RADIUSUserClient\NSSM\2.24\win32 >nul 2>&1
copy NSSM\2.24\win64\nssm.exe ..\RADIUSUserClient\NSSM\2.24\win64 >nul 2>&1

echo Building version: %BUILDVERSION%...
go build -ldflags "-X main.buildVersion=%BUILDVERSION%" client.go 2>&1

if exist client.exe (
  echo Build succesful - creating Zip archive...
  move client.exe ..\RADIUSUserClient\RADIUSUserClient-win64.exe >nul 2>&1
  set ZIPNAME=RADIUSUserClient-!BUILDNAME!
  if "%BUILDNAME%"=="main" (
    set ZIPNAME=RADIUSUserClient
  )
  tar -a -c -f "..\!ZIPNAME!.zip" ..\RADIUSUserClient >nul 2>&1
)
