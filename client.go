package main
// RADIUS User Client - intended to run as a privelged service a (managed) workstation to keep a RADIUS server informed of the current logged-in user.
// For more details, see https://www.sansay.co.uk/docs/radius-user-client

import (
	// Standard libraries.
	"os";
	"os/exec"
	"fmt";
	"log";
	"time";
	"strings";
	"io/ioutil";

	// A Go YAML parser.
	"gopkg.in/yaml.v3";
)

// The current release version - value provided via the command line at compile time.
var buildVersion string

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
	
	// Try "query user".
	queryCmd := exec.Command("cmd", "/C", "query user && exit 0")
	queryOut, _ := queryCmd.CombinedOutput()
	queryResult := strings.TrimSpace(string(queryOut))
	if strings.HasPrefix(queryResult, "No User exists for") {
		fmt.Println("To do: figure out what to do if no user reported.")
	} else {
		username = strings.TrimSpace(strings.Split(queryResult, " ")[3])
	}
	return username
}

// The main body of the program. This application can act as both a simple command-line application for sending a one-off RADIUS accounting packet to a given server, and as a service that can periodically check the current user.
func main() {
	// Set some default argument values.
	arguments["help"] = "false"
	arguments["debug"] = "false"
	arguments["service"] = "false"
	arguments["accountingPort"] = "1813"
	arguments["domain"] = ""
	setArgumentIfPathExists("config", []string {"config.txt", "/etc/radiususerclient/config.txt", "C:\\Program Files\\RadiusUserClient\\config.txt"})
	
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

	if arguments["debug"] == "true" {
		fmt.Println("Debug mode set - arguments:")
		for argName, argVal := range arguments {
			fmt.Println("   " + argName + ": " + argVal)
		}
	}
	
	if arguments["service"] == "true" {
		fmt.Println("Running as service.")
		fmt.Println("Service code goes here.")
	} else {
		fmt.Println("Current user: " + getCurrentUser())
	}
}
