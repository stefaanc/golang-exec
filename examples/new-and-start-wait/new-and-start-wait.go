package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "github.com/stefaanc/golang-exec/script"
    "github.com/stefaanc/golang-exec/runner"
)

func main() {
    // define connection to the server
    c := map[string]string{
       "Type": "ssh",
       "Host": "localhost",
       "Port": "22",
       "User": "me",
       "Password": "my-password",
       "Insecure": "true",
    }

    // create script runner
    r, err := runner.New(c, lsScript, lsArguments{
        Path: "~\\Projects\\golang-exec\\examples\\new-and-start-wait\\hhh",
    })
    if err != nil {
         log.Fatal(err)
    }
    defer r.Close()

    // get a stdout-reader
    stdout, err := r.StdoutPipe()
    if err != nil {
         log.Fatal(err)
    }

    // start script runner
    err = r.Start()
    if err != nil {
         log.Fatal(err)
    }

    // wait for stdout-reader to complete
    result, err := ioutil.ReadAll(stdout)
    if err != nil {
         log.Fatal(err)
    }

    // wait for script runner to complete
    err = r.Wait()
    if err != nil {
         log.Fatal(err)
    }

    // write the result
    fmt.Printf("exitcode: %d\n", r.ExitCode())
    fmt.Printf("result:   %q\n", string(result))
}

type lsArguments struct{
    Path string
}

var lsScript = script.New("ls", "powershell", `
    $ErrorActionPreference = 'Stop'

    $path = "{{.Path}}"
    Get-ChildItem -Path $path | Format-Table

    exit 0
`)
