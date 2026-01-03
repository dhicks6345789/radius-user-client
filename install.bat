@echo off
echo Starting install...

rem Parse any parameters.
:paramLoop
if "%1"=="" goto paramContinue
if "%1"=="--key" (
  shift
  set key=%2
)
if "%1"=="--subdomain" (
  shift
  set subdomain=%2
)
shift
goto paramLoop
:paramContinue

rem Stop any existing running service.
net stop RADIUSUserClient > nul 2>&1

rem Make sure the install folder exists.
mkdir "C:\Program Files\RADIUSUserClient" > nul 2>&1

rem Copy the executable and config file.
copy client.exe "C:\Program Files\RADIUSUserClient" > nul 2>&1
copy config.txt "C:\Program Files\RADIUSUserClient" > nul 2>&1

rem Set permissions on the config file so that only local admin accounts can read it.
icacls "C:\Program Files\RADIUSUserClient\config.txt" /reset
icacls "C:\Program Files\RADIUSUserClient\config.txt" /inheritance:r
icacls "C:\Program Files\RADIUSUserClient\config.txt" /remove:g * /T /C
icacls "C:\Program Files\RADIUSUserClient\config.txt" /grant Administrators:(F) /T /C

echo Setting up RADIUSUserClient as a Windows service...
nssm install RADIUSUserClient "C:\Program Files\RADIUSUserClient\.bat" > nul 2>&1
nssm set RADIUSUserClient DisplayName "RADIUS User Client" > nul 2>&1
nssm set RADIUSUserClient AppNoConsole 1 > nul 2>&1
nssm set RADIUSUserClient Start SERVICE_AUTO_START > nul 2>&1
net start RADIUSUserClient

echo Done.
