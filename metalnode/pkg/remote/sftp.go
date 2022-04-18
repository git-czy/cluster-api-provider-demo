package remote

import (
	"fmt"
	"io"
	"metalnode/utils/log"
	"os"
	"path"

	gosftp "github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
)

type sftp struct {
	sftpClient *gosftp.Client
	log        log.Logger
}

//NewSFTPClient 以ssh客户端为基础建立sftp连接客户端
func NewSFTPClient(sshClient *gossh.Client, log log.Logger) (*sftp, error) {
	sftpClient, err := gosftp.NewClient(sshClient)
	if err != nil {
		log.WithError(err).Errorln("Failed to new sftp")
		return nil, err
	}
	return &sftp{
		sftpClient: sftpClient,
		log:        log,
	}, nil
}

func (s *sftp) UploadFile(localFilePath string, remoteDirPath string) error {
	if s.sftpClient == nil {
		s.log.Error("Before run, have to new a sftp client")
		return nil
	}

	// 打开本地文件
	localFile, err := os.Open(localFilePath)
	defer localFile.Close()
	if err != nil {
		s.log.WithError(err).Errorln("Failed to open local file")
		return err
	}

	// 创建远程文件
	var remoteFileName = path.Base(localFilePath)
	remoteFile, err := s.sftpClient.Create(path.Join(remoteDirPath, remoteFileName))
	if err != nil {
		s.log.WithError(err).Errorln("Failed to create remote file")
		return err
	}
	defer remoteFile.Close()

	buf := make([]byte, 1024)
	for {
		n, err := localFile.Read(buf)
		if err != nil {
			if err != io.EOF {
				s.log.WithError(err).Errorln("Failed to read local file")
				return err
			} else {
				break
			}
		}
		if _, err := remoteFile.Write(buf[:n]); err != nil {
			s.log.WithError(err).Errorln("Failed to write buf to remote file")
			return err
		}
	}

	remoteFileStat, err := remoteFile.Stat()
	if err != nil {
		return err
	}
	localFileStat, err := localFile.Stat()
	if err != nil {
		return err
	}

	if remoteFileStat.Size() != localFileStat.Size() {
		s.log.Errorln("lost data when Upload File")
		if err := s.sftpClient.Remove(path.Join(remoteDirPath, remoteFileName)); err != nil {
			s.log.WithError(err).Errorln("remove damaged file Failed")
			return err
		}
		return fmt.Errorf("lost data when Upload File")
	}

	s.log.Infof("File %s successfully upload to %s", localFilePath, path.Join(remoteDirPath, remoteFileName))

	return nil

}
