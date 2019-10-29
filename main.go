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
	// bash script to traverse home directory of remote user
	bashScript = `
		traverse() {
        	for file in "$1"/*
        	do
                	if [ -d "$file" ]  
                        then 
				echo "*directory* ${file##*/} in $1" 
                        	traverse "$file"
                	elif [ -f "$file" ]
			then
                        	echo "*file* ${file##*/} in $1 "
                	fi
        	done
     		}
     		traverse "/home/user1"`
)

type RegFile struct {
	location string
	name string
}

type Dir struct {
	location string
	name string
	children []string
}

var Files []RegFile
var Dirs []Dir
var pw string

func main() {
	var line string
	var words []string
        var out []byte
	
	out = executeCommand("cd project-0;cat main.go")
	
	// cast bytes to string and display
	text := string(out)
	fmt.Println(text)

	// print home directory
	out =  executeCommand(bashScript)
	text = string(out)
	reader := strings.NewReader(text)
	scanner := bufio.NewScanner(reader)

	// store contents of home directory in structs
	for scanner.Scan(){
		line = scanner.Text()
		words = strings.Fields(line)

		if words[0] == "*directory*" {
			fmt.Println(line)
			var d = Dir{location: words[3], name: words[1], children: nil}
			Dirs = append(Dirs,d)
		} else if words[0] == "*file*" {
			fmt.Println(line)
			var f = RegFile{location: words[3], name: words[1]}
			Files = append(Files,f)
			addChildFile(f)
		} 
	}
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
func addChildFile(f RegFile) {
	lastIndex := strings.LastIndex(f.location,"/")
	runes := []rune(f.location)
	sub := string(runes[0:lastIndex])
	for _,v := range Dirs{
		if v.location == sub {
			v.children = append(v.children,string(runes[lastIndex+1]))
		}
	}
}


