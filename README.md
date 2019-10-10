# Golang Exec

**a golang package to run scripts locally or remotely**

The main use-case for this package is to run scripts that are embedded in a golang executable, and run them locally or remotely.  An example where this is used is in the development of a terraform provider (this is the reason why we developed this).

Scripts can be defined at development-time as a string or can be read at run-time from a file.  They are parsed as a golang template.  This template is then rendered with template-arguments.  The resulting rendered code is either executed on the local machine or remotely.  A shell is started on the machine, the script is loaded via `stdin`, and is executed in the shell.  A number of shells are supported: windows cmd, powershell, bash, sh, ... The script is uploaded via `stdin` to avoid having to separately upload it before running (for instance using SCP) and to avoid having to clean up after running.  Results from the script can be received from the shell's `stdout`.  Errors can be received from the shell's `stderr`.

As an alternative to using the `Connection` types from the specific runners - "golang-exec/runner/local" or "golang-exec/runner/ssh" - you can make your own connection struct type or embed this in your own bigger struct type.  Your struct type must contain the relevant fields with the same field-names as the fields required for the specific runner.  The golang-exec package uses reflection to extract the connection information from your struct.  This is useful when the connection type needs to be configurable, so it is not known in advance which specific runner will be used.

For a `"local"` runner, you need at least the following fields

```golang
type Connection struct {
    Type string   // must be "local"
}
```

For a `"ssh"` runner, you need at least the following fields

```golang
type Connection struct {
    Type     string   // must be "ssh"
    Host     string
    Port     uint16
    User     string
    Password string
    Insecure bool
}
```

As another alternative to using the `Connection` types from the specific runners, you can also use a map: `map[string]string`.  Disadvantage of this is that fields with a non-string type are not statically type-checked.



<br/>

## Basic Use

A couple of very basic examples.

### Using runner.Run()

To test this, make a file `~\Projects\golang-exec\examples\run\test.txt` (or change the path  in the code to a file of your choosing).

```golang
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
```

> Remark that in this example
> - we use the `Connection` type from the `"ssh"` runner
> - we don't capture any results

### Using runner.New() and r.Run()

To test this, make a folder `~/Documents/Test` and put some files in it.

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
```

> Remark that in this example 
> - we are using our own `Connection` type
> - we are capturing results using a stdout-writer

### Using runner.New() and r.Start() / r.Wait()

To test this, make a folder `~/Documents/Test` and put some files in it.

```golang
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
        Path: "~\\Projects\\golang-exec\\examples\\new-and-start-wait",
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
    fmt.Printf(string(result))
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
```

> Remark that in this example 
> - we are using a map for our connection info
> - we are capturing results using a stdout-reader



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

    StdoutPipe() (io.Reader, error)   // don't use in combination with Run()
    StderrPipe() (io.Reader, error)   // don't use in combination with Run()

    Run() error
    Start() error
    Wait() error
    Close() error

    ExitCode() int   // -1 when runner error without completing script
}

func Run(connection interface {}, s *script.Script, arguments interface{}) error { /*...*/ }

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
    Type string   // must be "local"
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
    Type     string   // must be "ssh"
    Host     string
    Port     uint16
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

- support for setting environment variables
- support for SSH auth using certificates instead of password
- support for Pageant on Windows
- support for SSH bastion server
- support for WinRM communication