@echo off

source VERSION
CURRENTDATE=`date +"%d/%m/%Y-%H:%M"`
BUILDVERSION="$VERSION-local-$CURRENTDATE"

if not exist "..\RADIUSClient" mkdir "..\RADIUSClient"

rem Module installation.
go mod download gopkg.in/yaml.v3

echo Building...
rem go build client.go 2>&1
go build -ldflags "-X main.buildVersion=$BUILDVERSION" client.go 2>&1
move client.exe ..\RADIUSClient
