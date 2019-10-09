//
// Copyright (c) 2019 Sean Reynolds, Stefaan Coussement
// MIT License
//
// more info: https://github.com/stefaanc/terraform-provider-hyperv
//
package ssh

import (
    "errors"
    "fmt"
    "github.com/mitchellh/go-homedir"
    "io"
    "golang.org/x/crypto/ssh/knownhosts"
    "log"
    "golang.org/x/crypto/ssh"

    "github.com/stefaanc/terraform-provider-hyperv/exec/script"
)

//------------------------------------------------------------------------------

type Connection struct {
    Type     string   // "ssh"
    Host     string
    Port     int16
    User     string
    Password string
    Insecure bool
}

type Runner struct {
    command string
    client  *ssh.Client
    session *ssh.Session
    running bool

    exitCode int
}

//------------------------------------------------------------------------------

func New(connection interface{}, s *script.Script, arguments interface{}) (*Runner, error) {
    if s.Error != nil {
        return nil, s.Error
    }

    c := connection.(Connection)
    r := new(Runner)
    r.command = s.Command()

    stdin, err := s.NewReader(arguments)
    if err != nil {
        r.exitCode = -1
        return nil, err
    }

    address := fmt.Sprintf("%s:%d", c.Host, c.Port)

    config := &ssh.ClientConfig{
        User: c.User,
        Auth: []ssh.AuthMethod{
            ssh.Password(c.Password),
        },
    }
    if c.Insecure {
        config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
    } else {
        f, err := homedir.Expand("~/.ssh/known_hosts")
        if err != nil {
            log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/ssh/New()] cannot find home directory of current user: %#v\n", err.Error())
            r.exitCode = -1
            return nil, err
        }

        hostKeyCallback, err := knownhosts.New(f)
        if err != nil {
            log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/ssh/New()] cannot access 'known_hosts'-file: %#v\n", err.Error())
            r.exitCode = -1
            return nil, err
        }
        config.HostKeyCallback = hostKeyCallback
    }

    client, err := ssh.Dial("tcp", address, config)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/ssh/New()] cannot dial host: %#v\n", err.Error())
        r.exitCode = -1
        return nil, err
    }
    r.client = client

    session, err := client.NewSession()
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/ssh/New()] cannot open session: %#v\n", err.Error())
        r.exitCode = -1
        return nil, err
    }
    r.session = session
    r.session.Stdin  = stdin

    return r, nil
}

//------------------------------------------------------------------------------

func (r *Runner) SetStdoutWriter(stdout io.Writer) {
    r.session.Stdout = stdout
}

func (r *Runner) SetStderrWriter(stderr io.Writer) {
    r.session.Stderr = stderr
}

func (r *Runner) StdoutPipe() (io.Reader, error) {
    reader, err := r.session.StdoutPipe()
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/ssh/StdoutPipe()] cannot create stdout reader: %#v\n", err.Error())
        r.exitCode = -1
    }

    return reader, err
}

func (r *Runner) StderrPipe() (io.Reader, error) {
    reader, err := r.session.StderrPipe()
    if err != nil {
       log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/ssh/StderrPipe()] cannot create stderr reader: %#v\n", err.Error())
        r.exitCode = -1
    }

    return reader, err
}

func (r *Runner) Start() error {
    err := r.session.Start(r.command)
    if err != nil {
        log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/ssh/Start()] cannot start runner: %#v\n", err.Error())
        r.exitCode = -1
        return err
    }
    r.running = true

    return nil
}

func (r *Runner) Wait() error {
    err := r.session.Wait()
    r.running = false
    if err != nil {
        var exitErr *ssh.ExitError
        if errors.As(err, &exitErr) {
            r.exitCode = exitErr.Waitmsg.ExitStatus()
        } else {
            r.exitCode = -1
        }

        log.Printf("[ERROR][terraform-provider-hyperv/exec/runner/ssh/Wait()] runner failed: %#v\n", err.Error())
        return err
    }

    r.exitCode = 0
    return nil
}

func (r *Runner) Close() error {
    if r.running {
        _ = r.session.Signal(ssh.SIGTERM)
    }

    if r.session != nil {
        _ = r.session.Close()
    }

    if r.client != nil {
        r.client.Close()
    }

    return nil
}

func (r *Runner) ExitCode() int {
    return r.exitCode
}

//------------------------------------------------------------------------------
