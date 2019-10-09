//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/golang-exec
//
package runner

import (
    "fmt"
    "io"
    "strings"

    "github.com/stefaanc/golang-exec/script"
    "github.com/stefaanc/golang-exec/runner/local"
    "github.com/stefaanc/golang-exec/runner/ssh"
)

//------------------------------------------------------------------------------

type Connection struct {
    Type string   // "local" or "ssh"
}

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

//------------------------------------------------------------------------------

func New(connection interface {}, s *script.Script, arguments interface{}) (Runner, error) {
    c := connection.(*Connection)
    switch strings.ToLower(c.Type) {
    case "local":
        return local.New(connection, s, arguments)
    case "ssh":
        return ssh.New(connection, s, arguments)
    default:
        return nil, fmt.Errorf("[ERROR][terraform-provider-hyperv/exec/runner/New()] invalid 'Type' in 'Connection'")
    }
}

//------------------------------------------------------------------------------
