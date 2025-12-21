package main
// RADIUS User Client - intended to run as a privelged service a (managed) workstation to keep a RADIUS server informed of the current logged-in user.
// For more details, see https://www.sansay.co.uk/docs/radius-user-client

import (
	// Standard libraries.
	"io"
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
	
	// Set valid authentication services.
	for _, authService := range authServices {
		authServiceNames[authService] = []string{}
	}
	
	// Set some default argument values.
	arguments["help"] = "false"
	arguments["start"] = "true"
	arguments["list"] = "false"
	arguments["new"] = "false"
	arguments["port"] = "8090"
	arguments["localonly"] = "true"
	arguments["debug"] = "false"
	arguments["shellprefix"] = ""
	arguments["cloudflare"] = "false"
	arguments["ngrok"] = "false"
	arguments["logreportinglevel"] = "none"
	setArgumentIfPathExists("webconsoleroot", []string {"/etc/webconsole", "C:\\Program Files\\WebConsole"})
	setArgumentIfPathExists("config", []string {"config.csv", "/etc/webconsole/config.csv", "C:\\Program Files\\WebConsole\\config.csv"})
	setArgumentIfPathExists("webroot", []string {"www", "/etc/webconsole/www", "C:\\Program Files\\WebConsole\\www", ""})
	setArgumentIfPathExists("taskroot", []string {"tasks", "/etc/webconsole/tasks", "C:\\Program Files\\WebConsole\\tasks", ""})
	arguments["pathprefix"] = ""
	if len(os.Args) == 1 {
		arguments["start"] = "true"
	} else {
		arguments["start"] = "false"
	}
	
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
	
	if arguments["debug"] == "true" {
		arguments["start"] = "true"
	}
	
	if arguments["start"] == "true" {
		fmt.Println("Webconsole v" + buildVersion + " - starting webserver. \"webconsole --help\" for more details.")
	}
	
	// Print the help / usage documentation if the user wanted.
	if arguments["help"] == "true" {
		//           12345678901234567890123456789012345678901234567890123456789012345678901234567890
		fmt.Println("Webconsole v" + buildVersion + ".")
		fmt.Println("")
		fmt.Println("A simple way to turn a command line application into a web app. Runs as a")
		fmt.Println("web server to host Task pages that allow the end-user to simply click a button")
		fmt.Println("to run a batch / script / etc file.")
		fmt.Println("")
		fmt.Println("Note that by itself, Webconsole doesn't handle HTTPS. If you are")
		fmt.Println("installing on a world-facing server you should use a proxy server that handles")
		fmt.Println("HTTPS - we recommend Caddy as it will automatically handle Let's Encrypt")
		fmt.Println("certificates. If you are behind a firewall then we recommend tunnelto.dev,")
		fmt.Println("giving you an HTTPS-secured URL to access. Both options can be installed via")
		fmt.Println("the install.bat / install.sh scripts.")
		fmt.Println("")
		fmt.Println("Usage: webconsole [--new] [--list] [--start] [--localOnly true/false] [--port int] [--config path] [--webroot path] [--taskroot path]")
		fmt.Println("")
		fmt.Println("--new: creates a new Task. Each Task has a unique 16-character ID which can be")
		fmt.Println("  passed as part of the URL or via a POST request, so for basic security you")
		fmt.Println("  can give a user a URL with an embedded ID. Use an external authentication")
		fmt.Println("  service for better security.")
		fmt.Println("--list: prints a list of existing Tasks.")
		fmt.Println("--start: runs as a web server, waiting for requests. Logs are printed straight to")
		fmt.Println("  stdout - hit Ctrl-C to quit. By itself, the start command can be handy for")
		fmt.Println("  quickly debugging. Run install.bat / install.sh to create a Windows service or")
		fmt.Println("  Linux / MacOS deamon.")
		fmt.Println("--debug: like \"start\", but prints more information.")
		fmt.Println("--localOnly: default is \"true\", in which case the built-in webserver will only")
		fmt.Println("  respond to requests from the local server.")
		fmt.Println("--port: the port number the web server should listen out on. Defaults to 8090.")
		fmt.Println("--config: where to find the config file. By default, on Linux this is")
		fmt.Println("  /etc/webconsole/config.csv.")
		fmt.Println("--webroot: the folder to use for the web root.")
		fmt.Println("--taskroot: the folder to use to store Tasks.")
		os.Exit(0)
	}
	
	// If we have an arument called "config", try and load the given config file (either an Excel or CSV file).
	if configPath, configFound := arguments["config"]; configFound {
		fmt.Println("Using config file: " + configPath)
		for argName, argVal := range readConfigFile(configPath) {
			arguments[argName] = argVal
		}
	}
	
	// See if we have any arguments that start with "mystart" - Page Names and API Keys for MyStart.Online login integration.
	for argName, argVal := range arguments {
		if strings.HasPrefix(argName, "mystart") {
			mystartName := ""
			if strings.HasSuffix(argName, "apikey") {
				mystartName = argName[7:len(argName)-6]
			}
			if strings.HasSuffix(argName, "pagename") {
				mystartName = argName[7:len(argName)-8]
			}
			authServiceNames["mystart"] = append(authServiceNames["mystart"], mystartName)
			if mystartName == "" {
				mystartName = "default"
			}
			if strings.HasSuffix(argName, "apikey") {
				mystartAPIKeys[mystartName] = argVal
			}
			if strings.HasSuffix(argName, "pagename") {
				mystartPageNames[mystartName] = argVal
			}
		} else if strings.HasPrefix(argName, "cloudflare") {
			if argVal != "false" {
				cloudflareName := argName[10:len(argName)]
				authServiceNames["cloudflare"] = append(authServiceNames["cloudflare"], cloudflareName)
			}
		} else if strings.HasPrefix(argName, "ngrok") {
			if argVal != "false" {
				ngrokName := argName[5:len(argName)]
				authServiceNames["ngrok"] = append(authServiceNames["ngrok"], ngrokName)
			}
		}
	}
	
	if arguments["start"] == "true" {
		// Start the thread that checks for and clears expired tokens.
		go clearExpiredTokens()
		
		// Handle the request URL.
		http.HandleFunc("/", func (theResponseWriter http.ResponseWriter, theRequest *http.Request) {
			// Make sure submitted form values are parsed.
			theRequest.ParseForm()
			
			// The default root - serve index.html.
			requestPath := theRequest.URL.Path
			
			// Print the request path.
			debug("Requested URL: " + requestPath)
			
			if strings.HasPrefix(requestPath, arguments["pathPrefix"]) {
				requestPath = requestPath[len(arguments["pathPrefix"]):]
			}
			
			userID := ""
			serveFile := false
			authorised := false
			fileToServe := filesToServeList[requestPath]
			// Handle the getPublicTaskList API call (the one API call that doesn't require authentication).
			if strings.HasPrefix(requestPath, "/api/getPublicTaskList") {
				taskList, taskErr := getTaskList()
				if taskErr == nil {
					// We return the list of public tasks in JSON format. Note that public tasks might still need authentication to run,
					// "public" here just means that they are listed by this API call for display on the landing page.
					taskListString := ""
					for _, task := range taskList {
						if task["public"] == "Y" {
							taskDetailsString, _ := json.Marshal(map[string]string{"title":task["title"], "shortDescription":task["shortDescription"], "fullDescription":task["fullDescription"], "authentication":task["authentication"]})
							taskListString = taskListString + "\"" + task["taskID"] + "\":" + string(taskDetailsString) + ","
						}
					}
					if taskListString == "" {
						fmt.Fprintf(theResponseWriter, "{}")
					} else {
						fmt.Fprintf(theResponseWriter, "{" + taskListString[:len(taskListString)-1] + "}")
					}
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
				}
			// Handle a view, run or API request. If taskID is not provided as a parameter, either via GET or POST, it defaults to "/".
			} else if fileToServe != "" || strings.HasPrefix(requestPath, "/api/") {
				taskID := theRequest.Form.Get("taskID")
				token := theRequest.Form.Get("token")
				if taskID == "" {
					taskID = "/"
				}
				taskDetails, taskErr := getTaskDetails(taskID)
				if taskErr == nil {
					// If we get to this point, we know we have a valid Task ID.
					
					authorisationError := "unknown error"
					permission := "E"
					currentTimestamp := time.Now().Unix()
					rateLimit, rateLimitErr := strconv.Atoi(taskDetails["ratelimit"])
					if rateLimitErr != nil {
						rateLimit = 0
					}
					// Check through request headers - handle a login from Cloudflare's Zero Trust product or ngrok's tunneling service. Validate
					// the details passed and check that the user ID given has permission to access this Task.
					if arguments["cloudflare"] == "true" || arguments["ngrok"] == "true" {
						for headerName, headerValue := range theRequest.Header {
							if (arguments["cloudflare"] == "true" && headerName == "Cf-Access-Authenticated-User-Email") || (arguments["ngrok"] == "true" && headerName == "Ngrok-Auth-User-Email") {
								// To do - actual authentication. Assuming local-only operation, only Cloudflare / ngrok will be passing traffic anyway, but best to check.
								userID = headerValue[0]
								// Okay - we've authenticated the user, now we need to check authorisation.
								permission = getTaskPermission(arguments["webconsoleroot"], taskDetails, userID)
								if permission == "" {
									authorisationError = "authetication attempted via header authorisation (Cloudflare / ngrok), but no valid permissions granted (you're probably missing a users file)"
								} else {
									authorised = true
									//debug("User permissions granted from header " + headerName + ", ID: " + userID + ", permission: " + permission)
								}
							}
						}
					// Handle a login from MyStart.Online - validate the details passed and check that the user ID given has
					// permission to access this Task.
					} else if strings.HasPrefix(requestPath, "/api/mystartLogin") {
						mystartLoginToken := theRequest.Form.Get("loginToken")
						if mystartLoginToken != "" {
							requestURL := fmt.Sprintf("https://dev.mystart.online/api/validateToken?loginToken=%s&pageName=%s", mystartLoginToken, arguments["mystartpagename"])
							mystartResult, mystartErr := http.Get(requestURL)
							if mystartErr != nil {
								fmt.Println("webconsole: mystartLogin - error when doing callback.")
							}
							if mystartResult.StatusCode == 200 {
								defer mystartResult.Body.Close()
								mystartJSON := new(mystartStruct)
								mystartJSONResult := json.NewDecoder(mystartResult.Body).Decode(mystartJSON)
								if mystartJSONResult == nil {
									if mystartJSON.Login == "valid" {
										debug("User authenticated via MyStart.Online login, ID: " + mystartJSON.EmailHash)
										// Okay - we've authenticated the user, now we need to check authorisation.
										permission = getTaskPermission(arguments["webconsoleroot"], taskDetails, mystartJSON.EmailHash)
										if permission != "" {
											authorised = true
											userID = mystartJSON.EmailHash
											debug("User permissions granted via MyStart.Online login, ID: " + userID + ", permission: " + permission)
										}
									}
								}
							}
						} else {
							fmt.Fprintf(theResponseWriter, "ERROR: Missing parameter loginToken.")
						}
					} else if token != "" {
						if tokens[token] == 0 {
							authorisationError = "invalid or expired token"
						} else {
							authorised = true
							permission = permissions[token]
							userID = userIDs[token]
							debug("User authorised - valid token found: " + token + ", permission: " + permission + ", user ID: " + userID)
						}
					} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretViewers"]) {
						authorised = true
						permission = "V"
						debug("User authorised via Task secret, permission: " + permission)
					} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretRunners"]) {
						authorised = true
						permission = "R"
						debug("User authorised via Task secret, permission: " + permission)
					} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretEditors"]) {
						authorised = true
						permission = "E"
						debug("User authorised via Task secret, permission: " + permission)
					} else {
						authorisationError = "no external authorisation used, no valid secret given, no valid token supplied"
					}
					if !authorised && taskDetails["authentication"] == "" {
						if taskDetails["public"] == "Y" {
							debug("User authorised - no other authentication method defined, Task is public, assigning Runner permsisions.")
							permission = "R"
						} else {
							debug("User authorised - no other authentication method defined, Task is not public, assigning Viewer permsisions.")
							permission = "V"
						}
						authorised = true
						authorisationError = ""
					}
					if authorised {
						// If we get this far, we know the user is authorised for this Task - they've either provided a valid
						// secret or no secret is set.
						if token == "" {
							token = generateRandomString()
							debug("New token generated: " + token)
						}
						tokens[token] = currentTimestamp
						permissions[token] = permission
						userIDs[token] = userID
							
						// Handle view and run requests - no difference server-side, only the client-side treates the URLs differently
						// (the "runTask" method gets called by the client-side code if the URL contains "run" rather than "view").
						if fileToServe != "" {
							doServeFile(theResponseWriter, theRequest, fileToServe, taskID, token, permission, taskDetails["title"], taskDetails["shortDescription"], taskDetails["fullDescription"])
						// API - Handle a request for a list of "private" Tasks, i.e. Tasks that the user has explicit
						// authorisation to view, run or edit. We return the list of private tasks in JSON format.
						} else if strings.HasPrefix(requestPath, "/api/getPrivateTaskList") {
							taskList, taskErr := getTaskList()
							taskListString := ""
							if taskErr == nil {
								for _, task := range taskList {
									// Don't list Tasks that would already be listed in the "public" list.
									// Also, don't list special Tasks like the "new-task" Task.
									if task["public"] != "Y" && task["taskID"] != "new-task" {
										listTask := false
										// If we have Edit permissions for the root Task
										// (ID "/"), then we have permissions to view all Tasks.
										if taskID == "/" && permission == "E" {
											listTask = true
										} else {
											// Otherwise, work out permissions for each Task.
											taskPermission := getTaskPermission(arguments["webconsoleroot"], task, userID)
											if taskPermission == "V" || taskPermission == "R" || taskPermission == "E" {
												listTask = true
											}
										}
										if listTask {
											taskDetailsString, _ := json.Marshal(map[string]string{"title":task["title"], "shortDescription":task["shortDescription"], "fullDescription":task["fullDescription"], "authentication":task["authentication"]})
											taskListString = taskListString + "\"" + task["taskID"] + "\":" + string(taskDetailsString) + ","
										}
									}
								}
								if taskListString == "" {
									fmt.Fprintf(theResponseWriter, "{}")
								} else {
									fmt.Fprintf(theResponseWriter, "{" + taskListString[:len(taskListString)-1] + "}")
								}
							} else {
								fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
							}
						// API - Exchange the secret for a token.
						} else if strings.HasPrefix(requestPath, "/api/getToken") {
							fmt.Fprintf(theResponseWriter, token)
						// API - Return the Task's title.
						} else if strings.HasPrefix(requestPath, "/api/getTaskDetails") {
							fmt.Fprintf(theResponseWriter, taskDetails["title"] + "\n" + taskDetails["shortDescription"] + "\n" + taskDetails["fullDescription"])
						// API - Return the Task's result URL (or blank if it doesn't have one).
						} else if strings.HasPrefix(requestPath, "/api/getResultURL") {
							_, checkWWWErr := os.Stat(arguments["taskroot"] + "/" + taskID + "/www")
							if taskDetails["resultURL"] == "" && checkWWWErr == nil {
								fmt.Fprintf(theResponseWriter, "www")
							} else {
								fmt.Fprintf(theResponseWriter, taskDetails["resultURL"])
							}
						// API - Run a given Task.
						} else if strings.HasPrefix(requestPath, "/api/runTask") {
							// Check for appropriate permissions (Editor or Runner) for the user to be able to run the given Task.
							if permission == "E" || permission == "R" {
								// If the Task is already running, simply return "OK".
								if taskIsRunning(taskID) {
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									// Check to see if there's any rate limit set for this task, and don't run the Task if we're still
									// within the rate limited time.
									if currentTimestamp - taskStopTimes[taskID] < int64(rateLimit) {
										fmt.Fprintf(theResponseWriter, "ERROR: Rate limit (%d seconds) exceeded - try again in %d seconds.", rateLimit, int64(rateLimit) - (currentTimestamp - taskStopTimes[taskID]))
									} else {
										// Get ready to run the Task - set up the Task's details...
										if strings.HasPrefix(taskDetails["command"], "webconsole ") {
											taskDetails["command"] = strings.Replace(taskDetails["command"], "webconsole ", "\"" + arguments["webconsoleroot"] + string(os.PathSeparator) + "webconsole\" ", 1)
										} else {
											taskDetails["command"] = strings.TrimSpace(strings.TrimSpace(arguments["shellprefix"]) + " " + taskDetails["command"])
										}
										commandArray := parseCommandString(taskDetails["command"])
										var commandArgs []string
										if len(commandArray) > 0 {
											commandArgs = commandArray[1:]
										}
										debug("Task ID " + taskID + " - running command: " + commandArray[0])
										debug("With arguments: " + strings.Join(commandArgs, ","))
										debug("Ran by user: " + userID)
										
										taskRunUsers[taskID] = userID
										runningTasks[taskID] = exec.Command(commandArray[0], commandArgs...)
										runningTasks[taskID].Dir = arguments["taskroot"] + "/" + taskID
										
										// ...get a list (if available) of recent run times...
										taskRunTimes[taskID] = make([]int64, 0)
										runTimesBytes, fileErr := ioutil.ReadFile(arguments["taskroot"] + "/" + taskID + "/runTimes.txt")
										if fileErr == nil {
											runTimeSplit := strings.Split(string(runTimesBytes), "\n")
											for pl := 0; pl < len(runTimeSplit); pl = pl + 1 {
												runTimeVal, runTimeErr := strconv.Atoi(runTimeSplit[pl])
												if runTimeErr == nil {
													taskRunTimes[taskID] = append(taskRunTimes[taskID], int64(runTimeVal))
												}
											}
										}
										
										// ...use those to guess the run time for this time (just use a simple mean of the
										// existing runtimes)...
										var totalRunTime int64
										totalRunTime = 0
										for pl := 0; pl < len(taskRunTimes[taskID]); pl = pl + 1 {
											totalRunTime = totalRunTime + taskRunTimes[taskID][pl]
										}
										if len(taskRunTimes[taskID]) == 0 {
											taskRuntimeGuesses[taskID] = float64(10)
										} else {
											taskRuntimeGuesses[taskID] = float64(totalRunTime / int64(len(taskRunTimes[taskID])))
										}
										taskStartTimes[taskID] = time.Now().Unix()
										
										// ...then run the Task as a goroutine (thread) in the background.
										go runTask(taskID)
										// Respond to the front-end code that all is okay.
										fmt.Fprintf(theResponseWriter, "OK")
									}
								}
							} else {
								fmt.Fprintf(theResponseWriter, "ERROR: runTask - don't have runner / editor permissions.")
							}
						} else if strings.HasPrefix(requestPath, "/api/cancelTask") {
							// Check for appropriate permissions for the user to be able to cancel the running task...
							if permission == "E" || permission == "R" {
								// ...and that the given Task ID is actually a currently running Task...
								if taskIsRunning(taskID) {
									// and that the current user is the user that ran the Task in the first place.
									if taskRunUsers[taskID] == userID {
										if cancelErr := runningTasks[taskID].Process.Kill(); cancelErr == nil {
											delete(runningTasks, taskID)
											fmt.Fprintf(theResponseWriter, "OK")
										} else {
											fmt.Fprintf(theResponseWriter, "ERROR: cancelTask - Unable to terminate specified Task.")
										}
									} else {
										fmt.Fprintf(theResponseWriter, "ERROR: cancelTask - Current user not original runner of Task.")
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: cancelTask - given Task not running.")
								}
							} else {
								fmt.Fprintf(theResponseWriter, "ERROR: cancelTask - don't have runner / editor permissions.")
							}
						// Designed to be called periodically, will return the given Tasks' output as a simple string,
						// with lines separated by newlines. Takes one parameter, "line", indicating which output line
						// it should return output from, to save the client-side code having to be sent all of the output each time.
						} else if strings.HasPrefix(requestPath, "/api/getTaskOutput") {
							var atoiErr error
							// Parse the "line" parameter - defaults to 0, so if not set this method will simply return
							// all current output.
							outputLineNumber := 0
							if theRequest.Form.Get("line") != "" {
								outputLineNumber, atoiErr = strconv.Atoi(theRequest.Form.Get("line"))
								if atoiErr != nil {
									fmt.Fprintf(theResponseWriter, "ERROR: Line number not parsable.")
								}
							}
							if _, runningTaskFound := runningTasks[taskID]; !runningTaskFound {
								// If the Task isn't currently running, load the previous run's log file (if it exists)
								// into the Task's output buffer.
								logContents, logContentsErr := ioutil.ReadFile(arguments["taskroot"] + "/" + taskID + "/log.txt")
								if logContentsErr == nil {
									taskOutputs[taskID] = strings.Split(string(logContents), "\n")
								}
							} else if taskDetails["progress"] == "Y" {
								// If the job details have the "progress" option set to "Y", output a (best guess, using previous
								// run times) progresss report line.
								currentTime := time.Now().Unix()
								percentage := int((float64(currentTime - taskStartTimes[taskID]) / taskRuntimeGuesses[taskID]) * 100)
								if percentage > 100 {
									percentage = 100
								}
								taskOutputs[taskID] = append(taskOutputs[taskID], fmt.Sprintf("Progress: Progress %d%%", percentage))
							}
							// Return to the user all the output lines from the given starting point.
							for outputLineNumber < len(taskOutputs[taskID]) {
								fmt.Fprintln(theResponseWriter, taskOutputs[taskID][outputLineNumber])
								outputLineNumber = outputLineNumber + 1
							}
							// If the Task is no longer running, make sure we tell the client-side code that.
							if _, runningTaskFound := runningTasks[taskID]; !runningTaskFound {
								if taskDetails["progress"] == "Y" {
									fmt.Fprintf(theResponseWriter, "Progress: Progress 100%%\n")
								}
								// Remove any files from the task's "output" folder that are past a defined age (in seconds - default is 129600, 36 hours).
								logfilePath := arguments["taskroot"] + "/" + taskID + "/output"
								logItems, itemErr := os.ReadDir(logfilePath)
								if itemErr == nil {
									for pl := 0; pl < len(logItems); pl = pl + 1 {
										parsedTimestamp, timestampErr := time.Parse(time.RFC3339, strings.Split(logItems[pl].Name(), "Z")[0]+"Z")
										if timestampErr == nil {
											taskLogPeriod, _ := strconv.ParseInt(taskDetails["logPeriod"], 10, 64)
											if parsedTimestamp.Unix() < taskStartTimes[taskID] - taskLogPeriod {
												os.Remove(logfilePath + "/" + logItems[pl].Name())
											}
										}
									}
								} else {
									debug("Error reading items in path: " + logfilePath)
								}
								if taskDetails["resultURL"] != "" {
									debug("Sending client resultURL: " + taskDetails["resultURL"])
									fmt.Fprintf(theResponseWriter, "ERROR: REDIRECT " + taskDetails["resultURL"])
								} else if _, err := os.Stat(arguments["taskroot"] + "/" + taskID + "/www"); err == nil {
									debug("www subfolder found, sending client redirect.")
									fmt.Fprintf(theResponseWriter, "ERROR: REDIRECT")
								} else {
									debug("Sending client EOF.")
									fmt.Fprintf(theResponseWriter, "ERROR: EOF")
								}
								//delete(taskOutputs, taskID)
							}
						// Simply returns "YES" if a given Task is running, "NO" otherwise.
						} else if strings.HasPrefix(requestPath, "/api/getTaskRunning") {
							if taskIsRunning(taskID) {
								fmt.Fprintf(theResponseWriter, "YES")
							} else {
								fmt.Fprintf(theResponseWriter, "NO")
							}
						// Submits the given string as input to the given Task. The user needs to be the runner of the Task.
						} else if strings.HasPrefix(requestPath, "/api/submitInputText") {
							// Check the user has permission (Editor or Runner) to submit inputs to the given Task...
							if permission == "E" || permission == "R" {
								// ...and that the given Task is still running...
								if taskIsRunning(taskID) {
									// ...and that the user is the runner of the Task...
									if taskRunUsers[taskID] == userID {
										// ...and that we have a value to submit.
										value := theRequest.Form.Get("value")
										if value != "" {
											debug("submitInputText - taskID: " + taskID + ", value: " + value);
											// Note - we might need to lose the newline addition on Windows.
											io.WriteString(taskInputs[taskID], value + "\n")
										} else {
											fmt.Fprintf(theResponseWriter, "ERROR: submitInputText - missing value.");
										}
									} else {
										fmt.Fprintf(theResponseWriter, "ERROR: submitInputText - current user not original runner of Task.");
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: submitInputText - given Task not a current running Task.")
								}
							} else {
								fmt.Fprintf(theResponseWriter, "ERROR: submitInput called - don't have runner / editor permissions.")
							}
						// Return a list of editable files for this task, as a JSON structure - needs edit permissions.
						} else if strings.HasPrefix(requestPath, "/api/getEditableFileList") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: getEditableFileList called - don't have edit permissions.")
							} else {
								path := theRequest.Form.Get("path")
								debug("getEditableFileList - path: " + path)
								if path != "" {
									outputString := "[\n"
									outputString = outputString + listOneFolderAsJSON(arguments["taskroot"] + "/" + taskID + path)
									outputString = outputString + "]"
									fmt.Fprintf(theResponseWriter, outputString)
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: getEditableFileList - missing path parameter.")
								}
							}
						// Return the contents of an editable file - needs edit permissions.
						} else if strings.HasPrefix(requestPath, "/api/getEditableFileContents") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: getEditableFileContents called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									http.ServeFile(theResponseWriter, theRequest, arguments["taskroot"] + "/" + taskID + "/" + filename)
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: getEditableFileContents - missing filename parameter.")
								}
							}
						// Return the (zipped) contents of a folder - needs edit permissions.
						} else if strings.HasPrefix(requestPath, "/api/getZippedFolderContents") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: getZippedFolderContents called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									// Create a buffer and writer to write the Zip file to.
									zipBuf := new(bytes.Buffer)
									zipWriter := zip.NewWriter(zipBuf)
									zipFolder := normalisePath(arguments["taskroot"] + string(os.PathSeparator) + taskID + string(os.PathSeparator) + filename)
									debug("Zipping folder: " + zipFolder)
									zipErr := getZippedFolderContents(zipWriter, zipFolder, "")
									if zipErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: getZippedFolderContents - %s", zipErr.Error())
									} else {
										// Make sure the Zip buffer / writer is properly finished (the Zip file will be invalid otherwise).
										zipWriter.Close()
										// Return the zipped folder data to the user.
										http.ServeContent(theResponseWriter, theRequest, filename + ".zip", time.Now(), bytes.NewReader(zipBuf.Bytes()))
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: getZippedFolderContents - missing filename parameter.")
								}
							}
						// Save a file.
						} else if strings.HasPrefix(requestPath, "/api/startSaveFile") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: startSaveFile called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									debug("Zero: " + arguments["taskroot"] + "/" + taskID + "/" + filename)
									ioutil.WriteFile(arguments["taskroot"] + "/" + taskID + "/" + filename, []byte{}, 0644)
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: startSaveFile - missing filename parameter.")
								}
							}
						// Save a file.
						} else if strings.HasPrefix(requestPath, "/api/saveFileChunk") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: saveFileChunk called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									contents := theRequest.Form.Get("contents")
									if contents != "" {
										debug("Write " + arguments["taskroot"] + "/" + taskID + "/" + filename)
										base64Contents, _ := base64.StdEncoding.DecodeString(contents)
										outfile, _ := os.OpenFile(arguments["taskroot"] + "/" + taskID + "/" + filename, os.O_APPEND|os.O_WRONLY, 0644)
										_, _ = outfile.Write(base64Contents)
										outfile.Close()
										// ioutil.WriteFile(arguments["taskroot"] + "/" + taskID + "/" + filename, []byte(base64Contents), 0644)
										fmt.Fprintf(theResponseWriter, "OK")
									} else {
										fmt.Fprintf(theResponseWriter, "ERROR: saveFileChunk - missing contents parameter.")
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: saveFileChunk - missing filename parameter.")
								}
							}
						// Delete a file.
						} else if strings.HasPrefix(requestPath, "/api/deleteFile") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: deleteFile called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									debug("Delete file " + arguments["taskroot"] + "/" + taskID + "/" + filename)
									os.Remove(arguments["taskroot"] + "/" + taskID + "/" + filename)
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: deleteFile - missing filename parameter.")
								}
							}
						// Delete a folder.
						} else if strings.HasPrefix(requestPath, "/api/deleteFolder") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: deleteFolder called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									debug("Delete folder " + arguments["taskroot"] + "/" + taskID + "/" + filename)
									os.RemoveAll(arguments["taskroot"] + "/" + taskID + "/" + filename)
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: deleteFolder - missing filename parameter.")
								}
							}
						// Rename a file.
						} else if strings.HasPrefix(requestPath, "/api/renameFile") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: renameFile called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									newFilename := theRequest.Form.Get("newFilename")
									if newFilename != "" {
										debug("Rename " + arguments["taskroot"] + "/" + taskID + "/" + filename + " to " + arguments["taskroot"] + "/" + taskID + "/" + newFilename)
										os.Rename(arguments["taskroot"] + "/" + taskID + "/" + filename, arguments["taskroot"] + "/" + taskID + "/" + newFilename)
										fmt.Fprintf(theResponseWriter, "OK")
									} else {
										fmt.Fprintf(theResponseWriter, "ERROR: renameFile - missing newFilename parameter.")
									}
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: renameFile - missing filename parameter.")
								}
							}
						// Create a new, empty, text-format file.
						} else if strings.HasPrefix(requestPath, "/api/newFile") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: newFile called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									debug("New file " + arguments["taskroot"] + "/" + taskID + "/" + filename)
									newFile, newFileErr := os.OpenFile(arguments["taskroot"] + "/" + taskID + "/" + filename, os.O_RDONLY|os.O_CREATE, 0644)
									if newFileErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: newFile - " + newFileErr.Error())
									}
									newFile.Close()
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: newFile - missing filename parameter.")
								}
							}
						// Create a new folder.
						} else if strings.HasPrefix(requestPath, "/api/newFolder") {
							if permission != "E" {
								fmt.Fprintf(theResponseWriter, "ERROR: newFolder called - don't have edit permissions.")
							} else {
								filename := theRequest.Form.Get("filename")
								if filename != "" {
									debug("New folder " + arguments["taskroot"] + "/" + taskID + "/" + filename)
									os.Mkdir(arguments["taskroot"] + "/" + taskID + "/" + filename, os.ModePerm)
									fmt.Fprintf(theResponseWriter, "OK")
								} else {
									fmt.Fprintf(theResponseWriter, "ERROR: newFolder - missing filename parameter.")
								}
							}
						// Return the hash value of the given secret value.
						} else if strings.HasPrefix(requestPath, "/api/hashSecret") {
							theSecret := theRequest.Form.Get("secret")
							hashedSecret, hashErr := hashPassword(theSecret)
							if hashErr == nil {
								hashPermission := "secret"
								if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretViewers"]) {
									hashPermission = "viewer"
								} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretRunners"]) {
									hashPermission = "runner"
								} else if checkPasswordHash(theRequest.Form.Get("secret"), taskDetails["secretEditors"]) {
									hashPermission = "editor"
								}
								fmt.Fprintf(theResponseWriter, hashPermission + "," + hashedSecret)
							} else {
								fmt.Fprintf(theResponseWriter, "ERROR: Problem hashing secret - " + hashErr.Error())
							}
						// A simple call that doesn't do anything except serve to keep the timestamp for the given Task up-to-date.
						} else if strings.HasPrefix(requestPath, "/api/keepAlive") {
							fmt.Fprintf(theResponseWriter, "OK")
						// To do: return API documentation here.
						} else if strings.HasPrefix(requestPath, "/api/") {
							fmt.Fprintf(theResponseWriter, "ERROR: Unknown API call: %s", requestPath)
						}
					} else if strings.HasPrefix(requestPath, "/login") {
						doServeFile(theResponseWriter, theRequest, fileToServe, taskID, "", "", taskDetails["title"], taskDetails["shortDescription"], taskDetails["fullDescription"])
					} else {
						fmt.Fprintf(theResponseWriter, "ERROR: Not authorised - %s.", authorisationError)
					}
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: %s", taskErr.Error())
				}
			} else if strings.HasSuffix(requestPath, "/site.webmanifest") {
				taskID := ""
				taskList, taskErr := getTaskList()
				if taskErr == nil {
					for _, task := range taskList {
						if strings.HasPrefix(requestPath, "/" + task["taskID"]) {
							taskID = task["taskID"] + "/"
						}
					}
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
				}
				webmanifestBuffer, fileReadErr := ioutil.ReadFile(arguments["webroot"] + "/" + "site.webmanifest")
				if fileReadErr == nil {
					webmanifestString := string(webmanifestBuffer)
					webmanifestString = strings.Replace(webmanifestString, "<<TASKID>>", arguments["pathPrefix"] + "/" + taskID, -1)
					http.ServeContent(theResponseWriter, theRequest, "site.webmanifest", time.Now(), strings.NewReader(webmanifestString))
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: Couldn't read site.webmanifest.")
				}
			} else {
				// Check to see if the request is for a favicon of some description.
				faviconTitle := ""
				faviconHyphens := 0
				faviconTitles := [7]string{ "favicon.*png", "mstile.*png", "android-chrome.*png", "apple-touch-icon.*png", "safari-pinned-tab.png", "safari-pinned-tab.svg", "favicon.ico" }
				for _, titleMatch := range faviconTitles {
					requestMatch, _ := regexp.MatchString(".*/" + titleMatch + "$", requestPath)
					if requestMatch {
						faviconTitle = titleMatch
						faviconHyphens = strings.Count(titleMatch, "-") + 1
					}
				}
				// If the request was for a favicon, serve something suitible.
				if faviconTitle != "" {
					faviconPath := arguments["webroot"] + "/" + "favicon.png"
					taskList, taskErr := getTaskList()
					if taskErr == nil {
						for _, task := range taskList {
							if strings.HasPrefix(requestPath, "/" + task["taskID"]) {
								// Does this Task have a custom favicon?
								faviconPath = arguments["taskroot"] + "/" + task["taskID"] + "/" + "favicon.png"
								if _, fileExistsErr := os.Stat(faviconPath); os.IsNotExist(fileExistsErr) {
									// Does all Tasks have a custom favicon?
									faviconPath = arguments["taskroot"] + "/" + "favicon.png"
									if _, fileExistsErr := os.Stat(faviconPath); os.IsNotExist(fileExistsErr) {
										// If there is no custom favicon set for this Task, use the default.
										faviconPath = arguments["webroot"] + "/" + "favicon.png"
									}
								}
							}
						}
					} else {
						fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
					}
					faviconFile, faviconFileErr := os.Open(faviconPath)
					if faviconFileErr == nil {
						serveFile = true
						faviconImage, _, faviconImageErr := image.Decode(faviconFile)
						faviconFile.Close()
						if faviconImageErr == nil {
							faviconWidth := faviconImage.Bounds().Max.X
							faviconHeight := faviconImage.Bounds().Max.Y
							if faviconTitle == "safari-pinned-tab.png" || faviconTitle == "safari-pinned-tab.svg" {
								silhouetteImage := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{faviconWidth, faviconHeight}})
								for silhouetteY := 0; silhouetteY < faviconHeight; silhouetteY++ {
									for silhouetteX := 0; silhouetteX < faviconWidth; silhouetteX++ {
										r, g, b, a := faviconImage.At(silhouetteX, silhouetteY).RGBA()
										if r > 128 || g > 128 || b > 128 || a < 255 {
											silhouetteImage.Set(silhouetteX, silhouetteY, color.RGBA{255, 255, 255, 0})
										} else {
											silhouetteImage.Set(silhouetteX, silhouetteY, color.RGBA{0, 0, 0, 255})
										}
									}
								}
								if faviconTitle == "safari-pinned-tab.png" {
									pngErr := png.Encode(theResponseWriter, silhouetteImage)
									if pngErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: Unable to encode PNG silhouette image.\n")
									}
								} else {
									tracedImage, _ := gotrace.Trace(gotrace.NewBitmapFromImage(silhouetteImage, nil), nil)
									theResponseWriter.Header().Set("Content-Type", "image/svg+xml")
									gotrace.WriteSvg(theResponseWriter, silhouetteImage.Bounds(), tracedImage, "")
								}
								serveFile = false
							} else {
								if faviconTitle == "apple-touch-icon.png" {
									faviconWidth = 180
									faviconHeight = 180
								} else if faviconTitle == "favicon.ico" {
									faviconWidth = 48
									faviconHeight = 48
								}
								// Resize the available (PNG) favicon to match the request.
								faviconSplit := strings.Split(requestPath, "/")
								faviconName := strings.Split(faviconSplit[len(faviconSplit)-1], ".")[0]
								faviconSplit = strings.Split(faviconName, "-")
								if len(faviconSplit) != faviconHyphens {
									faviconSizeSplit := strings.Split(faviconSplit[faviconHyphens], "x")
									if len(faviconSizeSplit) == 2 {
										var atoiError error
										faviconWidth, atoiError = strconv.Atoi(faviconSizeSplit[0])
										if atoiError == nil {
											faviconHeight, atoiError = strconv.Atoi(faviconSizeSplit[1])
										}
										if atoiError != nil {
											fmt.Fprintf(theResponseWriter, "ERROR: Non-integer in image dimensions.\n")
											serveFile = false
										}
									}
								}
								resizedImage := resize.Resize(uint(faviconWidth), uint(faviconHeight), faviconImage, resize.Lanczos3)
								if strings.HasSuffix(faviconTitle, "ico") {
									icoErr := ico.Encode(theResponseWriter, resizedImage)
									if icoErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: Unable to encode PNG image.\n")
									}
									serveFile = false
								} else {
									pngErr := png.Encode(theResponseWriter, resizedImage)
									if pngErr != nil {
										fmt.Fprintf(theResponseWriter, "ERROR: Unable to encode PNG image.\n")
									}
									serveFile = false
								}
							}
						} else {
							fmt.Fprintf(theResponseWriter, "ERROR: Couldn't decode favicon file: " + faviconPath + "\n")
						}
					} else {
						fmt.Fprintf(theResponseWriter, "ERROR: Couldn't open favicon file: " + faviconPath + "\n")
					}
				// ...otherwise, just serve the static file referred to by the request URL.
				} else {
					serveFile = true
				}
			}
			if serveFile == true {
				localFilePath := ""
				taskList, taskErr := getTaskList()
				if taskErr == nil {
					for _, task := range taskList {
						if strings.HasPrefix(requestPath, "/" + task["taskID"]) && serveFile == true {
							var filePath = strings.TrimSpace(requestPath[len(task["taskID"])+1:])
							if filePath == "" {
								http.Redirect(theResponseWriter, theRequest, "/" + task["taskID"] + "/", http.StatusPermanentRedirect)
								serveFile = false
							}
							if strings.HasSuffix(filePath, "/") {
								filePath = filePath + "index.html"
							}
							localFilePath = arguments["taskroot"] + "/" + task["taskID"] + "/www" + filePath
						}
					}
				} else {
					fmt.Fprintf(theResponseWriter, "ERROR: " + taskErr.Error())
					serveFile = false
				}
				if serveFile == true {
					// Serve a static file. A file found in the appropriate "webconsole/users/username" folder will override one found in "webconsole/www",
					// so you can serve user-specific content for authenticated users simply by placing files in the appropriate user subfolder.
					if localFilePath == "" {
						usersFilePath := arguments["webconsoleroot"] + "/users/" +  userID + requestPath
						if _, err := os.Stat(usersFilePath); !errors.Is(err, os.ErrNotExist) {
							localFilePath = usersFilePath
						} else {
							localFilePath = arguments["webroot"] + requestPath
						}
					}
					if _, err := os.Stat(localFilePath); !errors.Is(err, os.ErrNotExist) {
						debug("Serving: " + localFilePath)
						http.ServeFile(theResponseWriter, theRequest, localFilePath)
					} else {
						theResponseWriter.WriteHeader(http.StatusNotFound)
						errorFileContent, errorFileErr := ioutil.ReadFile(arguments["webroot"] + "/404.html")
						if errorFileErr == nil {
							fmt.Fprint(theResponseWriter, strings.Replace(string(errorFileContent), "{{FILENAME}}", localFilePath, -1))
						} else {
							fmt.Fprint(theResponseWriter, "Error 404: File " + localFilePath + " not found.")
						}
					}
				}
			}
		})
		// Run the main web server loop.
		hostname := ""
		if (arguments["localonly"] == "true") {
			fmt.Println("Web server limited to localhost only.")
			hostname = "localhost"
		}
		fmt.Println("Web server using webroot " + arguments["webroot"] + ", taskroot " + arguments["taskroot"] + ".")
		fmt.Println("Web server available at: http://localhost:" + arguments["port"] + "/")
		// If we detect ngrok is running, search the syslog for the ngrok URL to display to the user.
		if _, err := os.Stat("/usr/local/bin/ngrok"); err == nil {
			ngrokURL, ngrokErr := exec.Command("bash", "-c", "cat /var/log/syslog | grep ngrok.*localhost | tail -1 | cut -d '=' -f 8").CombinedOutput()
			if ngrokErr != nil {
				fmt.Println("ERROR: " + ngrokErr.Error())
			}
			fmt.Println("ngrok URL: " + strings.TrimSpace(string(ngrokURL)))
		}
		if arguments["debug"] == "true" {
			fmt.Println("Debug mode set - arguments:")
			for argName, argVal := range arguments {
				fmt.Println("   " + argName + ": " + argVal)
			}
		}
		log.Fatal(http.ListenAndServe(hostname + ":" + arguments["port"], nil))
	// Command-line option to print a list of all Tasks.
	} else if arguments["list"] == "true" {
		fmt.Println("Reading Tasks from " + arguments["taskroot"])
		taskList, taskErr := getTaskList()
		if taskErr == nil {
			for _, task := range taskList {
				secret := "Y"
				if task["secret"] == "" {
					secret = "N"
				}
				fmt.Println(task["taskID"] + ": " + task["title"] + ", Secret: " + secret + ", Public: " + task["public"] + ", Command: " + task["command"])
			}
		} else {
			fmt.Println("ERROR: " + taskErr.Error())
		}
	// Generate a new Task.
	} else if arguments["newdefaulttask"] == "true" {
		// Generate a new Task ID, checking it doesn't already exist.
		var newTaskID string
		for {
			newTaskID = generateRandomString()
			if _, err := os.Stat(arguments["taskroot"] + "/" + newTaskID); os.IsNotExist(err) {
				break
			}
		}
		
		os.Mkdir(arguments["taskroot"], os.ModePerm)
		os.Mkdir(arguments["taskroot"] + "/" + newTaskID, os.ModePerm)
		fmt.Println("New Task: " + newTaskID)
		
		// Write the config file - a simple text file, one value per line.
		outputString := "title: Task " + newTaskID + "\npublic: N\ncommand: "
		writeFileErr := ioutil.WriteFile(arguments["taskroot"] + "/" + newTaskID + "/config.txt", []byte(outputString), 0644)
		if writeFileErr != nil {
			fmt.Println("ERROR: Couldn't write config for Task " + newTaskID + ".")
		}
	} else if arguments["new"] == "true" {
		// Generate a new, unique Task ID.
		var newTaskID string
		var newTaskIDExists bool
		// Ask the user to provide a Task ID (or they can use the one we just generated).
		if newTaskID, newTaskIDExists = arguments["newtaskid"]; !newTaskIDExists {
			for {
				newTaskID = generateRandomString()
				if _, err := os.Stat(arguments["taskroot"] + "/" + newTaskID); os.IsNotExist(err) {
					break
				}
			}
			newTaskID = getUserInput("newtaskid", newTaskID, "Enter a new Task ID (hit enter to generate an ID)")
		}
		if _, err := os.Stat(arguments["taskroot"] + "/" + newTaskID); os.IsNotExist(err) {
			// We use simple text files in folders for data storage, rather than a database. It seemed the most logical choice - you can stick
			// any resources associated with a Task in that Task's folder, and editing options can be done with a basic text editor.
			os.Mkdir(arguments["taskroot"], os.ModePerm)
			os.Mkdir(arguments["taskroot"] + "/" + newTaskID, os.ModePerm)
			fmt.Println("New Task: " + newTaskID)
			
			// Get a title for the Task.
			newTaskTitle := "Task " + newTaskID
			newTaskTitle = getUserInput("newtasktitle", newTaskTitle, "Enter a title (hit enter for \"" + newTaskTitle + "\")")
			
			// Get a secret for the Task - blank by default, although that's not the same as a public Task.
			newTaskSecret := ""
			newTaskSecret = getUserInput("newtasksecret", newTaskSecret, "Set secret (type secret, or hit enter to skip)")
			
			// Ask the user if this Task should be public, "N" by default.
			var newTaskPublic string
			for {
				newTaskPublic = "N"
				newTaskPublic = strings.ToUpper(getUserInput("newtaskpublic", newTaskPublic, "Make this task public (\"Y\" or \"N\", hit enter for \"N\")"))
				if newTaskPublic == "Y" || newTaskPublic == "N" {
					break
				}
			}
			
			// The command the Task runs. Can be anything the system will run as an executable application, which of course depends on which platform
			// you are running.
			newTaskCommand := ""
			newTaskCommand = getUserInput("newtaskcommand", newTaskCommand, "Set command (type command, or hit enter to skip)")
			
			// Hash the secret (if not just blank).
			outputString := ""
			if newTaskSecret != "" {
				hashedPassword, hashErr := hashPassword(newTaskSecret)
				if hashErr == nil {
					outputString = outputString + "secret: " + hashedPassword + "\n"
				} else {
					fmt.Println("ERROR: Problem hashing password - " + hashErr.Error())
				}
			}
			
			// Write the config file - a simple text file, one value per line.
			outputString = outputString + "title: " + newTaskTitle + "\npublic: " + newTaskPublic + "\ncommand: " + newTaskCommand
			writeFileErr := ioutil.WriteFile(arguments["taskroot"] + "/" + newTaskID + "/config.txt", []byte(outputString), 0644)
			if writeFileErr != nil {
				fmt.Println("ERROR: Couldn't write config for Task " + newTaskID + ".")
			}
		} else {
			fmt.Println("ERROR: A task with ID " + newTaskID + " already exists.")
		}		
	}
}
