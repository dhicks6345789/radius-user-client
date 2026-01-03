# RADIUS User Client
A cross-platform client application to tell a RADIUS server the current user on a machine.

Note: as of 3rd January 2026, this project is not completed or production-ready, although it should compile and be usable as a (Windows-only, currently) command-line utility to send a basic accounting packet to your Smoothwall server.

## What This Application Is For
This is a small client-side application, intended to run on workstations in a managed network environment (maybe a school or corporate office setup), where those workstations might be used by multiple users during a typical day. It aims to periodically update a central RADIUS server with the username of the current user using a device.

This application is intended for use by managers of corporate networks with devices shared between users. It probably isn't going to be much use for personal devices, and will more probably be rolled at as part of a managed network's software deployment process. It is open source, so you are free to examine and tinker with the source code if you want, which might prove useful to some people.

## Why This Application Was Written
This application was written to replace a vendor's ([Smoothwall](https://smoothwall.com/)) solution (their [Cloud Filter Unified Client](https://software.smoothwall.com/)) in a particular situation where it didn't support the login mechanism being used by the managed Windows devices on a particular network. Smoothwall's on-premesis network traffic filtering applience provides good support for RADIUS clients, enabling users with particular requirements to implement their own solutions.

This application is intended to be usable with Smoothwall on-premesis appliences via RADIUS, it might also be useful for other vendors' similar products, other network products that utilise RADIUS in some way, or even just as a starting point for someone trying to write their own RADIUS client.

## Suitibility
This application is not affiliated with or endorsed in any way by Smoothwall or any other vendor, and no garuntee is made that it will function as intended on any given hardware or network configuration. It is written as quite a simple, single-file Go application, and should, hopefully, be of a suitible size and lack of complexity to be easily auditable by internal deployment teams if needed.

## Platforms
This application is written in Go, and should be able to be compiled and run on most current platforms, including Windows, MacOS and Linux running on a variety of architectures (x86, amd64, arm64).

The mechanisms used to get the current user and IP address on the client-side vary for each platform, so while the executable might compile and run on a given platform it might not actually be able to get the username and/or IP address properly. If that is so, it should still be usable as a command-line utility where you can specify the username and IP address as parameters yourself.

The application is distributed as a simple, single Zip archive contaning all the executables and install scripts for each of the platforms supported, as well as default/example config files.

Written in Go, the executable should be self-contained, with no external requirements (runtime libraries and so on), other than its own (very basic) config file. As such, it should be able to be used on minimal VMs designed for cloud computing.

## Installation
To install on a single device, unpack the Zip file and run the install script for your platform (install.bat for Windows, install.sh for other platforms). On a managed network, you are probably going to want to unpack that Zip file into a central repository of some sort, modify the given configuration file to match your setup, then distribute from there.

### Configuration - Client Devices
On your Windows / MacOS / etc devices, you will need to have a config file holding appropriate values for your network. You'll need the IP address of the RADIUS server, the shared secret provided by your RADIUS server, and the domain name you wish to append to the usernames.

### Configuration - RADIUS Server
If you are using a Smoothwall appliance, you can configure the appliance itself to accept RADIUS accounting packets. This is quite simple, it should just be a simple check-box to click. Don't forget to also add the appropriate RADIUS port (probably 1813 for accounting) to the allow list in your Smoothwall's firewall configuration section.

You will need to define a secret shared key on your RADIUS server for use by the client software. It is suggested you create a separate key for use by this application if possible, rather than re-using an existing one, just to keep some separation between applications.

This application hasn't been tested with other RADIUS servers or filtering appliances, any feedback would be useful.

## Usage Examples
### Windows Devices Using GCPW / pGina As a Login Provider
For the particular situation this application was written for: We have a network of Windows devices, in our case running Windows 11 Education (similar to Windows 11 Enterprise for corporpate customers). These are devices in a school setting, mostly classroom machines used by teaching staff. As a school network we have a gateway-level filtering setup, a Smoothwall appliance, that can filter traffic to various levels and issue alerts in line with current safeguarding guidence. As part of that filtering process, it is important for us to understand which user has triggered a filtering block or alert. Therefore when a user logs in to a machine, that machine needs to tell the Smoothwall server which user has just logged in. To this end, Smoothwall produce a Windows client application that runs as a service and communicates the current user to the Smoothwall server whenever a new usr logs in.

