package main

import (
	"bufio"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"golang.org/x/crypto/ssh"
)

const (
	remoteUser string = "user1"
	remoteHost string = "192.168.56.103"
	port       string = "22"
	// bash script to get info from proc directory
	bashScript = `
		regex="^-?[0-9]+$"
        	dir="/proc"
		for file in "$dir"/*
		do
			if [[ ${file##*/} =~ $regex ]]
	  	   	then 
				#printf "%s\n" "$file"
				cat $file/stat
			fi 
		done`
)

type Process struct {
	Pid      string
	Name     string
	State    string
	Ppid     string
	Priority string
	Niceness string
}

type ViewInfo struct {
	ByName bool
	ByPid  bool
	ByPpid bool
	NotFound bool
	Proc   []Process
}

var Processes []Process
var pw string

func main() {
	var line, text string
	var words []string
	var out []byte

	// execute bashScript and get output
	out = executeCommand(bashScript)
	text = string(out)
	reader := strings.NewReader(text)
	scanner := bufio.NewScanner(reader)

	// store output in structs
	for scanner.Scan() {
		line = scanner.Text()
		words = strings.Fields(line)
		words[1] = removeParentheses(words[1])
		var process = Process{Pid: words[0], Name: words[1], State: words[2], Ppid: words[3], Priority: words[17], Niceness: words[18]}
		Processes = append(Processes, process)
	}
	//printProcesses()
	http.HandleFunc("/", index)
	http.HandleFunc("/search", search)
	http.ListenAndServe(":7000", nil)
}

func connect() (*ssh.Client, *ssh.Session) {
	if pw == "" {
		fmt.Print("password: ")
		fmt.Scan(&pw)
		fmt.Print("\n")
	}
	// configure authentication
	sshConfig := &ssh.ClientConfig{
		User:            remoteUser,
		Auth:            []ssh.AuthMethod{ssh.Password(pw)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// start a client connection to SSH server
	connection, err := ssh.Dial("tcp", remoteHost+":"+port, sshConfig)
	if err != nil {
		connection.Close()
		panic(err)
	}
	// create session
	session, err := connection.NewSession()
	if err != nil {
		session.Close()
		panic(err)
	}

	return connection, session
}
func executeCommand(cmd string) []byte {
	//connect to remote host
	connection, session := connect()
	// execute bash script on remote host and return its combined standard output and standard error
	out, _ := session.CombinedOutput(cmd)

	defer connection.Close()
	defer session.Close()
	return out
}
func printProcesses() {
	for _, p := range Processes {
		fmt.Printf("pid: %s, name: %s, state: %s, ppid: %s, priority: %s, niceness: %s\n", p.Pid, p.Name, p.State, p.Ppid, p.Priority, p.Niceness)
	}
}
func removeParentheses(s string) string {
	var runes = []rune(s)
	runes = runes[1 : len(runes)-1]
	return string(runes)
}

func index(response http.ResponseWriter, request *http.Request) {
	temp, _ := template.ParseFiles("index.html")
	v := ViewInfo{}
	v.Proc = Processes
	//fmt.Println("name:",Processes[0].Name)
	temp.Execute(response, v)
}

func search(response http.ResponseWriter, request *http.Request) {
	temp, _ := template.ParseFiles("search.html")
	var Result []Process
	choice := request.FormValue("choice")
	query := request.FormValue("query")
	v := ViewInfo{}
	if choice == "byname" {
		v.ByName = true
		for _, p := range Processes {
			if p.Name == query {
				Result = append(Result, p)
			}
		}
	} else if choice == "bypid" {
		v.ByPid = true 
                for _, p := range Processes {
                        if p.Pid == query {
                                Result = append(Result, p)
                        }
                } 

	} else if choice == "byppid" { 
		v.ByPpid = true 
                for _, p := range Processes {
                        if p.Ppid == query {
                                Result = append(Result, p)
                        }
                } 

        }
	if len(Result) == 0 {
		v.NotFound = true
	}
	v.Proc = Result
	temp.Execute(response, v)
}
