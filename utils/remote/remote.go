package remote

import (
	"cluster-api-provider-demo/utils/log"
	"io"
)

type CLi struct {
	User     string
	Password string
	SSHKey   string
	Address  string
	Port     int
	SSH      *ssh
	SFTP     *sftp
	log      log.Logger
}

func (c CLi) Fields() (string, string, string, int, string, log.Logger) {
	return c.User, c.Password, c.Address, c.Port, c.SSHKey, c.log
}

func Run(hosts []Host, cmd Cmd) {
	stopChan := make(chan bool)

	var q = 0
	for _, h := range hosts {
		go run(h, cmd, stopChan)
	}

	for {
		if <-stopChan {
			q += 1
		}
		if q == len(hosts) {
			return
		}
	}
}

func run(h Host, cmd Cmd, stopChan chan bool) {
	RemoteClient, err := NewRemoteClient(&h)

	if err != nil || RemoteClient == nil {
		stopChan <- true
		return
	}

	defer RemoteClient.CloseRemoteCli(stopChan)

	for _, file := range cmd.FileUp {
		if err := RemoteClient.SFTP.UploadFile(file.Src, file.Dst); err != nil {
			return
		}
	}

	if err = RemoteClient.SSH.Exec(cmd); err != nil {
		return
	}

}

// NewRemoteClient 新建远程客户端
func NewRemoteClient(h *Host) (*CLi, error) {
	var err error

	h, err = h.Validate()
	if err != nil {
		log.WithError(err).Errorln("host validation failed")
		return nil, err
	}

	l := log.With("host", h.Address).With("user", h.User)

	if err = l.SetLevel("debug"); err != nil {
		return nil, err
	}

	l.Info("New RemoteClient ....")

	c := &CLi{
		User:     h.User,
		Password: h.Password,
		SSHKey:   h.SSHKey,
		Address:  h.Address,
		Port:     h.Port,
		log:      l,
	}

	c.SSH, err = NewSSHClient(c.Fields())
	if err != nil {
		c.log.WithError(err).Errorf("Failed to create ssh client")
		return nil, err
	}

	c.log.Info("ssh client connected")

	c.SFTP, err = NewSFTPClient(c.SSH.sshClient, c.log)
	if err != nil {
		c.log.WithError(err).Errorf("Failed to create sftp client")
		return nil, err
	}

	c.log.Info("sftp client connected")

	return c, nil
}

// CloseRemoteCli 关闭远程客户端
func (c *CLi) CloseRemoteCli(stopChan chan bool) {

	if err := c.SFTP.sftpClient.Close(); err != nil && err != io.EOF {
		c.SFTP.log.WithError(err).Infoln("Some errors happened when sftp client closed")
		return
	}

	stopChan <- true
	c.log.Infoln("RemoteCli closed")
}
