package main
// RADIUS User Client - intended to run as a privelged service a (managed) workstation to keep a RADIUS server informed of the current logged-in user.
// For more details, see https://www.sansay.co.uk/docs/radius-user-client

import (
	// Standard libraries.
	"io"
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
	"bytes"
	"strings"
	"strconv"
	"context"
	"encoding/hex"
	"encoding/json"
	"crypto/sha256"

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
// If the unifiKey argument is set, either at the command line or in a config file, the application will poll the UniFi server
// for new user logons and inform the RADIUS server.
var pollUnifi = false

// A map to store any arguments passed on the command line.
var arguments = map[string]string{}

type ClientUpdateRequest struct {
    Secret string `json:"secret"`
	Username string `json:"username"`
	IPAddress string `json:"ipaddress"`
}

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

	// See if we're on Windows.
	if _, existsErr := os.Stat("C:\\Program Files"); !os.IsNotExist(existsErr) {
		getUserMethod = 1
	}

	// Try "query user" - should work on Windows 11 Enterprise / Edu, but not the Home version.
	if getUserMethod == 0 || getUserMethod == 1 {
		// Try "query user".
		queryCmd := exec.Command("cmd", "/C", "query user")
		queryOut, _ := queryCmd.CombinedOutput()
		queryResult := strings.TrimSpace(string(queryOut))
		if strings.HasPrefix(queryResult, "No User exists for") {
			// To do: figure out what to do if no user reported. Could be a user-defineable option of a username to return.
			username = ""
		} else {
			// To do: more actual parsing goes here to get the current username from a possible list of several.
			// fmt.Printf("%q\n", strings.Fields(queryResult))
			for _, queryLine := range strings.Split(queryResult, "\n") {
				lineSplit := strings.Fields(queryLine)
				if len(lineSplit) > 3 {
					if strings.TrimSpace(lineSplit[3]) == "Active" {
						username = strings.TrimLeft(lineSplit[0], ">")
						getUserMethod = 1
					}
				}
			}
		}
	}
	// Try getting the current console user - should work on MacOS.
	// stat -f "%Su" /dev/console
	if getUserMethod == 0 || getUserMethod == 2 {
		shellCmd := exec.Command("stat", "-f", "%Su", "/dev/console")
		shellOut, _ := shellCmd.CombinedOutput()
		shellResult := strings.TrimSpace(string(shellOut))
		if !strings.HasPrefix(shellResult, "stat: cannot read file system information") {
			username = shellResult
			getUserMethod = 2
		}
	}
	// Try "whoami" - should work on Linux and MacOS.
	// Note: And Windows.
	if getUserMethod == 0 || getUserMethod == 3 {
		shellCmd := exec.Command("whoami")
		shellOut, _ := shellCmd.CombinedOutput()
		shellResult := strings.TrimSpace(string(shellOut))
		username = shellResult
		getUserMethod = 3
	}
	if arguments["domain"] != "" {
		username  = username + "@" + arguments["domain"]
	}
	return username
}

