# 卸载旧版本

```
$ sudo yum remove docker \
                  docker-client \
                  docker-client-latest \
                  docker-common \
                  docker-latest \
                  docker-latest-logrotate \
                  docker-logrotate \
                  docker-engine
```

# 使用repo安装

在新主机上首次安装Docker Engine-Community之前，需要设置Docker存储库。之后，您可以从存储库安装和更新Docker。

## 设置yum repo

- 安装依赖

  ```
    $ sudo yum install -y yum-utils \
      device-mapper-persistent-data \
      lvm2   
  ```

- 使用以下命令来建立稳​​定的存储库

  ```
    $ sudo yum-config-manager \
        --add-repo \
        https://download.docker.com/linux/centos/docker-ce.repo 
  ```
  
  
## 安装docker-ce

- 安装最新版本的docker-ce
  
  ```
    $ sudo yum install docker-ce docker-ce-cli containerd.io
  ```
  
- 安装指定版本的docker-ce

$ yum list docker-ce --showduplicates | sort -r

    ```
        docker-ce.x86_64  3:18.09.1-3.el7                     docker-ce-stable
        docker-ce.x86_64  3:18.09.0-3.el7                     docker-ce-stable
        docker-ce.x86_64  18.06.1.ce-3.el7                    docker-ce-stable
        docker-ce.x86_64  18.06.0.ce-3.el7                    docker-ce-stable
    ```
- 启动docker
    ```
    $ sudo systemctl start docker
    ```
  
- 验证

    ```
$ sudo docker run hello-world
    ```
    
    
# 使用安装包安装

- 下载安装包

[地址](https://download.docker.com/linux/centos/7/x86_64/stable/Packages/)

- 安装

```
$ sudo yum install /path/to/package.rpm
```
    
- 启动
```
$ sudo systemctl start docker
```
    
- 验证

```
$ sudo docker run hello-world
```

# 卸载

- 卸载安装包

```
$ sudo yum remove docker-ce
```

- 删除数据卷及镜像，容器等

```
$ sudo rm -rf /var/lib/docker
```