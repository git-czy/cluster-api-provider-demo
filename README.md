### [cluster-api-provider-demo](https://github.com/git-czy/cluster-api-provider-demo)

#### 1.简介

- cluster-api-provider-demo包含 demoCluster CRD、demoMachine CRD
- cluster-api-provider-demo需要配合[cluster-api-metalnode](https://github.com/git-czy/cluster-api-metalnode)项目使用
- 主要实现 cluster-api Infrastructure逻辑

#### 2.部署

##### 2.1.部署前的准备

1. 准备一台机器安装kind，kubectl
2. 确保已经安装clusterctl
3. 使用kind拉起一个集群

cluster-api环境配置可参考init_cluster_api_env.sh脚本

##### 2.2.开始部署

1. 部署[cluster-api-metalnode](https://github.com/git-czy/cluster-api-metalnode)项目

2. 初始化使用clusterctl 初始化cluster-api

   ```
   cluster init
   ```

3. 下载项目代码到您本地，并进入项目目录

4. 执行make run可在集群外运行项目

5. 执行make deploy将controller部署到集群

   1. 如果部署失败，请提前下载一下镜像 使用kind导入到集群

      ```
      # iamges
      ccr.ccs.tencentyun.com/oldcc/cluster-api-provider-demo:latest
      
      kind load docker-image ccr.ccs.tencentyun.com/oldcc/cluster-api-provider-demo:latest
      ```

6. 发布集群

   ```
   kubectl apply -f config/samples/demo-cluster.yaml
   ```

7. 查看集群

   ```
   clusterctl describe cluster demo-cluster --namespace demo-cluster
   ```

   



