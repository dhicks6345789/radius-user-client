# RADIUS User Client

A cross-platform client application to tell a RADIUS server the current user on a machine.

## What This Application Is For

This is a small client-side application, intended to run on workstations in a managed network environment (maybe a school or corporate office setup), where those workstations might be used by multiple users during a typical day. It aims to periodically update a central RADIUS server with the username of the current user using a device.

This application is intended for use by managers of corporate networks with devices shared between users. It probably isn't going to be much use for personal devices, and will more probably be rolled at as part of a managed network's software deployment process. It is open source, so you are free to examine and tinker with the source code if you want, which might prove useful to some people.

## Why This Application Was Written

This application was written to replace a vendor's ([Smoothwall](https://smoothwall.com/)) solution (their [Cloud Filter Unified Client](https://software.smoothwall.com/)) in a particular situation where it didn't support the login mechanism being used by the managed Windows devices on a particular network. Smoothwall's on-premesis network traffic filtering applience provides good support for RADIUS clients, enabling users with particular requirements to implement their own solutions. This application is intended to be usable with Smoothwall on-premesis appliences via RADIUS, it might also be useful for other vendors' similar products, other network products that utilise RADIUS in some way, or even just as a starting point for someone trying to write their own RADIUS client.

## Suitibility

This application is not affiliated with or endorsed in any way by Smoothwall or any other vendor, and no garuntee is made that it will function as intended on any given hardware or network configuration. It is written as quite a simple, single-file Go application, and should, hopefully, be auditable by internal deployment teams if needed.

## Platforms

This application is written in Go, and should be able to be compiled and run on most current platforms, including Windows, MacOS and Linux running on a variety of architectures (x86, amd64, arm64).

The application is distributed as a simple, single Zip archive contaning all the executables and install scripts for each of the platforms supported.

## Installation

Unpack the Zip file and run the install script for your platform (install.bat for Windows, install.sh for other platforms).

## Building

If you want to build the application from the source code, build scripts (`build.bat`, `build.sh`) are included in the Git repository for both Windows and Linux. The whole build process is basically `git build client.go`.

If you are building the client on a test machine, there are also included test scripts (`test.bat`, `test.sh`) which build the code then do a local deploy of the executables.
