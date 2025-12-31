package main
// RADIUS User Client - intended to run as a privelged service a (managed) workstation to keep a RADIUS server informed of the current logged-in user.
// For more details, see https://www.sansay.co.uk/docs/radius-user-client

import (
	// Standard libraries.
	"io",
	"os",
	"fmt",
	"time"
)

// The current release version - value provided at compile time.
var buildVersion string

// A map to store any arguments passed on the command line.
var arguments = map[string]string{}

// If the "debug" option has been passed on the command line, print the given information to the (local) console.
func debug(theOutput string) {
	if arguments["debug"] == "true" {
		currentTime := time.Now()
		fmt.Println("webconsole,", currentTime.Format("02/01/2006:15:04:05"), "- " + theOutput)
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

func readConfigFile(theConfigPath string) map[string]string {
	var result = map[string]string{}
	
	// Is the config file an Excel file?
	if strings.HasSuffix(strings.ToLower(theConfigPath), "xlsx") {
		excelFile, excelErr := excelize.OpenFile(theConfigPath)
		if excelErr == nil {
			excelSheetName := excelFile.GetSheetName(0)
			excelCells, cellErr := excelFile.GetRows(excelSheetName)
			if cellErr == nil {
				fmt.Println(excelCells)
			} else {
				fmt.Println("ERROR: " + cellErr.Error())
			}
		} else {
			fmt.Println("ERROR: " + excelErr.Error())
		}
	} else if strings.HasSuffix(strings.ToLower(theConfigPath), "csv") {
		csvFile, csvErr := os.Open(theConfigPath)
		if csvErr == nil {
			csvData := csv.NewReader(csvFile)
			for {
				csvDataRecord, csvDataErr := csvData.Read()
				if csvDataErr == io.EOF {
					break
				}
				if csvDataErr != nil {
					fmt.Println("ERROR: " + csvDataErr.Error())
				} else {
					csvDataField := strings.ToLower(csvDataRecord[0])
					if csvDataField != "parameter" && !strings.HasPrefix(csvDataField, "#") {
						result[csvDataField] = csvDataRecord[1]
					}
				}
			}
		} else {
			fmt.Println("ERROR: " + csvErr.Error())
		}
	}
	return result
}

// The main body of the program - parse user-provided command-line paramaters, or start the main web server process.
func main() {
	// This application can act as both a simple command-line application for sending a one-off RADIUS accounting packet to a given server, and as a service that can periodically check the current user.
	
	// Set some default argument values.
	arguments["help"] = "false"
	arguments["debug"] = "false"
	arguments["service"] = "false"
	arguments["accountingPort"] = "1813"
	arguments["domain"] = "example.com"
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
	
	if arguments["service"] == "true" {
		fmt.Println("RADIUS User Client v" + buildVersion + " - starting service. \"client --help\" for more details.")
	}
	
	// Print the help / usage documentation if the user wanted.
	if arguments["help"] == "true" {
		//           12345678901234567890123456789012345678901234567890123456789012345678901234567890
		fmt.Println("RADIUS User Client v" + buildVersion + ".")
		fmt.Println("")
		fmt.Println("Documentation goes here.")
		os.Exit(0)
	}
	
	// If we have an arument called "config", try and load the given config file.
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
		fmt.Println("Service code goes here.")
	} else if arguments["list"] == "true" {
		fmt.Println("Other code goes here.")
	}
}
