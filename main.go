package main

import (
	"fmt"
	"bufio"
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
	pid string
	name string
	state string
	ppid string
	priority string
	niceness string
}

var Processes []Process
var pw string

func main() {
	var line, text string
	var words []string
        var out []byte

	// execute bashScript and get output
	out =  executeCommand(bashScript)
	text = string(out)
	reader := strings.NewReader(text)
	scanner := bufio.NewScanner(reader)

	// store output in structs
	for scanner.Scan(){
		line = scanner.Text()
		words = strings.Fields(line)
		words[1] = removeParentheses(words[1])
		var process = Process{pid:words[0],name:words[1],state:words[2],ppid:words[3],priority:words[17],niceness:words[18]}
		Processes = append(Processes, process) 
	}
	printProcesses()
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
func printProcesses(){
	for _,p := range(Processes){
		fmt.Printf("pid: %s, name: %s, state: %s, ppid: %s, priority: %s, niceness: %s\n",p.pid,p.name,p.state,p.ppid,p.priority,p.niceness)
	}
}
func removeParentheses(s string) string {
	var runes = []rune(s)
	runes = runes[1:len(runes)-1]
	return string(runes)
}

