package main

import (
    "fmt"
    "log"
    "os"
    "github.com/stefaanc/golang-exec/runner"
    "github.com/stefaanc/golang-exec/runner/ssh"
    "github.com/stefaanc/golang-exec/script"
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
    wd, _ := os.Getwd()
    err := runner.Run(c, rmScript, rmArguments{
//        Path: wd + "\\doesn't exist",
        Path: wd + "\\test",
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

    $dirpath = "{{.Path}}"
    Get-ChildItem -Path $dirpath | Remove-Item

    exit 0
`)
