package main
// RADIUS User Client - intended to run as a privelged service a (managed) workstation to keep a RADIUS server informed of the current logged-in user.
// For more details, see https://www.sansay.co.uk/docs/radius-user-client

import (
	// Standard libraries.
	"os"
	"os/exec"
	"fmt"
	"log"
	"net"
	"time"
	"strings"
	"strconv"
	"context"
	"io/ioutil"

	// A YAML parser.
	"gopkg.in/yaml.v3"

	// A RADIUS library.
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

// The current release version - value provided via the command line at compile time.
var buildVersion string

// Remember how we get the current user / IP address.
var getUserMethod = 0
var getIPMethod = 0

// If running in service mode, this application can act as an intermidiate service between UniFi and your (Smoothwall, etc) RADIUS accounting server.
// If the unifiServer and unifiKey arguments are set, either at the command line or in a config file, the application will poll the UniFi server
// for new user logons and inform the RADIUS server.
var pollUnifi = false

// A map to store any arguments passed on the command line.
var arguments = map[string]string{}

// If the "debug" option has been passed on the command line, print the given information to the (local) console.
func debug(theOutput string) {
	if arguments["debug"] == "true" {
		currentTime := time.Now()
		fmt.Println("radius-user-client,", currentTime.Format("02/01/2006:15:04:05"), "- " + theOutput)
	}
}

// A helper function that sets the given "arguments" value to the first discovered valid path from a list given as an array of strings.
func setArgumentIfPathExists(theArgument string, thePaths []string) {
	for _, path := range thePaths {
		if _, existsErr := os.Stat(path); !os.IsNotExist(existsErr) {
			arguments[theArgument] = path
			return
		}
	}
}

// A function to read a config file in YAML format. Returns a map[string]string holding the config file's data.
func readConfigFile(theConfigPath string) map[string]string {
	// Map to store the parsed YAML data.
	var result = map[string]string{}

	// Read the YAML data (just a bunch of strings) from a file...
	YAMLFileData, YAMLFileErr := ioutil.ReadFile(theConfigPath)
	if YAMLFileErr != nil {
		log.Fatalf("Error reading YAML config file: %v", YAMLFileErr)
    }
	
	// ...and unmarshal it into the result data map.
	YAMLParseErr := yaml.Unmarshal(YAMLFileData, &result)
	if YAMLParseErr != nil {
		log.Fatalf("Error parsing YAML config file: %v", YAMLParseErr)
	}
	
	return result
}

// Get the current username from the system. Method used varies as to the system this client is running on.
func getCurrentUser() string {
	username := ""

	tmpCmd := exec.Command("cmd", "/C", "query user > \"C:\\Program File\\RADIUSUserClient\\log.txt\"")

	// Try "query user" - should work on Windows 11 Enterprise / Edu, but not the Home version.
	if getUserMethod == 0 || getUserMethod == 1 {
		// Try "query user".
		queryCmd := exec.Command("cmd", "/C", "query user")
		queryOut, _ := queryCmd.CombinedOutput()
		queryResult := strings.TrimSpace(string(queryOut))
		if strings.HasPrefix(queryResult, "No User exists for") {
			// To do: figure out what to do if no user reported. Could be a user-defineable option of a username to return.
			username = "default"
		} else {
			// To do: more actual parsing goes here to get the current username from a possible list of several.
			// fmt.Printf("%q\n", strings.Fields(queryResult))
			for _, queryLine := range strings.Split(queryResult, "\n") {
				lineSplit := strings.Fields(queryLine)
				if len(lineSplit) > 3 {
					if lineSplit[3] == "Active" {
						username = strings.TrimLeft(lineSplit[0], ">")
						getUserMethod = 1
					}
				}
			}
		}
	}
	// Try "whoami" - should work on Linux and MacOS.
	if getUserMethod == 0 || getUserMethod == 2 {
		shellCmd := exec.Command("whoami")
		shellOut, _ := shellCmd.CombinedOutput()
		shellResult := strings.TrimSpace(string(shellOut))
		username = shellResult
		getUserMethod = 2
	}
	if arguments["domain"] != "" {
		username  = username + "@" + arguments["domain"]
	}
	return username
}

func getCurrentIPAddress() string {
	IPAddress := ""

	// Try "ipconfig", should work on Windows.
	if getIPMethod == 0 || getIPMethod == 1 {
		ipconfigCmd := exec.Command("cmd", "/C", "ipconfig | findstr /c:IPv4")
		ipconfigOut, _ := ipconfigCmd.CombinedOutput()
		ipconfigResult := string(ipconfigOut)
		lineSplit := strings.Split(ipconfigResult, ":")
		if len(lineSplit) > 1 {
			IPAddress = strings.TrimSpace(lineSplit[1])
			getIPMethod = 1
		}
	}
	// Try "hostname", should work on Linux.
	if getIPMethod == 0 || getIPMethod == 2 {
		shellCmd := exec.Command("hostname", "--all-ip-addresses")
		shellOut, _ := shellCmd.CombinedOutput()
		shellResult := strings.TrimSpace(strings.Fields(string(shellOut))[0])
		if shellResult != "hostname:" {
			IPAddress = shellResult
			getIPMethod = 2
		}
	}
	// Try "ifconfig", should work onMacOS.
	if getIPMethod == 0 || getIPMethod == 3 {
		shellCmd := exec.Command("bash", "-c", "ifconfig | grep inet | grep -v inet6")
		shellOut, _ := shellCmd.CombinedOutput()
		shellResult := string(shellOut)
		for _, shellLine := range strings.Split(shellResult, "\n") {
			lineSplit := strings.Fields(shellLine)
			if len(lineSplit) > 1 {
				if lineSplit[1] != "127.0.0.1" {
					IPAddress = lineSplit[1]
					getIPMethod = 3
				}
			}
		}
	}
	return IPAddress
}

// Sends a RADIUS accounting request to the specified server.
func sendAccountingPacket(serverAddr string, secret string, username string, IPAddress string, statusType rfc2866.AcctStatusType) {
	// Create a new RADIUS accounting packet.
	packet := radius.New(radius.CodeAccountingRequest, []byte(secret))
	
	// Set accounting attributes as expected by Smoothwall.
	// Other filters / gateways might expect different / additional fields - device ID (possibly the device MAC address) or a unique session ID of some sort.
	rfc2866.AcctStatusType_Add(packet, statusType)
	rfc2865.UserName_SetString(packet, username)
	rfc2865.FramedIPAddress_Add(packet, net.ParseIP(IPAddress))
	
	// Exchange the packet with the server - waits for a response.
	debug("Sending to RADIUS server - username: " + username + ", IP address: " + IPAddress + ".")
	response, RADIUSErr := radius.Exchange(context.Background(), packet, serverAddr)
	if RADIUSErr != nil {
		debug(fmt.Sprintf("Failed to send packet to RADIUS server: %v", RADIUSErr))
	} else {
		debug(fmt.Sprintf("Received response from server: %v", response.Code))
	}
}

// The main body of the program. This application can act as both a simple command-line application for sending a one-off RADIUS accounting packet to a given server, and as a service that can periodically check the current user.
func main() {
	// Set some default argument values.
	arguments["help"] = "false"
	arguments["debug"] = "false"
	arguments["service"] = "false"
	arguments["daemon"] = "false"
	arguments["accountingPort"] = "1813"
	arguments["username"] = ""
	arguments["ipaddress"] = ""
	arguments["domain"] = ""
	arguments["server"] = ""
	arguments["unifiServer"] = ""
	arguments["unifiKey"] = ""
	arguments["userCheckInterval"] = "30"
	arguments["serverSendInterval"] = "4"
	setArgumentIfPathExists("config", []string {"config.txt", "/etc/radiususerclient/config.txt", "/Library/Application Support/RADIUSUserClient/config.txt", "C:\\Program Files\\RadiusUserClient\\config.txt"})
	
	// Parse any command line arguments.
	currentArgKey := ""
	for _, argVal := range os.Args {
		if strings.HasPrefix(argVal, "--") {
			if currentArgKey != "" {
				arguments[strings.ToLower(currentArgKey[2:])] = "true"
			}
			currentArgKey = argVal
		} else {
			if currentArgKey != "" {
				arguments[strings.ToLower(currentArgKey[2:])] = argVal
			}
			currentArgKey = ""
		}
	}
	if currentArgKey != "" {
		arguments[strings.ToLower(currentArgKey[2:])] = "true"
	}

	fmt.Println("RADIUS User Client v" + buildVersion + ". \"client --help\" for more details.")
	
	// Print the help / usage documentation if the user wanted.
	if arguments["help"] == "true" {
		//           12345678901234567890123456789012345678901234567890123456789012345678901234567890
		fmt.Println("Documentation goes here.")
		os.Exit(0)
	}
	
	// If we have an argument called "config", try and load the given config file.
	if configPath, configFound := arguments["config"]; configFound {
		fmt.Println("Using config file: " + configPath)
		for argName, argVal := range readConfigFile(configPath) {
			arguments[argName] = argVal
		}
	}
	
	// Figure out the username of the current user, unless specifically overridden by a provided command-line parameter.
	username := arguments["username"]
	if arguments["username"] == "" {
		username = getCurrentUser()
	}

	// Figure out the IP address of the current device, unless specifically overridden by a provided command-line parameter.
	ipaddress := arguments["ipaddress"]
	if arguments["ipaddress"] == "" {
		ipaddress = getCurrentIPAddress()
	}

	// Set the User Check Interval - the number of seconds where the client will check the current username.
	userCheckInterval, userCheckErr := strconv.Atoi(arguments["userCheckInterval"])
	if userCheckErr != nil {
		log.Fatalf("Error converting User Check Interval value to int: %v", userCheckErr)
	}
	// Set the Server Send Intrval - the period (this values times the User Check Interval) where the client will send the current user to the server, where that user value has changed or not.
	serverSendInterval, serverSendErr := strconv.Atoi(arguments["serverSendInterval"])
	if serverSendErr != nil {
		log.Fatalf("Error converting Server Send Interval value to int: %v", serverSendErr)
	}
	
	if arguments["debug"] == "true" {
		fmt.Println("Debug mode set - arguments:")
		for argName, argVal := range arguments {
			fmt.Println("   " + argName + ": " + argVal)
		}
	}
	
	if arguments["unifiServer"] != "" && arguments["unifiKey"] != "" {
		pollUnifi = true
	}
	
	if arguments["service"] == "true" || arguments["daemon"] == "true" {
		debug("Running as service / daemon.")
		for {
			oldUsername := ""
			for pl := 0; pl < serverSendInterval; pl = pl + 1 {
				if pollUnifi {
					debug("Polling UniFi server...")
				} else {
					if arguments["username"] == "" {
						username = getCurrentUser()
					}
					if oldUsername != username {
						// Send the username and IP address to the RADIUS server.
						sendAccountingPacket(arguments["server"] + ":" + arguments["accountingPort"], arguments["secret"], username, ipaddress, rfc2866.AcctStatusType_Value_Start)
						oldUsername = username
					}
				}
				time.Sleep(time.Duration(userCheckInterval) * time.Second)
			}
		}
	} else {
		// Send the username and IP address to the RADIUS server.
		sendAccountingPacket(arguments["server"] + ":" + arguments["accountingPort"], arguments["secret"], username, ipaddress, rfc2866.AcctStatusType_Value_Start)
	}
}
