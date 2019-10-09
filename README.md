# Golang Exec

**a golang package to run scripts locally or via SSH**

Scripts can be defined as a string or can be loaded from a file.  They are parsed as a golang template.  This template can then be rendered with template-arguments.  The resulting rendered code is either executed on the local machine or remotely via SHH.  A shell is started on the machine, the script is loaded via stdin, and is executed in the shell.  Results from the script can be received from stdout.  Errors can be received from stderr.

The main use-case for this is
- script-code needs to be embedded in the golang executable
- script needs to run local or on remote servers


A typical example of this is the development of a terraform provider - this is the reason why we developed this.




<br/>

### !!! UNDER CONSTRUCTION !!!!!!!!

<br/>



## Basic Use

A couple of very (, very, very) basic examples

### Using runner.Run()

### Using runner.New() and r.Run()

```golang
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
    Port     int16
    User     string
    Password string
    Insecure bool
}

func main {
    // check if script is successfully parsed
    err := dirScript.Error
    if err != nil {
         log.Fatal(err)
    }

    // define connection to the server
    c := myConnection{
       Type: "ssh"
       Host: "localhost"
       Port: 22
       User: "me"
       Password: "my-password"
       Insecure: true
    }

    // create script runner
    r, err := runner.New(c, dirScript, dirArguments{
        Path: "~/Documents",
    })
    if err != nil {
         log.Fatal(err)
    }
    defer r.Close()

    // create buffer to capture stdout
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

type dirArguments struct{
    Path string
}

var dirScript = script.New("dir", "powershell", `
$ErrorActionPreference = 'Stop'

$path = "{{.Path}}"
Get-ChildItem -Path $path | Format-Table

exit 0
`)
```

### Using runner.New() and r.Start() / r.Wait()

```golang

```



<br/>

## More Info

A brief overview of the most important elements of this package, to give an idea what the main user-structs, -funcs and -methods are, and to give an idea how this package is build and hangs together.

```golang
// script/script.go
package script

import (
    "text/template"
    //...
)

type Script struct {
    Name       string
    Error      error    // error from New()
 
    shell      string   // "cmd", powershell", "bash", "sh", ...
    template   *template.Template
    //...
}

func New(name string, shell string, code string) *Script {...}
    // remark that New() doesn't return any errors directly
    // instead, error are saved in the 'Error'-field of the returned script
    // this allows using New() in a package scope, while checking for errors in a function scope

func NewFromString(name string, shell string, code string) (*Script, error) { /*...*/ }

func NewFromFile(name string, shell string, file string) (*Script, error) { /*...*/ }

func (s *Script) Command() string {
    // returns the command(s) to execute a script that is read from stdin
    switch strings.ToLower(s.shell) {
    case "cmd":
        return "set T=_temp~%RANDOM%.bat && more > %T% && cmd /C %T% && del /Q %T%"
    case "powershell":
        return "PowerShell -NoProfile -ExecutionPolicy ByPass -Command -"
    default:
        return s.shell + " -"
    }
}
```

```
// runner/runner.go
package runner

import (
    "io"
    "github.com/stefaanc/golang-exec/script"
    //...
)

type Runner interface {
    SetStdoutWriter(io.Writer)
    SetStderrWriter(io.Writer)
    StdoutPipe() (io.Reader, error)   // use in combination with Start() & Wait(), don't use in combination with Run()
    StderrPipe() (io.Reader, error)   // use in combination with Start() & Wait(), don't use in combination with Run()

    Run() error
    Start() error
    Wait() error
    Close() error

    ExitCode() int   // -1 when runner error without completing script
}

func New(connection interface {}, s *script.Script, arguments interface{}) (Runner, error) { /*...*/ }
```

For a local runner

```golang
// runner/local/runner.go
package local

import (
    "context"
    "os/exec"
    //...
)

type Connection struct {
    Type string   // "local"
}

type Runner struct {
    cmd      *exec.Cmd
    cancel   context.CancelFunc
    exitCode int
    //...
}
```

For a remote runner

```golang
// runner/ssh/runner.go
package ssh

import (
    "golang.org/x/crypto/ssh/knownhosts"
    "golang.org/x/crypto/ssh"
    //...
)

type Connection struct {
    Type     string   // "ssh"
    Host     string
    Port     int16
    User     string
    Password string
    Insecure bool
}

type Runner struct {
    client  *ssh.Client
    session *ssh.Session
    exitCode int
    //...
}
```



<br/>

## For Further Investigation

- support for SSH auth using certificates instead of password
- support for Pageant on Windows
- support for SSH bastion server
- support for WinRM communication