func getCurrentIPAddress() string {
	IPAddress := ""

	// See if we're on Windows.
	if _, existsErr := os.Stat("C:\\Program Files"); !os.IsNotExist(existsErr) {
		getIPMethod = 1
	}

	// Try "ipconfig", should work on Windows.
	if getIPMethod == 0 || getIPMethod == 1 {
		ipconfigInterfaces := [2]string{"", ""}
		ipconfigInterface := 0
		ipconfigCmd := exec.Command("cmd", "/C", "ipconfig")
		ipconfigOut, _ := ipconfigCmd.CombinedOutput()
		ipconfigScanner := bufio.NewScanner(strings.NewReader(string(ipconfigOut)))
		for ipconfigScanner.Scan() {
			ipconfigLine := ipconfigScanner.Text()
			if strings.HasPrefix(ipconfigLine, "Ethernet") {
				ipconfigInterface = 0
			}
			if strings.HasPrefix(ipconfigLine, "Wireless") {
				ipconfigInterface = 1
			}
			if strings.HasPrefix(ipconfigLine, "   IPv4") {
				ipconfigInterfaces[ipconfigInterface] = strings.TrimSpace(strings.Split(ipconfigLine, ":")[1])
			}
		}
		if ipconfigInterfaces[0] != "" {
			IPAddress = ipconfigInterfaces[0]
			getIPMethod = 1
		} else if ipconfigInterfaces[1] != "" {
			IPAddress = ipconfigInterfaces[1]
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
	if username == "" || IPAddress == "" {
		debug("Missing value, not sending - username: " + username + ", IP address: " + IPAddress + ".")
		return
	}
	// Create a new RADIUS accounting packet.
	packet := radius.New(radius.CodeAccountingRequest, []byte(secret))
	
	// Set accounting attributes as expected by Smoothwall.
	// Other filters / gateways might expect different / additional fields - device ID (possibly the device MAC address) or a unique session ID of some sort.
	rfc2866.AcctStatusType_Add(packet, statusType)
	
	var sessionIDHash [32]byte = sha256.Sum256([]byte(username + IPAddress))
	var sessionIDString string = hex.EncodeToString(sessionIDHash[:])
    debug("SessionID: " + sessionIDString)
	
	rfc2866.AcctSessionID_SetString(packet, sessionIDString)
	rfc2865.UserName_SetString(packet, username)
	rfc2865.FramedIPAddress_Add(packet, net.ParseIP(IPAddress))
	
	// Exchange the packet with the server - waits for a response.
	debug("Sending to RADIUS server " + serverAddr + " - username: " + username + ", IP address: " + IPAddress + ".")
	response, RADIUSErr := radius.Exchange(context.Background(), packet, serverAddr)
	if RADIUSErr != nil {
		debug(fmt.Sprintf("Failed to send packet to RADIUS server: %v", RADIUSErr))
	} else {
		debug(fmt.Sprintf("Received response from server: %v", response.Code))
	}
}

// Sends a JSON update to the specified server.
func sendJSONPacket(serverAddr string, secret string, username string, IPAddress string) {
	JSONString := "{\"secret\":\"" + secret + "\",\"username\":\"" + username + "\",\"ipaddress\":\"" + IPAddress + "\"}"
	debug("Sending JSON to server " + serverAddr + ": " + JSONString)

    // Send an HTTP POST request to the specified server.
    sendJSONResponse, sendJSONErr := http.Post("http://" + serverAddr + "/clientUpdate", "application/json", bytes.NewBufferString(JSONString))
	if sendJSONErr != nil {
		debug("HTTP request to server " + serverAddr + " failed: " + sendJSONErr.Error())
		return
    }
    defer sendJSONResponse.Body.Close()

    // Read and display the response returned by the server.
    sendJSONResult, _ := io.ReadAll(sendJSONResponse.Body)
    debug("Response Status: " + string(sendJSONResponse.Status))
    debug("Response Body: " + string(sendJSONResult))
}

// Sends an iDex packet to the specified server.
func sendIDEXPacket(serverAddr string, username string, IPAddress string) {
	JSONString := "{\"username\":\"" + username + "\",\"ipaddress\":\"" + IPAddress + "\"}"
	debug("Sending iDex packet to server " + serverAddr + ": " + JSONString)

    // Send an HTTP POST request to the specified server.
    //sendJSONResponse, sendJSONErr := http.Post("http://" + serverAddr + "/clientUpdate", "application/json", bytes.NewBufferString(JSONString))
	//if sendJSONErr != nil {
		//debug("HTTP request to server " + serverAddr + " failed: " + sendJSONErr.Error())
		//return
    //}
    //defer sendJSONResponse.Body.Close()

    // Read and display the response returned by the server.
    //sendJSONResult, _ := io.ReadAll(sendJSONResponse.Body)
    //debug("Response Status: " + string(sendJSONResponse.Status))
    //debug("Response Body: " + string(sendJSONResult))
}


func sendPacket(username string, ipaddress string) {
	serverAddr := arguments["server"] + ":" + arguments["accountingport"]
	if arguments["idex"] == "true" {
		sendIDEXPacket(serverAddr, username, ipaddress)
	} else if arguments["json"] == "true" {
		sendJSONPacket(serverAddr, arguments["secret"], username, ipaddress)
	} else if arguments["radius"] == "true" {
		// Send the username and IP address to the RADIUS server.
		sendAccountingPacket(serverAddr, arguments["secret"], username, ipaddress, rfc2866.AcctStatusType_Value_Start)
	}
}
			  
func parseArguments() {
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
}

// The main body of the program. This application can act as both a simple command-line application for sending a one-off RADIUS accounting packet to a given server, and as a service that can periodically check the current user.
func main() {
	// Set some default argument values.
	arguments["help"] = "false"
	arguments["debug"] = "false"
	arguments["radius"] = "true"
	arguments["idex"] = "false"
	arguments["json"] = "false"
	arguments["service"] = "false"
	arguments["daemon"] = "false"
	arguments["jsonserver"] = "false"
	arguments["accountingport"] = "1813"
	arguments["username"] = ""
	arguments["ipaddress"] = ""
	arguments["domain"] = ""
	arguments["server"] = ""
	arguments["unifikey"] = ""
	arguments["usercheckinterval"] = "30"
	arguments["serversendinterval"] = "4"
	setArgumentIfPathExists("config", []string {"config.txt", "/etc/radiususerclient/config.txt", "/Library/Application Support/RADIUSUserClient/config.txt", "C:\\Program Files\\RadiusUserClient\\config.txt"})
	
	fmt.Println("RADIUS User Client v" + buildVersion + ". \"client --help\" for more details.")
	
	// Parse any command-line arguments.
	parseArguments()
	
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

	if arguments["json"] == "true" {
		arguments["accountingport"] = "8079"
	}
	
	// Figure out the username of the current user, unless specifically overridden by a provided config / command-line parameter.
	username := arguments["username"]
	if arguments["username"] == "" {
		username = getCurrentUser()
	}

	// Figure out the IP address of the current device, unless specifically overridden by a provided config / command-line parameter.
	ipaddress := arguments["ipaddress"]
	if arguments["ipaddress"] == "" {
		ipaddress = getCurrentIPAddress()
	}

	// Re-parse any command-line arguments - command-line arguments should override values set in the config file or found by the application.
	parseArguments()

	// Set the User Check Interval - the number of seconds where the client will check the current username.
	userCheckInterval, userCheckErr := strconv.Atoi(arguments["usercheckinterval"])
	if userCheckErr != nil {
		log.Fatalf("Error converting User Check Interval value to int: %v", userCheckErr)
	}
	// Set the Server Send Intrval - the period (this values times the User Check Interval) where the client will send the current user to the server, where that user value has changed or not.
	serverSendInterval, serverSendErr := strconv.Atoi(arguments["serversendinterval"])
	if serverSendErr != nil {
		log.Fatalf("Error converting Server Send Interval value to int: %v", serverSendErr)
	}
	
	// If debug mode is on, give a list of arguments.
	if arguments["debug"] == "true" {
		fmt.Println("Debug mode set - arguments:")
		for argName, argVal := range arguments {
			fmt.Println("   " + argName + ": " + argVal)
		}
	}
	
	if arguments["unifikey"] != "" {
		pollUnifi = true
	}
	
	if arguments["jsonserver"] == "true" {
		http.HandleFunc("/clientUpdate", func(clientUpdateResponse http.ResponseWriter, clientUpdateRequest *http.Request) {
			// We expect a JSON string passed in the body of the request, in the format: {"secret":secret, "username":username, "ipaddress":IPAddress}.
			var JSONRequest ClientUpdateRequest
			debug("clientUpdate request received - parsing request JSON string.")
			
			// Copy the JSON reader into a buffer so we can use it twice.
			var jsonBuffer bytes.Buffer
			jsonTee := io.TeeReader(clientUpdateRequest.Body, &jsonBuffer)
			
			// Decode the request's JSON...
			jsonDecoder := json.NewDecoder(jsonTee)
			clientUpdateErr := jsonDecoder.Decode(&JSONRequest)
			// ...if there's an error, we can show the original JSON request.
			if clientUpdateErr != nil {
				debug("Invalid JSON: " + clientUpdateErr.Error())
				debug(jsonBuffer.String())
				http.Error(clientUpdateResponse, "Invalid JSON: " + clientUpdateErr.Error(), http.StatusBadRequest)
				return
			}
			if JSONRequest.Secret != arguments["secret"] {
				debug("Invalid secret passed from client: " + JSONRequest.Secret)
				http.Error(clientUpdateResponse, "Invalid secret: " + JSONRequest.Secret, http.StatusBadRequest)
				return
			}
			debug("Parsed JSONrequest: " + JSONRequest.Secret + ", " + JSONRequest.Username + ", " + JSONRequest.IPAddress)
			sendPacket(JSONRequest.Username, JSONRequest.IPAddress)
			clientUpdateResponse.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(clientUpdateResponse, "{\"result\":\"" + "ok" + "\"}")
		})
		fmt.Println("Running as server on port 8079...")
		log.Fatal(http.ListenAndServe(":8079", nil))
	} else if arguments["service"] == "true" || arguments["daemon"] == "true" {
		debug("Running as service / daemon.")
		for {
			oldUsername := ""
			oldIpaddress := ""
			for pl := 0; pl < serverSendInterval; pl = pl + 1 {
				if pollUnifi {
					debug("Polling UniFi server...")
				} else {
					if arguments["username"] == "" {
						username = getCurrentUser()
					}
					if oldUsername != username || oldIpaddress != ipaddress {
						sendPacket(username, ipaddress)
						oldUsername = username
						oldIpaddress = ipaddress
					}
				}
				time.Sleep(time.Duration(userCheckInterval) * time.Second)
			}
		}
	} else {
		sendPacket(username, ipaddress)
	}
}
