echo Starting install...

# Work out what architecture we are installing on.
ARCH=$(uname -m)
BINARY=linux-amd64
[[ $ARCH == arm* ]] && BINARY=linux-arm32
[[ $ARCH == aarch64 ]] && BINARY=linux-arm64

if [ "$ARCH" = "arm64" ]; then
  echo "On MacOS!"
else
  # Copy the executable to the system.
  # cp RADIUSUserClient-lin-amd64 /usr/local/bin/RADIUSUserClient
  # chmod u+x /usr/local/bin/RADIUSUserClient
fi
