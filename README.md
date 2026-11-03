# IOT SDK Go

[English](./README_en.md) | 简体中文

IOT SDK Go 是用于开发 IOT 平台扩展服务的 Go SDK，覆盖 Driver、Algorithm、DataRelay、Flow、FlowExtension、Service、Task。

## 目录

- [模块概览](#模块概览)
- [安装](#安装)
- [快速开始](#快速开始)
- [核心接口](#核心接口)
- [各模块开发与部署](#各模块开发与部署)
- [Driver 配置示例](#driver-配置示例)
- [示例目录](#示例目录)
- [打包与部署](#打包与部署)
  - [授权动态库](#授权动态库)
- [FAQ](#faq)
- [环境要求](#环境要求)
- [许可证](#许可证)

## 模块概览

| 模块 | 包路径 | 说明 |
| --- | --- | --- |
| Driver | `github.com/kesiai/sdk-go/driver` | 设备接入、点位/事件/告警上报、指令执行 |
| Algorithm | `github.com/kesiai/sdk-go/algorithm` | 算法服务 |
| DataRelay | `github.com/kesiai/sdk-go/data_relay` | 数据中继 |
| Flow | `github.com/kesiai/sdk-go/flow` | 流程节点 |
| FlowExtension | `github.com/kesiai/sdk-go/flow_extension` | 可配置流程扩展 |
| Service | `github.com/kesiai/sdk-go/service` | Gin HTTP 服务 |
| Task | `github.com/kesiai/sdk-go/task` | Cron 定时任务 |

## 安装

```bash
go get github.com/kesiai/sdk-go
```

## 快速开始

```go
package main

import "github.com/kesiai/sdk-go/driver"

func main() {
	app := driver.NewApp()
	app.Start(yourDriver)
}
```

运行仓库示例：

```bash
go run ./example/driver -config ./example/driver/etc/
go run ./example/algorithm -config ./example/algorithm/etc/
go run ./example/data_relay -config ./example/data_relay/etc/
go run ./example/flow -config ./example/flow/etc/
go run ./example/flow_extension -config ./example/flow_extension/etc/
go run ./example/service
go run ./example/task
```

## 核心接口

### Driver

```go
type Driver interface {
	Schema(ctx context.Context, app App, locale string) (schema string, err error)
	Start(ctx context.Context, app App, driverConfig []byte) (err error)
	RegisterRoutes(router *gin.RouterGroup)
	Run(ctx context.Context, app App, command *entity.Command) (result interface{}, err error)
	BatchRun(ctx context.Context, app App, command *entity.BatchCommand) (result interface{}, err error)
	WriteTag(ctx context.Context, app App, command *entity.Command) (result interface{}, err error)
	Debug(ctx context.Context, app App, debugConfig []byte) (result interface{}, err error)
	HttpProxy(ctx context.Context, app App, t string, header http.Header, data []byte) (result interface{}, err error)
	ConfigUpdate(ctx context.Context, app App, data *pb.ConfigUpdateRequest) (err error)
	Stop(ctx context.Context, app App) (err error)
}
```

### Algorithm

```go
type Service interface {
	Schema(context.Context, App, string) (result string, err error)
	Start(context.Context, App) error
	Run(ctx context.Context, app App, bts []byte) (result interface{}, err error)
	Stop(context.Context, App) error
}
```

### DataRelay

```go
type DataRelay interface {
	Start(ctx context.Context, app App, config []byte) (err error)
	HttpProxy(ctx context.Context, app App, t string, header http.Header, data []byte) (result []byte, err error)
}
```

### Flow

```go
type Flow interface {
	Handler(ctx context.Context, app App, request *Request) (result map[string]interface{}, err error)
	Debug(ctx context.Context, app App, request *DebugRequest) (result *DebugResult, err error)
}
```

### FlowExtension

```go
type Extension interface {
	Schema(ctx context.Context, app App, locale string) (schema string, err error)
	Run(ctx context.Context, app App, input []byte) (result map[string]interface{}, err error)
}
```

说明：`flow_extension` 历史包名为 `flow_extionsion`，建议使用别名导入。

### Service

```go
type Service interface {
	Start(App) error
	Stop(App) error
}
```

在 `Start` 中通过 `app.GetHttpServer()` 注册 Gin 路由；`Stop` 用于退出前清理资源。

### Task

```go
type Task interface {
	Start(App) error
	Stop(App) error
}
```

在 `Start` 中通过 `app.GetCron()` 注册定时任务；SDK 启动调度器，并在退出时调用 `Stop`。

## 各模块开发与部署

各模块均以 `NewApp().Start(实现对象)` 启动。选择模块时，先确定平台如何调用它，以及程序是否需要监听入站端口：

| 模块 | 开发入口 | 主要配置 | 运行方式 |
| --- | --- | --- | --- |
| Driver | [设备驱动示例](./example/driver/main.go) | `driver.id`、`driver.name`、`driverGrpc.host/port` | 接入设备，连接平台驱动 gRPC 服务；打包步骤见下文 |
| Algorithm | [算法示例](./example/algorithm/main.go) | `algorithm.id/name`、`algorithmGrpc.host/port` | 实现 `Schema`、`Start`、`Run`、`Stop`，连接平台算法 gRPC 服务 |
| DataRelay | [数据中继示例](./example/data_relay/main.go) | `service.id/name`、`dataRelayGrpc.host/port` | 实现 `Start`、`HttpProxy`，连接平台数据中继 gRPC 服务 |
| Flow | [流程节点示例](./example/flow/main.go) | `flow.name`、`flow.mode`、`flowEngine.host/port` | 实现 `Handler`、`Debug`，由流程引擎调用节点逻辑 |
| FlowExtension | [流程扩展示例](./example/flow_extension/main.go) | `extension.id/name`、`flowEngine.host/port` | 实现 `Schema`、`Run`，向流程引擎提供可配置的扩展节点 |
| Service | [HTTP 服务示例](./example/service/main.go) | `server.port` | 实现 `Start`、`Stop`，自行监听 HTTP 端口 |
| Task | [定时任务示例](./example/task/main.go) | `log` 等应用配置 | 实现 `Start`、`Stop`，运行 Cron 任务；默认没有入站端口 |

Algorithm、DataRelay、Flow、FlowExtension 示例默认读取 `./etc/config.yaml`，也可用 `--config` 指定配置目录。表中的 gRPC `host/port` 是程序**主动连接**的平台地址，不是程序监听的端口。部署到容器时，应将这些地址改为容器可访问的平台服务地址。

Service、Task 示例读取当前工作目录的 `./etc/config.yaml`，不使用 `--config` 参数。上文从仓库根目录运行这两个示例时会使用 SDK 默认配置；要加载示例里的配置，请进入各自的 `example/service` 或 `example/task` 目录后执行 `go run .`。Service 的 `server.port` 是实际监听端口；Task 默认不监听端口。

发布到平台时，按目标运行模式准备 `service.yml`、可执行文件或 Docker 镜像，以及运行所需的 `etc/config.yaml`。驱动使用 `GroupName: driver`，数据中继使用 `GroupName: dataRelay`；算法、流程、流程扩展、HTTP 服务和定时任务作为普通服务安装时使用 `GroupName: server`。`Name` 填服务的安装包标识，原生包的 `Command` 指向实际可执行文件。容器包用 `Service: None` 表示没有入站服务；提供 HTTP 等入站接口时用 `Internal` 或 `External`，并填写 `Path` 和 `Ports`。这些字段与原生包的端口字段不同。

例如 HTTP Service 配置 `server.port: 9000` 后，Windows 原生包的 `service.yml` 可写为：

```yaml
Name: go-http-demo
Version: 1.0.0
GroupName: server
ConfigType: config.yaml
Command: go-http-demo.exe
Path: /go-http-demo
Port: 9000
```

Linux 容器包使用以下端口写法，并确保镜像内的 `etc/config.yaml` 同样配置 `server.port: 9000`：

```yaml
Name: go-http-demo
Version: 1.0.0
GroupName: server
Service: Internal
Path: /go-http-demo
Ports:
  - Host: "9000"
    Container: "9000"
    Protocol: ""
    AppProtocol: http
```

`Internal` 供平台内部访问；需要发布到宿主机时改为 `External`，并在 `Host` 填宿主机端口。Task 及仅主动连接平台、没有入站接口的模块无需 `Path` 和端口字段。下文的构建命令和文件结构以驱动为例，其他模块需替换入口、配置及服务分组；授权动态库说明仅适用于驱动的相应启动路径。

安装顺序按模块区分：

| 模块 | 发布与安装 |
| --- | --- |
| Driver | 先通过运维管理的**离线上传驱动**将安装包加入驱动仓库，再在项目中选择该驱动并安装或创建驱动实例。`GroupName` 为 `driver`。 |
| DataRelay | 先通过**离线上传数据中继服务**将安装包加入数据中继仓库，再为目标项目或实例安装数据中继服务。`GroupName` 为 `dataRelay`。 |
| Algorithm、Flow、FlowExtension、Service、Task | 按平台的普通服务安装入口部署；作为普通服务安装时，`GroupName` 为 `server`。 |

上传仅把 Driver 或 DataRelay 的安装包放入对应仓库；具体运行实例由后续安装步骤创建。界面名称和位置可能随平台版本变化。

## Driver 配置示例

```yaml
serviceId: your-service-id
groupId: your-group-id
project: your-project-id

driver:
  id: go-driver-demo
  name: Go Driver Demo

driverGrpc:
  enable: true
  host: localhost
  port: 9224
  health:
    requestTime: 10s
    retry: 3
  stream:
    heartbeat: 30s
  waitTime: 5s
  timeout: 600s
  limit: 100

http:
  enable: false
  host: 0.0.0.0
  port: 8080

dataFile:
  enable: true
  path: ./data.json

license: ./license

mq:
  type: mqtt
  mqtt:
    schema: tcp
    host: localhost
    port: 1883
    username: admin
    password: public

log:
  level: 4
  format: json
```

## 示例目录

- [example/driver](./example/driver)
- [example/driver_lazy](./example/driver_lazy)
- [example/algorithm](./example/algorithm)
- [example/data_relay](./example/data_relay)
- [example/flow](./example/flow)
- [example/flow_extension](./example/flow_extension)
- [example/service](./example/service)
- [example/task](./example/task)

## 打包与部署

本节介绍使用 Go SDK 开发的 kesi 驱动如何打包并部署到平台。可从 [`example/driver/main.go`](./example/driver/main.go) 开始；发布自己的驱动时，将示例中的 `go-driver-mqtt-demo` 换成实际驱动 ID，并保持程序名、`service.yml` 的 `Name`、`config.yaml` 的 `driver.id` 一致。原生包的 `Command` 还需指向实际可执行文件。

### 先确认目标环境

| 包类型 | 安装包根目录包含 | 如何启动 | 入站端口写在哪里 |
| --- | --- | --- | --- |
| Windows ZIP | `service.yml`、`.exe`、`etc/config.yaml` | `service.yml` 的 `Command` 指向 `.exe` | 程序的配置文件 |
| Linux Docker `.tar.gz` | `service.yml`、Docker 镜像归档 `.tar.gz` | 镜像的 `ENTRYPOINT` | `service.yml` 的 `Service`、`Path`、`Ports` |
| Linux/macOS binary `.tar.gz` | `service.yml`、可执行文件、`etc/config.yaml` | `service.yml` 的 `Command` 指向可执行文件 | 程序的配置文件 |

这里的 **binary 包**指直接运行可执行文件，**Docker 包**指平台导入并运行镜像。两者虽都以 `.tar.gz` 结尾，包内内容和启动方式不同，不能互换。按平台运行模式和目标 CPU 架构选择安装包。Go 版本至少为 `1.25`。

**平台 gRPC 模式是否需要授权动态库：**只通过 gRPC 接收平台的启动配置时不需要。保持 `dataFile.enable=false`（SDK 默认值），且不调用 SDK 的 HTTP 启动接口。即使已连接 gRPC，只要同时启用本地 `dataFile` 启动或调用 HTTP 启动接口，那条启动路径仍需要动态库。连接方式和安装包类型是两回事：Windows 原生包与 Linux 容器包都可以使用平台 gRPC 模式。

### 通用配置

驱动默认从当前工作目录的 `./etc/config.yaml` 读取配置，也可通过 `--config` 指定配置目录。至少配置 `driver.id`、`driver.name`。本地调试时可通过 `--project`、`--serviceId` 指定实例信息；平台安装并创建实例时由平台注入，发布包无需固定开发环境的实例 ID。

Windows 原生部署的 `etc/config.yaml` 可从以下内容开始，并按实际平台地址调整：

```yaml
driver:
  id: go-driver-mqtt-demo
  name: Go 驱动示例
driverGrpc:
  host: 127.0.0.1
  port: 9224
mq:
  type: mqtt
  mqtt:
    host: 127.0.0.1
    port: 1883
```

Linux 容器环境可参考 [`example/driver/etc/config.docker.yaml`](./example/driver/etc/config.docker.yaml)：只保留驱动身份字段时，SDK 使用容器环境的默认 `driverGrpc.host=driver`、`mq.mqtt.host=mqtt`。如启用 `dataFile.enable`，还需让 `dataFile.path` 指向容器或进程能访问的 `data.json`；默认值为 `false`。

### Windows 原生包

在仓库根目录使用 PowerShell 编译示例。自己的项目需把 `./example/driver` 换为实际 `main` 包路径。

```powershell
$env:CGO_ENABLED = '0'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
go build -tags netgo -o .\go-driver-mqtt-demo.exe ./example/driver
```

准备 `service.yml`：

```yaml
Name: go-driver-mqtt-demo
Version: 1.0.0
Description: Go 驱动示例
ConfigType: config.yaml
GroupName: driver
Command: go-driver-mqtt-demo.exe
```

把文件放在 ZIP 根目录，而不是多包一层项目目录：

```text
go-driver-mqtt-demo-windows-x86_64.zip
├── service.yml
├── go-driver-mqtt-demo.exe
├── etc/
│   └── config.yaml
└── lib/                         # 仅本地 dataFile / HTTP 启动模式需要
    └── license_core_windows_amd64.dll
```

在项目根目录准备 `stage/` 暂存目录，放入上述文件后执行 `Compress-Archive -Path .\stage\* -DestinationPath .\go-driver-mqtt-demo-windows-x86_64.zip`。自带配置页面、Schema 文件或其他资源的驱动，也需一并放入暂存目录。若发布环境要求 Windows 签名，应在压缩前完成。

### Linux 容器包

以下以 Linux amd64 为例，命令在项目根目录的 Bash 环境执行。arm64 将 `GOARCH` 改为 `arm64`、`--platform` 改为 `linux/arm64`，并使用对应架构的基础镜像；若选用需要授权动态库的启动路径，动态库架构也须匹配。loong64 需在支持该架构的构建环境编译。

```bash
mkdir -p build/driver-image/etc build/driver-package
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -o build/driver-image/go-driver-mqtt-demo ./example/driver
cp example/driver/etc/config.docker.yaml build/driver-image/etc/config.yaml
```

在 `build/driver-image/Dockerfile` 中放入：

```dockerfile
FROM debian:bookworm-slim
WORKDIR /app
COPY go-driver-mqtt-demo /app/go-driver-mqtt-demo
COPY etc/config.yaml /app/etc/config.yaml
ENTRYPOINT ["/app/go-driver-mqtt-demo"]
```

上述镜像用于平台 gRPC 启动模式，不需要授权动态库。若还启用 `dataFile` 本地启动或 SDK 的 HTTP 启动接口，构建前须把匹配目标架构和镜像运行时的 `license_core_linux_amd64.so` 放到 `build/driver-image/lib/`，并在 Dockerfile 中加上 `COPY lib/license_core_linux_amd64.so /app/lib/`。此处采用 Debian 基础镜像，需确认动态库及其依赖与镜像中的 libc 兼容。随后执行：

```bash
docker build --platform linux/amd64 -t gtsiot/go-driver-mqtt-demo:v1.0.0 build/driver-image
docker save -o build/driver-package/go-driver-mqtt-demo.tar gtsiot/go-driver-mqtt-demo:v1.0.0
gzip build/driver-package/go-driver-mqtt-demo.tar
```

在 `build/driver-package/service.yml` 中放入：

```yaml
Name: go-driver-mqtt-demo
Version: 1.0.0
Description: Go 驱动示例
GroupName: driver
Service: None
```

`Service` 可取 `None`、`Internal`、`External`。驱动只主动连接平台或设备、没有入站服务时使用 `None`。驱动提供 HTTP 接口供平台内部访问时，填写 `Internal`、访问路径 `Path` 和 `Ports`。例如驱动的 `config.yaml` 设置 `http.enable: true`、`http.port: 8080`，安装配置可写为：

```yaml
Name: go-driver-mqtt-demo
Version: 1.0.0
GroupName: driver
Service: Internal
Path: /go-driver-mqtt-demo
Ports:
  - Host: "8080"
    Container: "8080"
    Protocol: ""
    AppProtocol: http
```

需要把端口发布到宿主机时使用 `External`。以下示例在提供上述 HTTP 接口的同时，向宿主机发布驱动监听的 TCP `8558` 端口：

```yaml
Name: go-driver-mqtt-demo
Version: 1.0.0
GroupName: driver
Service: External
Path: /go-driver-mqtt-demo
Ports:
  - Host: "18080"
    Container: "8080"
    Protocol: ""
    AppProtocol: http
  - Host: "8558"
    Container: "8558"
    Protocol: ""
```

`Container` 是程序在容器内实际监听的端口；`Host` 是选择 `External` 时发布到宿主机的端口，两个端口可以不同。选择 `Internal` 时仍按上述格式填写 `Host`，平台不会将其发布到宿主机。`Protocol` 留空按 TCP 处理，UDP 端口填写 `udp`。有多个端口时，用 `AppProtocol: http` 标明用于 HTTP 路由的端口，避免把设备通信端口选为 HTTP 端口。`Path` 是平台转发 HTTP 请求的路径前缀。选择 `Internal` 或 `External` 时必须填写 `Ports`；容器内的监听端口还要与 `config.yaml` 一致。这里的 `Ports` 用于 Linux 容器包；SDK 用 `driverGrpc.host`、`driverGrpc.port` **主动连接**平台，这组连接参数不是需要发布的入站端口。

最后从**内层包目录**打包，确保外层归档根目录直接包含 `service.yml` 和镜像归档：

```bash
tar -C build/driver-package -czf go-driver-mqtt-demo-linux-x86_64.tar.gz service.yml go-driver-mqtt-demo.tar.gz
tar -tzf go-driver-mqtt-demo-linux-x86_64.tar.gz
```

上传到平台的是外层 `go-driver-mqtt-demo-linux-x86_64.tar.gz` 安装包；内层 `go-driver-mqtt-demo.tar.gz` 是 `docker save` 的镜像。安装包根目录需有 `service.yml` 和镜像文件，不能只上传内层镜像。

### Linux/macOS 原生包

仅在平台以原生进程运行驱动时使用。Linux amd64 示例先编译可执行文件：

```bash
mkdir -p build/driver-binary/etc
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -o build/driver-binary/go-driver-mqtt-demo ./example/driver
```

压缩前的目录结构：

```text
build/driver-binary/
├── service.yml
├── go-driver-mqtt-demo
├── etc/
│   └── config.yaml
└── lib/                         # 仅本地 dataFile / HTTP 启动模式需要
    └── license_core_linux_amd64.so
```

`service.yml` 与 Windows 版类似，但 `Command: go-driver-mqtt-demo`。`config.yaml` 使用目标环境的配置，不要原样发布示例中的开发用 `serviceId`、`project` 和平台地址。准备好上述文件后，从目录内生成归档：

```bash
tar -C build/driver-binary -czf go-driver-mqtt-demo-linux-x86_64-binary.tar.gz service.yml go-driver-mqtt-demo etc
```

`-C build/driver-binary` 只切换目录；后面的三个路径指定要打包的文件。归档根目录不会包含 `build/driver-binary` 这层目录。

需要 `lib/` 时，在命令末尾再加入 `lib`。macOS 使用 `GOOS=darwin`、相应 `GOARCH` 和 `.dylib`，并采用 `darwin-...-binary.tar.gz` 命名。原生包与容器包的安装入口由平台运行模式决定。

### 授权动态库

平台通过 gRPC 下发启动配置时，SDK 直接调用 `Driver.Start`，**不调用授权动态库**；只使用这一模式的包无需附带动态库。启用 `dataFile.enable=true` 并从本地 `data.json` 启动，或通过 SDK 的 HTTP 启动接口时，SDK 会在调用 `Driver.Start` 前通过 `license_core` 动态库校验。`license` 配置项是授权目录路径。

| 目标 | 默认库名 |
| --- | --- |
| Windows amd64 / arm64 | `license_core_windows_amd64.dll` / `license_core_windows_arm64.dll` |
| Linux amd64 / arm64 / loong64 | `license_core_linux_amd64.so` / `license_core_linux_arm64.so` / `license_core_linux_loong64.so` |
| macOS amd64 / arm64 | `license_core_darwin_amd64.dylib` / `license_core_darwin_arm64.dylib` |

需要动态库的模式，建议将其放在可执行文件旁的 `lib/` 目录；Linux/macOS 代码也搜索当前工作目录、`gtsiot/lib/driver/` 及可执行文件目录下的对应位置，Windows 还搜索 `license/lib/`。**这些模式没有动态库时，校验会报 `load driver license library failed`，驱动的 `Start` 不会成功**。如果 `dataFile.enable=true`，进程仍可能保持运行并打印错误，不能仅凭进程存在判断驱动已启动。

#### 下载与打包

使用本地 `dataFile` 或 SDK HTTP 启动接口时，按目标操作系统和 CPU 架构下载对应动态库：

| 目标平台 | 动态库下载地址 |
| --- | --- |
| Windows amd64 | `https://d.gtsiot.cn/driverjs/licenselib/license_core_windows_amd64.dll` |
| Linux amd64 | `https://d.gtsiot.cn/driverjs/licenselib/license_core_linux_amd64.so` |
| Linux arm64 | `https://d.gtsiot.cn/driverjs/licenselib/license_core_linux_arm64.so` |
| macOS arm64 | `https://d.gtsiot.cn/driverjs/licenselib/license_core_darwin_arm64.dylib` |

将动态库放入驱动项目的 `lib/` 目录。例如 Windows amd64 在项目根目录执行：

```powershell
New-Item -ItemType Directory -Force .\lib | Out-Null
Invoke-WebRequest 'https://d.gtsiot.cn/driverjs/licenselib/license_core_windows_amd64.dll' -OutFile '.\lib\license_core_windows_amd64.dll'
```

Linux amd64：

```bash
mkdir -p lib
curl -fL --retry 3 -o lib/license_core_linux_amd64.so https://d.gtsiot.cn/driverjs/licenselib/license_core_linux_amd64.so
```

下载后确认文件存在，再将 `lib/` 连同可执行文件放入原生安装包；容器包则按上文 Dockerfile 的可选 `COPY` 步骤放入镜像。需要上表未列出的 Windows arm64、Linux loong64、macOS amd64 动态库时，请联系平台交付方获取。只使用平台 gRPC 启动流时，跳过本节下载步骤。

### 安装与核验

1. 在运维管理系统选择**离线上传驱动**，上传**外层** ZIP 或 `tar.gz` 包，先将安装包加入驱动仓库；界面位置随平台版本变化。也可按本环境的发布流程将包放入离线仓库。
2. 确认 `service.yml` 的 `GroupName: driver`、`Name`、`Version` 和包内文件完整。Linux 容器包的 `service.yml` 必须位于归档根目录；Windows ZIP 同样建议直接放在根目录。
3. 在项目中选择已上传的驱动并安装或创建驱动实例。查看运维安装日志和驱动运行日志，确认镜像导入或进程启动成功，再检查连接、配置下发与数据上报。
4. 若安装失败，先检查包格式与架构、`service.yml` 的 `Command` 或 `Service`/`Ports`、配置文件路径。使用本地 `dataFile` 或 HTTP 启动路径时，再检查授权动态库是否存在及其依赖是否与运行环境兼容。

## FAQ

### 为什么 `healthRequestTime` 不生效？

请改用嵌套字段：

- `driverGrpc.health.requestTime`
- `driverGrpc.health.retry`

### 什么时候需要授权动态库？

仅使用平台 gRPC 启动流时不需要。使用本地 `dataFile` 或 SDK HTTP 启动接口时，需要匹配目标平台的 `license_core` 动态库；获取方式及部署位置见[授权动态库](#授权动态库)。

### `dataFile` 有什么作用？

用于本地驱动配置加载与热更新。启用后会读取并监听 `data.json`。

## 环境要求

- Go `>= 1.25`（以 `go.mod` 的 `go 1.25.0` 为准）

## 许可证

请参考仓库许可证文件。
