# mDocker介绍

## mdocker是什么?

正如项目名称那样, 这是一个mini版的docker实现, 实现参考了mydocker:

[mydocker仓库地址](https://github.com/xianlubird/mydocker)

本项目的目的是熟悉docker的实现原理, 而不将其作作为生产工具使用。

实现中涉及使用到的Linux技术在我的博客中发表了相关的系列文章, 希望能给大家学习这个项目带来帮助。

[之深海](http://toseafloor.xyz)

## 项目环境准备

在编译mDocker之前, 你需要准备的环境有:

1. go version >= 1.17
2. ubuntu >= ubuntu 18.04, docker依赖的linux机制意味着你必须拥有一个linux环境


## 编译二进制文件

在项目根目录下, 执行以下命令进行编译:
```shell
go mod tidy
mkdir ./bin
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o mdocker entry.go
```


# mdocker用法

## mdocker load/save - 准备镜像

mdocker可以运行docker镜像, 比如我们将docker的busybox镜像导出:
```shell
docker pull busybox:latest
docker run -d busybox top
docker export -o busybox.tar [刚才启动容器id] 
```

将busybox镜像导入mdocker:
```shell
mdocker load ./busybox.tar busybox
```

## mdocker run - 启动一个容器

创建一个sh交互式容器:
```shell
mdocker run -ti -name sample busybox sh
```

## mdocker network 

创建一个bridge网络:
```shell
mdocker network create --driver bridge --subnet 192.168.10.0/24 mdocker0
```
创建容器时加入指定网络:
```shell
mdocker run -ti -name sample -net mdocker0 busybox sh
```