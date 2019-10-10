package main

import (
    "bytes"
    "fmt"
    "log"
    "github.com/stefaanc/golang-exec/script"
    "github.com/stefaanc/golang-exec/runner"
)

type myConnection struct {
    Type     string
    Host     string
    Port     uint16
    User     string
    Password string
    Insecure bool
}

func main() {
    // define connection to the server
    c := myConnection{
       Type: "ssh",
       Host: "localhost",
       Port: 22,
       User: "me",
       Password: "my-password",
       Insecure: true,
    }

    // create script runner
    r, err := runner.New(c, lsScript, lsArguments{
        Path: "~\\Projects\\golang-exec\\examples\\new-and-run",
    })
    if err != nil {
         log.Fatal(err)
    }
    defer r.Close()

    // create buffer to capture stdout, set a stdout-writer
    var stdout bytes.Buffer
    r.SetStdoutWriter(&stdout)

    // run script runner
    err = r.Run()
    if err != nil {
         log.Fatal(err)
    }

    // write the result
    fmt.Printf(stdout.String())
}

type lsArguments struct{
    Path string
}

var lsScript = script.New("", "powershell", `
    $ErrorActionPreference = 'Stop'

    $path = "{{.Path}}"
    Get-ChildItem -Path $path | Format-Table

    exit 0
`)
