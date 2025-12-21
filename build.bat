@echo off

echo Checking libraries are installed...
rem go get github.com/nfnt/resize

if not exist "..\RADIUSClient" mkdir "..\RADIUSClient"

echo Building...
go build client.go
move client.exe ..\RADIUSClient

rem copy webconsole.exe "C:\Program Files\WebConsole" > nul 2>&1
rem xcopy /E /Y www "C:\Program Files\WebConsole\www" > nul 2>&1
rem xcopy /E /Y ..\ace-builds\src-noconflict "C:\Program Files\WebConsole\www\ace" > nul 2>&1
rem net start WebConsole > nul 2>&1
