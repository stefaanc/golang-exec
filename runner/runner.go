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
    "reflect"
    "strings"

    "github.com/stefaanc/golang-exec/script"
    "github.com/stefaanc/golang-exec/runner/local"
    "github.com/stefaanc/golang-exec/runner/ssh"
)

//------------------------------------------------------------------------------

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

func Run(connection interface {}, s *script.Script, arguments interface{}) error {
    if s.Error != nil {
        return s.Error
    }

    r, err := New(connection, s, arguments)
    if err != nil {
        return err
    }
    defer r.Close()

    err = r.Run()
    if err != nil {
        return err
    }

    return nil
}

func New(connection interface {}, s *script.Script, arguments interface{}) (Runner, error) {
    if s.Error != nil {
        return nil, s.Error
    }

    v := reflect.Indirect(reflect.ValueOf(connection))
    t := reflect.TypeOf(connection)

    var cType string
    if t.Kind() == reflect.Struct {
        cType = strings.ToLower(v.FieldByName("Type").String())
    } else {
        iter := v.MapRange()
        for iter.Next() {
            if iter.Key().String() == "Type" {
                cType = strings.ToLower(iter.Value().String())
                break
            }
        }
    }

    switch cType {
    case "local":
        return local.New(connection, s, arguments)
    case "ssh":
        return ssh.New(connection, s, arguments)
    default:
        return nil, fmt.Errorf("[golang-exec/runner/New()] invalid 'Type' in 'connection' parameter")
    }
}

//------------------------------------------------------------------------------
