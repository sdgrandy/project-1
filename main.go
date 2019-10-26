package main

import (
    "bufio"
    "os/exec"
    "fmt"
    "strings"
)


func main() {
    var line string
    var words []string
    bashScript := `
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
    
    cmd := exec.Command("ssh", "user1@192.168.56.103", bashScript)
    // "ps -eo pid;touch file101.txt")
    reader, _ := cmd.StdoutPipe()
    scanner := bufio.NewScanner(reader)
    go func() {
	for scanner.Scan() {
		line = scanner.Text()
		words = strings.Fields(line)
		fmt.Println(words[0])
	}
    }()
    /*cmd.Stdout = &out
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    } else {
	fmt.Println(out.String())
    }
    cmd.Run()*/
    cmd.Start()
    cmd.Wait()
}
