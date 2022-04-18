__set_config() {
  firewall-cmd --state
  systemctl stop firewalld.service
  systemctl disable firewalld.service
  setenforce 0
  sed -i 's,^SELINUX=.*$,SELINUX=disabled,' /etc/selinux/config
}
__set_config

__set_mirrors() {
  curl -o /etc/yum.repos.d/epel.repo http://mirrors.aliyun.com/repo/epel-7.repo
  curl -o /etc/yum.repos.d/CentOS-Base.repo https://mirrors.aliyun.com/repo/Centos-7.repo
  sed -i -e '/mirrors.cloud.aliyuncs.com/d' -e '/mirrors.aliyuncs.com/d' /etc/yum.repos.d/CentOS-Base.repo

  yum clean all
  yum makecache fast
  yum install -y sudo

}
__set_mirrors

__install_docker() {
  yum install -y yum-utils device-mapper-persistent-data lvm2
  yum-config-manager --add-repo https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
  sed -i 's+download.docker.com+mirrors.aliyun.com/docker-ce+' /etc/yum.repos.d/docker-ce.repo
  yum makecache fast
  yum -y install docker-ce

  usermod -aG docker root
  mkdir -p /etc/docker
  cat >/etc/docker/daemon.json <<EOF
{
    "registry-mirrors": [
        "https://mirror.ccs.tencentyun.com",
        "https://docker.mirrors.ustc.edu.cn"
    ],
    "exec-opts": ["native.cgroupdriver=systemd"]
}
EOF

  systemctl daemon-reload
  systemctl enable docker
  systemctl restart docker

}

__install_docker
