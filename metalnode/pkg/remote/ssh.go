package remote

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"time"

	gossh "golang.org/x/crypto/ssh"

	"metalnode/utils/log"
)

type ssh struct {
	sshClient *gossh.Client
	log       log.Logger
}

func NewSSHClient(user string, password string, host string, port int, sshKey string, log log.Logger) (*ssh, error) {
	var (
		sshClient *gossh.Client
		err       error
	)

	switch {
	case user != "" && password != "" && host != "":
		if sshClient, err = NewNormalSSHClient(user, password, host, port); err != nil {
			return nil, err
		}
		break
	case sshKey != "":
		if sshClient, err = NewWithOutPassSSHClient(sshKey); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("some fields are blank")
	}

	return &ssh{
		sshClient: sshClient,
		log:       log,
	}, nil
}

// NewNormalSSHClient 使用账号密码创建ssh客户端
func NewNormalSSHClient(user string, password string, host string, port int) (*gossh.Client, error) {
	config := &gossh.ClientConfig{
		User:            user,
		Auth:            []gossh.AuthMethod{gossh.Password(password)},
		Timeout:         30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil },
	}

	address := fmt.Sprintf("%s:%d", host, port)

	client, err := gossh.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewWithOutPassSSHClient 使用sshKey创建ssh客户端
func NewWithOutPassSSHClient(sshKey string) (*gossh.Client, error) {
	// todo
	return nil, nil
}

// Exec 执行shell命令
func (s *ssh) Exec(cmd string) ([]string, error) {
	var stderrs []string
	l := s.log.With("command", cmd)

	if s.sshClient == nil {
		l.Error("Before run, have to new a ssh client")
		return stderrs, nil
	}

	// 不执行命令直接返回
	if cmd == "" {
		return stderrs, nil
	}

	session, err := s.sshClient.NewSession()
	if err != nil {
		return stderrs, err
	}

	defer func(session *gossh.Session) {
		err := session.Close()
		if err != nil && err != io.EOF {
			l.WithError(err).Infoln("Some errors happened when ssh client session closed")
		}
	}(session)

	r, _ := session.StdoutPipe()
	e, _ := session.StderrPipe()

	go func() {
		err := session.Run(cmd)
		if err != nil {
			l.WithError(err).Errorln("run command failed")
			return
		}
	}()

	reader1 := bufio.NewReader(r)
	reader2 := bufio.NewReader(e)

	for {
		if err := readStdoutPipe(reader1, l); err != nil {
			if err == io.EOF {
				return stderrs, nil
			}
			return stderrs, err
		}
		stderr, err := readStderrPipe(reader2, l)
		if stderr != "" {
			stderrs = append(stderrs, stderr)
		}
		if err != nil {
			if err == io.EOF {
				return stderrs, nil
			}
			return stderrs, err
		}

	}
}

func readStdoutPipe(reader *bufio.Reader, log log.Logger) error {
	line, _, err := reader.ReadLine()
	if err == io.EOF {
		return err
	}
	if err != nil && err != io.EOF {
		log.WithError(err).Errorln("Read pipe buffer failed")
		return err
	}
	log.Debugln(string(line))
	return nil
}

func readStderrPipe(reader *bufio.Reader, log log.Logger) (string, error) {
	line, _, err := reader.ReadLine()
	if err == io.EOF {
		return "", err
	}
	if err != nil && err != io.EOF {
		log.WithError(err).Errorln("Read pipe buffer failed")
		return "", err
	}
	stderr := string(line)
	log.Errorln(stderr)
	return stderr, nil
}
