package main

import (
	"fmt"

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
                	if [ -d "$file" ]; then 
                        	echo "*directory* ${file##*/} in $1" 
                        	traverse "$file"
                	fi
                	if [ -f "$file" ]; then
                        	echo "*file* ${file##*/} in $1 "
				echo "*beginFile*"
                        	#less "$file"
				echo "*endFile*"
                	fi
        	done
     		}
     		traverse "/home/user1"`
)

func main() {
	// connect to remote host
	connection, session := connect()

	// execute bash script on remote host and return its combined standard output and standard error
	out, err := session.CombinedOutput(bashScript)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	connection.Close()
}

func connect() (*ssh.Client, *ssh.Session) {
	var pw string
	fmt.Print("password: ")
	fmt.Scan(&pw)
	fmt.Print("\n")

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
