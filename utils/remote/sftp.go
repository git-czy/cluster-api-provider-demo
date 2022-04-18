package remote

import (
	"cluster-api-provider-demo/utils/log"
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
		n, _ := localFile.Read(buf)
		if n == 0 {
			break
		}
		if _, err := remoteFile.Write(buf); err != nil {
			s.log.WithError(err).Errorln("Failed to write buf to remote file")
			return err
		}
	}

	s.log.Infof("File %s successfully uploaded", localFilePath)

	return nil

}
