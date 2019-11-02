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
	port string = "22"
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
	Html template.HTML
	Proc   []Process
}

var Processes []Process
var remoteUser, remoteHost, pw string
var html string

func main() {
	// populate struct slice with info about processes in remote machine
	fillSlice()
	// handle requests
	http.HandleFunc("/", index)
	http.HandleFunc("/search", search)
	http.HandleFunc("/tree", tree)
	http.ListenAndServe(":7000", nil)
}

func connect() (*ssh.Client, *ssh.Session) {
	if remoteUser == "" {
		fmt.Print("remoteUser: ")
                fmt.Scan(&remoteUser)
                fmt.Print("\n")
		fmt.Print("remoteHost: ")
                fmt.Scan(&remoteHost)
                fmt.Print("\n")
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
// for debugging
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
	// check if reached page as a result of user clicking refresh button
	ref := request.FormValue("rf")
	if ref == "refresh" {
		fillSlice()
	}
	// send information about processes in response
	v := ViewInfo{}
	//printTree(Process{Pid: "0"},"")
	v.Proc = Processes
	temp.Execute(response, v)
}
func tree(response http.ResponseWriter, request *http.Request){
	temp, _ := template.ParseFiles("tree.html")
	p := Process{Pid: "0",Name: "", State: "", Ppid: "", Priority: "", Niceness: ""}
	html = "<!DOCTYPE html><head><title></title></head><body>"
	printTree(p)
	html += "</body></html>"
	v := ViewInfo{}
	v.Html = template.HTML(html)
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
func fillSlice(){
	var line, text string
        var words []string
        var out []byte
	
	// clear slice
	Processes = nil	

        // execute bashScript and get output
        out = executeCommand(bashScript)
        text = string(out)
        reader := strings.NewReader(text)
        scanner := bufio.NewScanner(reader)

        // store output in Processes struct slice
        for scanner.Scan() {
                line = scanner.Text()
                words = strings.Fields(line)
                words[1] = removeParentheses(words[1])
		var process = Process{}
                process.Pid = words[0]
		process.Name = words[1]
		process.State = words[2]
		process.Ppid = words[3]
		process.Priority = words[17]
		process.Niceness = words[18]
		Processes = append(Processes, process)
        }

}
func hasChildren(pid string) bool{
	for _,p := range(Processes){
		if p.Ppid == pid{
			return true
		}
        }
	return false
}
func printTree(proc Process){
	if proc.Ppid == "0"{
		html += proc.Pid
		html += "&nbsp"
		html += proc.Name
	} else if proc.Pid != "0"{
		html += "<li>"
		html += proc.Pid
		html += "&nbsp"
		html += proc.Name
		html += "</li>"
	}
	if hasChildren(proc.Pid) {
		if proc.Pid != "0"{
			html += "<ul>"
		}
		for _,p :=range(Processes){
			if p.Ppid == proc.Pid{
				printTree(p)
			}	
		}
		if proc.Pid != "0"{
			html += "</ul>"
		}
	}
}
