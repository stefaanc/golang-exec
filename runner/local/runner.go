//
// Copyright (c) 2019 Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/golang-exec
//
package local

import (
    "context"
    "errors"
    "io"
    "log"
    "os/exec"
    "strings"

    "github.com/stefaanc/golang-exec/script"
)

//------------------------------------------------------------------------------

type Connection struct {
    Type string   // "local"
}

type Runner struct {
    cmd      *exec.Cmd
    cancel   context.CancelFunc

    exitCode int
}

//------------------------------------------------------------------------------

func New(connection interface{}, s *script.Script, arguments interface{}) (*Runner, error) {
    if s.Error != nil {
        return nil, s.Error
    }

    r := new(Runner)

    stdin, err := s.NewReader(arguments)
    if err != nil {
        r.exitCode = -1
        return nil, err
    }

    // create command, ready to start
    ctx, cancel := context.WithCancel(context.Background())
    args := strings.Split(s.Command(), " ")
    cmd := exec.CommandContext(ctx, args[0], args[1:]...)
    r.cmd = cmd
    r.cmd.Stdin  = stdin
    r.cancel = cancel

    return r, nil
}

//------------------------------------------------------------------------------

func (r *Runner) SetStdoutWriter(stdout io.Writer) {
    r.cmd.Stdout = stdout
}

func (r *Runner) SetStderrWriter(stderr io.Writer) {
    r.cmd.Stderr = stderr
}

func (r *Runner) StdoutPipe() (io.Reader, error) {
    reader, err := r.cmd.StdoutPipe()
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/local/StdoutPipe()] cannot create stdout reader: %#v\n", err.Error())
        r.exitCode = -1
    }

    return reader, err
}

func (r *Runner) StderrPipe() (io.Reader, error) {
    reader, err := r.cmd.StderrPipe()
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/local/StderrPipe()] cannot create stderr reader: %#v\n", err.Error())
        r.exitCode = -1
    }

    return reader, err
}

func (r *Runner) Run() error {
    err := r.cmd.Run()
    if err != nil {
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            r.exitCode = exitErr.ProcessState.ExitCode()
        } else {
            r.exitCode = -1
        }

        log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/local/Run()] runner failed: %#v\n", err.Error())
        return err
    }

    r.exitCode = 0
    return nil
}

func (r *Runner) Start() error {
    err := r.cmd.Start()
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/local/Start()] cannot start runner: %#v\n", err.Error())
        r.exitCode = -1
        return err
    }

    return nil
}

func (r *Runner) Wait() error {
    err := r.cmd.Wait()
    if err != nil {
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            r.exitCode = exitErr.ProcessState.ExitCode()
        } else {
            r.exitCode = -1
        }

        log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/local/Wait()] runner failed: %#v\n", err.Error())
        return err
    }

    r.exitCode = 0
    return nil
}

func (r *Runner) Close() error {
    if r.cancel != nil {
        r.cancel()
    }

    return nil
}

func (r *Runner) ExitCode() int {
    return r.exitCode
}

//------------------------------------------------------------------------------
