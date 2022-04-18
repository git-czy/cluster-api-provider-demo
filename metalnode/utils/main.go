package utils

import (
	"fmt"
	remote2 "metalnode/pkg/remote"
)

func main() {
	host := []remote2.Host{
		{
			User:     "centos",
			Password: "Ccc51521!",
			Address:  "10.20.9.148",
			Port:     22,
			SSHKey:   "",
		},
	}

	cmd := remote2.Cmd{
		Cmds: []string{
			//"sudo chmod +x /tmp/init_k8s_env.sh",
			//"sudo /tmp/init_k8s_env.sh",
			//"sudo docker run hello-word",
			"sudo gpasswd -a $USER docker",
			"newgrp docker",
			"docker version",
		},
		//FileUp: []remote.File{
		//	{Src: "script/init_k8s_env.sh", Dst: "/tmp"},
		//},
	}
	errs := remote2.Run(host, cmd)
	fmt.Println(errs)
}
