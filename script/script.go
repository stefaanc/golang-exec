//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/golang-exec
//
package script

import (
    "bytes"
    "fmt"
    "io"
    "strings"
    "text/template"
)

//------------------------------------------------------------------------------

type Script struct {
    Name       string

    shell      string   // "cmd", powershell", "bash", "sh", ...
    template   *template.Template

    Error      error    // error from NewScript()
}

//------------------------------------------------------------------------------

func New(name string, shell string, code string) *Script {
    // remark that New() doesn't return any errors directly
    // instead, error are saved in the 'Error'-field of the returned script
    // this allows using New() in a package scope, while checking for errors in a function scope
    template, err := template.New(name).Parse(code)
    if err != nil {
        err = fmt.Errorf("[golang-exec/script/New()] cannot parse script: %#w\n", err)
    }

    s := new(Script)
    s.Name = name

    if err == nil {
        s.shell = shell
        s.template = template
    } else {
        s.Error = err
    }

    return s
}

func NewFromString(name string, shell string, code string) (*Script, error) {
    template, err := template.New(name).Parse(code)
    if err != nil {
        return nil, fmt.Errorf("[golang-exec/script/NewFromString()] cannot parse script: %#w\n", err)
    }

    s := new(Script)
    s.Name = name
    s.shell = shell
    s.template = template

    return s, nil
}

func NewFromFile(name string, shell string, file string) (*Script, error) {
    template, err := template.New(name).ParseFiles(file)
    if err != nil {
        return nil, fmt.Errorf("[golang-exec/script/NewFromFile()] cannot parse script: %#w\n", err)
    }

    s := new(Script)
    s.Name = name
    s.shell = shell
    s.template = template

    return s, nil
}

//------------------------------------------------------------------------------

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

func (s *Script) NewReader(arguments interface{}) (io.Reader, error) {
    // returns a reader for the parsed & rendered script
    var rendered bytes.Buffer
    if s.template != nil {
        err := s.template.Execute(&rendered, arguments)
        if err != nil {
            return nil, fmt.Errorf("[golang-exec/script/NewReader()] cannot render script: %#w\n", err)
        }
    }

    return &rendered, nil
}

//------------------------------------------------------------------------------