The current available Smoothwall client can handle Windows users logging in via a local Active Directory server, and also via a cloud-based AzureAD server. However, our machines use the [pGina](http://pgina.org/) login handler to allow users to log in with their credentials held in a Google Workspace instance (via Google's cloud LDAP service). Other schools use the [Google Credential Provider for Windows, (GCPW")](https://tools.google.com/dlpage/gcpw/), a solution that allows for direct login to Windows devices using Google authentication (although, due to a limitation with how Windows handles USB port permissions, it doesn't seem to allow for USB-based 2FA keys to be used, which was a problem in our particular setup).

Both solutions work well, but result in what Windows sees as a "local" Windows account on the workstation. The Smoothwall client then reports this to the Smoothwall appliance in a format that doesn't match what the appliance is expecting - "TESTMACHINE01\t.user" instead of "t.user@example.com". The applince does still apply filtering (and, indeed, records the given username against requests in the logs), but using the default setting for unknown users. Users can log in using the appliances web-based login page if you have that configured, but if they don't then remember to log out using that same mechanism the next user of the machine will still be seen as them.

Therefore, instead of deploying Smoothwall's Unified Client package to our workstations we deploy RADIUS User Client, which serves much the same function but communicates with the Smoothwall server via RADIUS and can transmit the username in a format that matches up with what the appliance is expecting.

### RADIUS Notification Utility As Part Of A Remote Desktop Gateway
We provide our users with web-based a [remote desktop gateway](https://github.com/dhicks6345789/remote-gateway) solution, so users can log in to a familier desktop environment from pretty much any device able to run a web browser. The servers running the remote desktops are hosted on-site, behind our filtering gateway (Smoothwall), with external access provided, in this example, by a [Cloudflare Zero-Trust Tunnel](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/). Being behind the filtering gateway, network access from the remote desktop is filtered by the gateway, so even remote users get a consistant filtering policy applied.

As part of the login process, the remote desktop gateway could use RADIUS User Client to inform the Smoothwall server which user has just logged in to a particular remote machine. Individual users can be assigned individual machines, which might work quite nicely in an education environment for access to Raspberry Pi or similar devices - your users (school pupils) can have full root access to a Linux instance, be able to access that instance (using their existing school login details via seemless SSO) from any location, inside or outside the school, and filtering policies would still be applied. Remote desktops on shared machines would need more work - filtering gateways such as Smoothwall's tend to work by applying filtering policies to IP addresses, so you would need an IP address per user to be able to distinguish individual users.

## Building
If you want to build the application from the source code, build scripts (`build.bat`, `build.sh`) are included in the Git repository for both Windows and Linux. The whole build process is basically `git build client.go`.

If you are building the client on a test machine, there are also included test scripts (`test.bat`, `test.sh`) which build the code then do a local deploy of the executables.

## Command-Line Usage
`client [--username value] [--ipaddress value] [--service|daemon] [--debug] [--help]`

* --username: The username of the user using the device. If not defined, the client will try and figure out the current username.
* --ipaddress: The IP address of the device. If not defined, the client will try and figure out the current IP address.
* --service: Run in service / daemon mode. If run from the command line in this mode, the client will simply sit in a loop checking the current username every 30 seconds. Use the appropriate included install script to install as a Windows service (install.bat) /  Linux (or MacoOS / Unix) daemon (install.sh).
* --debug: Print debug information during execution.
* --help: Displays a help message.

The username / ipaddress options are probably more useful for testing or possibly as part of a remote gateway solution - you can inform the RADIUS server of the username associated with any IP address you like, it doesn't have to be the machine the client is running on.

In most use cases, you will want to install the client as a service / daemon, where it will inform the RADIUS server of the current username associated with the device it is running on.

## Config File
The executable expects to find a file called config.txt containing some basic configuration values. This file is a simple YAML file, with `variable:value` style pairs. An example config file is provided:
