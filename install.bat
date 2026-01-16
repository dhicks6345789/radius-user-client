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

echo Stop / remove any existing running service.
net stop RADIUSUserClient 2>&1
nssm remove RADIUSUserClient confirm 2>&1

echo Make sure the install folder exists.
mkdir "C:\Program Files\RADIUSUserClient" 2>&1

echo Copy the executable and config file.
copy RADIUSUserClient-win-amd64.exe "C:\Program Files\RADIUSUserClient\RADIUSUserClient.exe" 2>&1
copy NSSM\2.24\win64\nssm.exe "C:\Program Files\RADIUSUserClient" 2>&1
copy config.txt "C:\Program Files\RADIUSUserClient" 2>&1

echo Set permissions on the config file so that only local admin accounts can read it.
icacls "C:\Program Files\RADIUSUserClient\config.txt" /reset
icacls "C:\Program Files\RADIUSUserClient\config.txt" /inheritance:r
icacls "C:\Program Files\RADIUSUserClient\config.txt" /remove:g * /T /C
icacls "C:\Program Files\RADIUSUserClient\config.txt" /grant Administrators:(F) /T /C

echo Setting up RADIUSUserClient as a Windows service...
"C:\Program Files\RADIUSUserClient\nssm.exe" install RADIUSUserClient "C:\Program Files\RADIUSUserClient\RADIUSUserClient.exe" 2>&1
"C:\Program Files\RADIUSUserClient\nssm.exe" set RADIUSUserClient DisplayName "RADIUS User Client" 2>&1
"C:\Program Files\RADIUSUserClient\nssm.exe" set RADIUSUserClient AppNoConsole 1 2>&1
"C:\Program Files\RADIUSUserClient\nssm.exe" set RADIUSUserClient Start SERVICE_AUTO_START 2>&1
net start RADIUSUserClient

rem Handy optional extra: if an HTTPS certificate (a .crt file) is found, install it as a Windows Trusted Root Certification Authorities certificate.
rem This allows you to distribute the Smoothwall (or other system's) trusted root certificate in one go along with the RADIUS client.
for %%f in (*.crt) do (
  echo Installing "%%f" in Trusted Root Certification Authorities...
  certutil -addstore -f "Root" "%%f"
)

echo Done.
