package main

import (
    "fmt"
    "log"
    "github.com/stefaanc/golang-exec/script"
    "github.com/stefaanc/golang-exec/runner"
    "github.com/stefaanc/golang-exec/runner/ssh"
)

func main() {
    // define connection to the server
    c := ssh.Connection{
       Type: "ssh",
       Host: "localhost",
       Port: 22,
       User: "me",
       Password: "my-password",
       Insecure: true,
    }

    // create script runner
    err := runner.Run(c, rmScript, rmArguments{
        Path: "~\\Projects\\golang-exec\\examples\\run\\test.txt",
    })
    if err != nil {
         log.Fatal(err)
    }

    // write the result
    fmt.Printf("done")
}

type rmArguments struct{
    Path string
}

var rmScript = script.New("rm", "powershell", `
    $ErrorActionPreference = 'Stop'

    $path = "{{.Path}}"
    Get-ChildItem -Path $path | Remove-Item

    exit 0
`)
