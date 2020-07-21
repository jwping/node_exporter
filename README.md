# Node exporter

## 1. 简述

该项目是在官方主分支`202ecf9`基础上进行定制，主要为我们从open-falcon迁移到Prometheus监控平台中端口以及URL监控策略适应而进行的二次开发。

## 2. 安装

### 2.1 预编译版本

对于发布版本预编译的二进制是可用的，下载方式如下，可参考[releases](https://github.com/jwping/prometheus/releases)

```shell
wget https://github.com/jwping/node_exporter/releases/download/v0.0.1/node_exporter

---or---

curl -o node_exporter https://github.com/jwping/node_exporter/releases/download/v0.0.1/node_exporter

./node_exporter <flags>
```

### 2.2 源码编译安装

#### 2.2.1 前置条件：

* GO环境
* RHEL/CentOS: `glibc-static` 依赖.

#### 2.2.2 编译

这里编译可以直接`go build`或者是使用官方的`make`

> 注意，如果使用官方的方式来进行编译的话，需要安装有`golangci-lint`

```shell
go get github.com/jwping/node_exporter
cd ${GOPATH-$HOME/go}/src/github.com/prometheus/node_exporter

# go build
# make

./node_exporter <flags>
```

#### 2.2.3 运行测试

```shell
make test
```

### 2.3 Docker部署

该`node_exporter`设计用于监控主机系统。不建议将其部署为Docker容器，因为它需要访问主机系统。请注意，您要监视的所有非root挂载点都需要绑定挂载到容器中。如果启动用于宿主机监视的容器，请指定`path.rootfs`参数。此参数必须与主机`/`中挂载安装的路径匹配。node_exporter将 `path.rootfs`用作访问主机文件系统的前缀。

```shell
docker run -d \
  --net="host" \
  --pid="host" \
  -v "/:/host:ro,rslave" \
  quay.io/prometheus/node-exporter \
  --path.rootfs=/host
```

在某些系统上，`timex`收集器需要附加的Docker标志: `--cap-add=SYS_TIME`才能访问所需的syscall。

## 3. 新增功能

### 3.1 支持端口连通性采集

> 新增collector/port_checking.go

需要在Prometheus上配置params：

```shell
params:
  portlist:
    - 127.0.0.1:22
    - :9100
```

上列配置相当于告诉Prometheus在请求该Target时附带URL参数，如下：

```shell
http://192.168.14.130:9100/metrics?portlist=127.0.0.1:22&&portlist=:9100
```

node_export接收到附加`portlist`参数的请求后，会使用`probeTCP`函数对目标端口的连通性进行采集，具体源码实现请参考[collector/port_checking.go#L29](https://github.com/jwping/node_exporter/blob/master/collector/port_checking.go#L29)行。

### 3.2 支持GET请求状态码和请求耗时采集

> 新增http_checking.go

基本同上，需要在Prometheus上配置params：

```shell
params:
  httplist:
    - https://www.baidu.com
    - https://aliyun.com
```

上列配置相当于告诉Prometheus在请求该Target时附带URL参数，如下：

```shell
http://192.168.14.130:9100/metrics?httplist=https://www.baidu.com&&httplist=https://aliyun.com
```

node_export接收到附加`httplist`参数的请求后，会使用`probeHTTP`函数对目标端口的连通性进行采集，具体源码实现请参考[collector/http_checking.go#L29](https://github.com/jwping/node_exporter/blob/master/collector/http_checking.go#L29)行。

### 3.3 修改node_export.go以支持上述修改

```shell
node_export.go

// L75新增url参数解析，用于http、port参数获取
```

### 3.4 返回数据

```shell
...
# HELP port_connectivity_detection Running on each node
# TYPE port_connectivity_detection gauge
port_connectivity_detection{port="127.0.0.1:22"} 1
port_connectivity_detection{port=":9100"} 1
...
# HELP http_connectivity_detection Running on each node
# TYPE http_connectivity_detection gauge
http_connectivity_detection{http="https://aliyun.com",httpStatusCode="200"} 399
http_connectivity_detection{http="https://www.baidu.com",httpStatusCode="200"} 130
...
```

**另请注意，URL请求的超时时间为3秒，返回数据的耗时单位为毫秒**

