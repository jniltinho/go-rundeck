package service

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHResult holds the output of a remote command execution.
type SSHResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// SSHService executes commands on remote nodes via SSH.
type SSHService struct {
	connectTimeout time.Duration
}

// NewSSHService creates a new SSHService.
func NewSSHService(connectTimeoutSec int) *SSHService {
	if connectTimeoutSec <= 0 {
		connectTimeoutSec = 10
	}
	return &SSHService{
		connectTimeout: time.Duration(connectTimeoutSec) * time.Second,
	}
}

// RunCommand connects to the host via SSH and executes cmd, returning the result.
func (s *SSHService) RunCommand(host string, port int, user string, authMethod ssh.AuthMethod, cmd string) (*SSHResult, error) {
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         s.connectTimeout,
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, s.connectTimeout)
	if err != nil {
		return nil, fmt.Errorf("tcp dial %s: %w", addr, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, fmt.Errorf("ssh handshake %s: %w", addr, err)
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	return s.runSession(client, cmd)
}

// RunCommandWithPassword connects using a plain-text password.
func (s *SSHService) RunCommandWithPassword(host string, port int, user, password, cmd string) (*SSHResult, error) {
	return s.RunCommand(host, port, user, ssh.Password(password), cmd)
}

// RunCommandWithKey connects using a PEM-encoded private key.
func (s *SSHService) RunCommandWithKey(host string, port int, user string, pemKey []byte, cmd string) (*SSHResult, error) {
	signer, err := ssh.ParsePrivateKey(pemKey)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return s.RunCommand(host, port, user, ssh.PublicKeys(signer), cmd)
}

func (s *SSHService) runSession(client *ssh.Client, cmd string) (*SSHResult, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	exitCode := 0
	if err := session.Run(cmd); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			return nil, fmt.Errorf("run command: %w", err)
		}
	}

	return &SSHResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}

// TestConnection attempts to open an SSH connection and returns nil on success.
func (s *SSHService) TestConnection(host string, port int, user string, authMethod ssh.AuthMethod) error {
	_, err := s.RunCommand(host, port, user, authMethod, "echo ok")
	return err
}

// RunCommandWithPasswordDebug is a copy of RunCommandWithPassword with verbose step-by-step logging.
func (s *SSHService) RunCommandWithPasswordDebug(host string, port int, user, password, cmd string, logf func(string, ...any)) (*SSHResult, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	logf("[1] target: %s@%s  password_len=%d", user, addr, len(password))

	// Step 2: TCP dial
	logf("[2] TCP dial %s (timeout=%s)...", addr, s.connectTimeout)
	conn, err := net.DialTimeout("tcp", addr, s.connectTimeout)
	if err != nil {
		logf("[2] FAIL tcp dial: %v", err)
		return nil, fmt.Errorf("tcp dial %s: %w", addr, err)
	}
	logf("[2] OK tcp connected")

	// Step 3: build SSH client config
	logf("[3] building SSH config: auth=[password, keyboard-interactive]")
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			ssh.KeyboardInteractive(func(u, instruction string, questions []string, echos []bool) ([]string, error) {
				logf("[3] keyboard-interactive challenge: user=%q instruction=%q questions=%d", u, instruction, len(questions))
				answers := make([]string, len(questions))
				for i, q := range questions {
					logf("[3]   question[%d]: %q", i, q)
					answers[i] = password
				}
				logf("[3] keyboard-interactive: returning %d answers", len(answers))
				return answers, nil
			}),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         s.connectTimeout,
	}

	// Step 4: SSH handshake
	logf("[4] starting SSH handshake...")
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		logf("[4] FAIL handshake: %v", err)
		return nil, fmt.Errorf("ssh handshake %s: %w", addr, err)
	}
	logf("[4] OK handshake — server version: %s", sshConn.ServerVersion())

	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	// Step 5: open session
	logf("[5] opening session...")
	session, err := client.NewSession()
	if err != nil {
		logf("[5] FAIL new session: %v", err)
		return nil, fmt.Errorf("new session: %w", err)
	}
	defer session.Close()
	logf("[5] OK session opened")

	// Step 6: run command
	logf("[6] running command: %s", cmd)
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	exitCode := 0
	if err := session.Run(cmd); err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
			logf("[6] command exited with code %d", exitCode)
		} else {
			logf("[6] FAIL run: %v", err)
			return nil, fmt.Errorf("run command: %w", err)
		}
	} else {
		logf("[6] OK exit code 0")
	}

	logf("[7] stdout_len=%d  stderr_len=%d", stdout.Len(), stderr.Len())
	return &SSHResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
