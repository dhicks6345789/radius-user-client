echo Starting install...

# Work out what architecture we are installing on.
ARCH=$(uname -m)
BINARY=linux-amd64
[[ $ARCH == arm* ]] && BINARY=linux-arm32
[[ $ARCH == aarch64 ]] && BINARY=linux-arm64

if [ "$ARCH" = "arm64" ]; then
  echo "Doing MacOS install..."
  cp RADIUSUserClient-mac-arm64 /usr/local/bin/RADIUSUserClient
  mkdir "/Library/Application Support/RADIUSUserClient" > /dev/null 2>&1
  cp config.txt "/Library/Application Support/RADIUSUserClient"
else
  echo "Doing Linux install..."
  cp RADIUSUserClient-lin-amd64 /usr/local/bin/RADIUSUserClient
fi
chmod u+x /usr/local/bin/RADIUSUserClient

echo Done.
