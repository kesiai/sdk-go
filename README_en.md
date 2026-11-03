# IOT SDK Go

English | [Chinese](./README.md)

IOT SDK Go is the Go SDK for building IOT extension services, including Driver, Algorithm, DataRelay, Flow, FlowExtension, Service, and Task modules.

Source: [kesiai/sdk-go](https://github.com/kesiai/sdk-go). Published versions are listed under repository Tags.

## Table of Contents

- [Module Overview](#module-overview)
- [Install](#install)
- [Quick Start](#quick-start)
- [Core Interfaces](#core-interfaces)
- [Developing and Deploying Modules](#developing-and-deploying-modules)
- [Driver Configuration Example](#driver-configuration-example)
- [Configuration and Runtime Conventions](#configuration-and-runtime-conventions)
- [Driver Reporting and Management APIs](#driver-reporting-and-management-apis)
- [Development Checks](#development-checks)
- [Example Projects](#example-projects)
- [Packaging and Deployment](#packaging-and-deployment)
  - [License Library](#license-library)
- [FAQ](#faq)
- [Requirements](#requirements)
- [License](#license)

## Module Overview

| Module | Package | Description |
| --- | --- | --- |
| Driver | `github.com/kesiai/sdk-go/driver` | Device access, point/event/warning reporting, command handling |
| Algorithm | `github.com/kesiai/sdk-go/algorithm` | Algorithm service |
| DataRelay | `github.com/kesiai/sdk-go/data_relay` | Data relay and proxy |
| Flow | `github.com/kesiai/sdk-go/flow` | Flow node logic |
| FlowExtension | `github.com/kesiai/sdk-go/flow_extension` | Configurable flow extension |
| Service | `github.com/kesiai/sdk-go/service` | Gin HTTP service |
| Task | `github.com/kesiai/sdk-go/task` | Cron task runner |

## Install

```bash
go get github.com/kesiai/sdk-go
```

## Quick Start

```go
package main

import (
	"github.com/kesiai/sdk-go/driver"
	exampledriver "github.com/kesiai/sdk-go/example/driver/app"
)

func main() {
	app := driver.NewApp()
	app.Start(new(exampledriver.TestDriver))
}
```

This uses the repository's MQTT example implementation. For your own driver, implement the complete `driver.Driver` interface below. First obtain the source and dependencies:

```bash
git clone https://github.com/kesiai/sdk-go.git
cd sdk-go
go mod download
```

For an initial check without platform services, run the HTTP Service example in a separate terminal:

```bash
cd example/service
go run .
```

Visit `http://localhost:9000/` for `hello world`; stop with `Ctrl+C`. For Task, change into `example/task` and run `go run .`. Its Cron expressions include seconds and have six fields.

Run the other examples from the repository root after preparing reachable platform services and configuration:

```bash
go run ./example/algorithm -config ./example/algorithm/etc/
go run ./example/data_relay -config ./example/data_relay/etc/
go run ./example/flow -config ./example/flow/etc/
go run ./example/flow_extension -config ./example/flow_extension/etc/
```

## Core Interfaces

These are per-module interface excerpts. `App` refers to that module's App, and dependent types require the imports shown in source; do not combine all snippets into one Go file. See the module paths above and example directories below for parameter types and complete implementations.

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

Note: historical package name is `flow_extionsion`, so alias import is recommended.

### Service

```go
type Service interface {
	Start(App) error
	Stop(App) error
}
```

Register Gin routes through `app.GetHttpServer()` in `Start`; use `Stop` to clean up resources before shutdown.

### Task

```go
type Task interface {
	Start(App) error
	Stop(App) error
}
```

Register scheduled jobs through `app.GetCron()` in `Start`. The SDK starts the scheduler and calls `Stop` during shutdown.

## Developing and Deploying Modules

Each module starts with `NewApp().Start(implementation)`. Choose a module based on how the platform invokes it and whether the process needs an incoming port:

| Module | Starting point | Main configuration | Runtime behavior |
| --- | --- | --- | --- |
| Driver | [Driver example](./example/driver/main.go) | `driver.id`, `driver.name`, `driverGrpc.host/port` | Connects devices and the platform driver gRPC service; packaging steps follow below |
| Algorithm | [Algorithm example](./example/algorithm/main.go) | `algorithm.id/name`, `algorithmGrpc.host/port` | Implements `Schema`, `Start`, `Run`, `Stop`; connects to the platform algorithm gRPC service |
| DataRelay | [Data relay example](./example/data_relay/main.go) | `service.id/name`, `dataRelayGrpc.host/port` | Implements `Start`, `HttpProxy`; connects to the platform data relay gRPC service |
| Flow | [Flow node example](./example/flow/main.go) | `flow.name`, `flow.mode`, `flowEngine.host/port` | Implements `Handler`, `Debug`; the flow engine invokes its node logic |
| FlowExtension | [Flow extension example](./example/flow_extension/main.go) | `extension.id/name`, `flowEngine.host/port` | Implements `Schema`, `Run` to provide a configurable node to the flow engine |
| Service | [HTTP service example](./example/service/main.go) | `server.port` | Implements `Start`, `Stop` and listens on an HTTP port |
| Task | [Scheduled task example](./example/task/main.go) | Application settings such as `log` | Implements `Start`, `Stop` and runs Cron jobs; no incoming port by default |

The Algorithm, DataRelay, Flow, and FlowExtension examples read `./etc/config.yaml` by default; use `--config` to select another config directory. Their gRPC `host/port` settings are **outbound** platform addresses, not listening ports. In a container, set them to platform service addresses reachable from that container.

Service and Task read `./etc/config.yaml` from the working directory during package initialization and do not accept `--config`. The file is required; a missing file causes a panic. Change into `example/service` or `example/task` and run `go run .`. Service listens on `server.port`; Task has no listening port by default.

For platform deployment, package `service.yml`, the executable or Docker image, and any required `etc/config.yaml` for the target runtime. Use `GroupName: driver` for drivers and `GroupName: dataRelay` for data relays. Use `GroupName: server` when installing Algorithm, Flow, FlowExtension, HTTP Service, or Task as a regular service. Set `Name` to the installation package's service identifier and point `Command` to the executable in a native package. In a container package, `Service: None` means there is no incoming service. For an incoming HTTP endpoint, use `Internal` or `External` with `Path` and `Ports`. Native packages use different port fields.

For example, with HTTP Service configured as `server.port: 9000`, a native Windows `service.yml` can contain:

```yaml
Name: go-http-demo
Version: 1.0.0
GroupName: server
ConfigType: config.yaml
Command: go-http-demo.exe
Path: /go-http-demo
Port: 9000
```

For a Linux container package, use this port configuration and set `server.port: 9000` in the image's `etc/config.yaml`:

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

`Internal` makes the endpoint available within the platform. To publish it on the host, use `External` and set `Host` to the host port. Task and other modules that only connect outward need no `Path` or port fields. The build commands and file layouts below use a driver as the example; replace the entry point, configuration, and service group for other modules. The license library guidance applies only to the relevant driver startup paths.

The installation sequence depends on the module:

| Module | Publish and install |
| --- | --- |
| Driver | First use **offline driver upload** in operations management to add the package to the driver repository. Then select the driver in a project and install or create a driver instance. Use `GroupName: driver`. |
| DataRelay | First use **offline data relay upload** to add the package to the data relay repository. Then install the data relay service for the target project or instance. Use `GroupName: dataRelay`. |
| Algorithm, Flow, FlowExtension, Service, Task | Use the platform's regular service installation entry. Use `GroupName: server` when installing as a regular service. |

Uploading a Driver or DataRelay package only adds it to its repository; the later installation step creates the running instance. UI names and locations may vary by platform version.

## Driver Configuration Example

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
  enable: false
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

This configuration uses platform gRPC startup. Replace the placeholders: `driver.id`, `project`, and `serviceId` must match the platform driver and running instance. Local file startup and licensing requirements are described below.

## Configuration and Runtime Conventions

- Driver, Algorithm, DataRelay, Flow, and FlowExtension use `--config <directory>` to select `config.yaml`. This does not change the working directory; `dataFile.path`, `license`, and resource paths remain relative to it.
- Use one module App per executable. Configuration uses global Viper/pflag state; do not repeatedly construct multiple Apps in one process.
- Driver and DataRelay support `mq.type` values `mqtt`, `rabbit`, `kafka`, and `local`. Local MQ stays inside the process, does not report to the platform, and does not disable the driver's device connections.
- Driver's `mq.timeout`, `driverGrpc.waitTime/timeout`, `health.requestTime`, and `stream.heartbeat` use Go duration strings such as `5s` and `1m`. Algorithm's `algorithmGrpc.waitTime`, `health.requestTime`, and `algorithm.timeout` use integer seconds; do not copy Driver duration strings into them.
- DataRelay requires `project`, `instanceId`, and `service.id/name`. Its current `NewApp()` reads etcd at `etcdConfig` even with `api.liteMode=true`; configure reachable etcd and API services.
- Encrypted configuration uses `CONFIG_CIPHER_KEY`; see the [decrypt implementation](./utils/decrypt/config.go) for formats. Legacy `AES(...)` uses GCM in Go and CBC in Node.js, so the values cannot be exchanged directly. `ENC(...)` is Base64 encoding.

Run the local MQTT Driver example with `go run .` from `example/driver`. Its checked-in configuration enables `dataFile` and HTTP and disables platform gRPC. Prepare `data.json`, the target platform's license library, and the device MQTT service first. `data.json` contains driver instance configuration, including `device`, `tables`, devices, and tags; it is separate from SDK `config.yaml`. For platform mode, disable `dataFile.enable`, enable `driverGrpc.enable`, and configure platform gRPC/MQ addresses. See [data.json](./example/driver/data.json) and the [MQTT script guide](./example/driver/README.md).

## Driver Reporting and Management APIs

Use `driver.App` inside a collection callback and handle the returned error:

```go
func report(ctx context.Context, app driver.App) error {
	return app.WritePoints(ctx, entity.Point{
		Table: "devices", ID: "sensor-1",
		Fields: []entity.Field{{Tag: entity.Tag{ID: "temperature"}, Value: 23.5}},
	})
}
```

Import `context`, `github.com/kesiai/sdk-go/driver`, and `github.com/kesiai/sdk-go/driver/entity`. `Table` is the table identifier, `ID` the device identifier, and `Tag.ID` the tag identifier; they must match the model. `UnixTime` is a 13-digit millisecond timestamp and defaults to the current time. `WritePoints` accepts a field array and applies tag conversion. `SavePoints(ctx, table, *entity.WritePoint)` accepts a `Fields` map and saves directly without those conversions. Success does not guarantee platform persistence; local MQ does not verify platform delivery either.

Other methods include `WriteEvent`, `WriteWarning`, `WriteWarningRecovery`, `FindDevice`, `GetCommands`, `UpdateCommand`, and `RunLog`. See the [App interface](./driver/app.go) and [entity definitions](./driver/entity). Events, device queries, commands, and other platform RPCs require a working gRPC connection; warnings and recovery use MQ. Return or log errors instead of silently losing data.

With `http.enable`, `App.Start` registers and starts the HTTP server automatically:

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/health`, `/info` | Process health and instance information |
| GET | `/driver/schema`, `/driver/config` | Schema and local instance configuration file |
| POST | `/driver/start` | Start with instance configuration, including license verification |
| POST | `/driver/run`, `/driver/batch-run`, `/driver/write-tag`, `/driver/debug` | Commands, batch commands, tag writes, and debugging |
| WebSocket | `/driver/ws?type=data&table=devices&device=sensor-1` | Filter data by table/device; use `type=status` for status |

Register custom routes in `RegisterRoutes` under `/<driver.id>`. The HTTP routes do not implement application authentication; restrict management access through a gateway or access controls. A healthy `/health` response does not establish device connectivity or successful reporting. See the [HTTP implementation](./driver/http_server.go) and [Driver interface](./driver/driver.go) for HTTP and platform gRPC request shapes.

## Development Checks

```bash
go test ./... -count=1 -timeout=60s
go build ./...
```

Ordinary tests do not require etcd. To run [etcd integration tests](./etcd/etcd_test.go), set `SDK_GO_TEST_ETCD_ENDPOINTS` and optionally `SDK_GO_TEST_ETCD_USERNAME` and `SDK_GO_TEST_ETCD_PASSWORD`; without endpoints they are skipped. These tests write and clean temporary keys under `/sdk-go-test/`, so use a test instance. Also verify platform connectivity, configuration delivery, device data, commands, and shutdown against your target platform. Validate packaging on the target OS/architecture.

## Example Projects

- [example/driver](./example/driver)
- [example/driver_lazy](./example/driver_lazy)
- [example/algorithm](./example/algorithm)
- [example/data_relay](./example/data_relay)
- [example/flow](./example/flow)
- [example/flow_extension](./example/flow_extension)
- [example/service](./example/service)
- [example/task](./example/task)

## Packaging and Deployment

This section shows how to package a kesi driver built with the Go SDK. You can start from [`example/driver/main.go`](./example/driver/main.go). Replace `go-driver-mqtt-demo` with your own driver ID, and keep the program name, `service.yml`'s `Name`, and `config.yaml`'s `driver.id` consistent. For native packages, `Command` must name the actual executable.

### Choose the Target Package

| Package | Files at the archive root | Startup | Incoming port configuration |
| --- | --- | --- | --- |
| Windows ZIP | `service.yml`, `.exe`, `etc/config.yaml` | `Command` in `service.yml` names the `.exe` | Application config |
| Linux Docker `.tar.gz` | `service.yml`, Docker image archive `.tar.gz` | Image `ENTRYPOINT` | `Service`, `Path`, and `Ports` in `service.yml` |
| Linux/macOS binary `.tar.gz` | `service.yml`, executable, `etc/config.yaml` | `Command` in `service.yml` names the executable | Application config |

A **binary package** runs an executable directly. A **Docker package** imports and runs an image. Both Linux package types use `.tar.gz`, but their contents and startup methods differ. Choose the package matching the platform's runtime and CPU architecture. Go 1.25 or later is required.

**License library in platform gRPC mode:** You do not need it when the driver only receives its start configuration through platform gRPC. Keep `dataFile.enable=false` (the SDK default) and do not call the SDK HTTP start endpoint. If you also use local `dataFile` startup or the HTTP start endpoint, that startup path still requires the library. Windows, Docker, and binary packaging can each use the platform gRPC mode.

### Common Configuration

By default, the driver reads `./etc/config.yaml` relative to its working directory. Use `--config` to specify another config directory. Set at least `driver.id` and `driver.name`. For local development, `--project` and `--serviceId` can select an instance; do not hard-code development instance IDs into a release package.

For a native Windows deployment, start with this `etc/config.yaml` and adjust the platform addresses:

```yaml
driver:
  id: go-driver-mqtt-demo
  name: Go Driver Example
driverGrpc:
  host: 127.0.0.1
  port: 9224
mq:
  type: mqtt
  mqtt:
    host: 127.0.0.1
    port: 1883
```

For Linux containers, see [`example/driver/etc/config.docker.yaml`](./example/driver/etc/config.docker.yaml). With only the driver identity fields set, the SDK defaults to `driverGrpc.host=driver` and `mq.mqtt.host=mqtt`. If you enable `dataFile`, set its path to a `data.json` file accessible to the process. `dataFile.enable` defaults to `false`.

### Windows Native Package

Build the example from the repository root in PowerShell. Replace `./example/driver` with your own `main` package path:

```powershell
$env:CGO_ENABLED = '0'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
go build -tags netgo -o .\go-driver-mqtt-demo.exe ./example/driver
```

Create `service.yml`:

```yaml
Name: go-driver-mqtt-demo
Version: 1.0.0
Description: Go driver example
ConfigType: config.yaml
GroupName: driver
Command: go-driver-mqtt-demo.exe
```

Place the files at the ZIP root, without an extra project directory:

```text
go-driver-mqtt-demo-windows-x86_64.zip
├── service.yml
├── go-driver-mqtt-demo.exe
├── etc/
│   └── config.yaml
└── lib/                         # Only for local dataFile / HTTP startup
    └── license_core_windows_amd64.dll
```

Prepare `stage/` at the project root with those files, then run `Compress-Archive -Path .\stage\* -DestinationPath .\go-driver-mqtt-demo-windows-x86_64.zip` from the project root. Include any configuration page, schema, or other resources your driver uses. Complete Windows signing before compression if your release environment requires it.

### Linux Docker Package

The commands below build Linux amd64 from the project root in Bash. For arm64, use `GOARCH=arm64`, `--platform linux/arm64`, and an arm64 base image. Any required license library must also match the target architecture. Build loong64 in a compatible environment.

```bash
mkdir -p build/driver-image/etc build/driver-package
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -o build/driver-image/go-driver-mqtt-demo ./example/driver
cp example/driver/etc/config.docker.yaml build/driver-image/etc/config.yaml
```

Create `build/driver-image/Dockerfile`:

```dockerfile
FROM debian:bookworm-slim
WORKDIR /app
COPY go-driver-mqtt-demo /app/go-driver-mqtt-demo
COPY etc/config.yaml /app/etc/config.yaml
ENTRYPOINT ["/app/go-driver-mqtt-demo"]
```

This image uses platform gRPC startup and does not need a license library. If you also use local `dataFile` startup or the SDK HTTP start endpoint, place the matching `license_core_linux_amd64.so` in `build/driver-image/lib/` and add `COPY lib/license_core_linux_amd64.so /app/lib/` to the Dockerfile. Confirm that the library and its dependencies are compatible with the image's libc.

```bash
docker build --platform linux/amd64 -t gtsiot/go-driver-mqtt-demo:v1.0.0 build/driver-image
docker save -o build/driver-package/go-driver-mqtt-demo.tar gtsiot/go-driver-mqtt-demo:v1.0.0
gzip build/driver-package/go-driver-mqtt-demo.tar
```

For a driver without an incoming service, create `build/driver-package/service.yml`:

```yaml
Name: go-driver-mqtt-demo
Version: 1.0.0
Description: Go driver example
GroupName: driver
Service: None
```

Use `Service: Internal` when the platform should reach a driver HTTP endpoint without publishing its port on the host. For example, with `http.enable: true` and `http.port: 8080` in `config.yaml`:

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

Use `Service: External` to publish ports on the host. This example exposes HTTP on host port 18080 and a device TCP listener on port 8558:

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

`Container` is the port the process listens on inside the container. With `External`, `Host` is the published host port and may differ from `Container`. With `Internal`, fill in `Host` as shown, but the platform does not publish it on the host. An empty `Protocol` means TCP; use `udp` for UDP. For multiple ports, mark the HTTP route with `AppProtocol: http` so a device port is not selected as the HTTP port. `Path` is the HTTP routing prefix. `Internal` and `External` require `Ports`, and the container port must match `config.yaml`. These `Ports` belong to the Linux Docker package. The SDK's `driverGrpc.host` and `driverGrpc.port` configure an **outbound** connection to the platform; they are not incoming ports to publish.

The outer archive must have `service.yml` and the compressed image archive directly at its root:

```bash
tar -C build/driver-package -czf go-driver-mqtt-demo-linux-x86_64.tar.gz service.yml go-driver-mqtt-demo.tar.gz
tar -tzf go-driver-mqtt-demo-linux-x86_64.tar.gz
```

Upload the outer `go-driver-mqtt-demo-linux-x86_64.tar.gz`. The inner `go-driver-mqtt-demo.tar.gz` is the image produced by `docker save`; do not upload it alone.

### Linux/macOS Binary Package

Use this package only when the platform runs the driver as a native process. For Linux amd64, build the executable:

```bash
mkdir -p build/driver-binary/etc
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -o build/driver-binary/go-driver-mqtt-demo ./example/driver
```

Prepare this directory before compression:

```text
build/driver-binary/
├── service.yml
├── go-driver-mqtt-demo
├── etc/
│   └── config.yaml
└── lib/                         # Only for local dataFile / HTTP startup
    └── license_core_linux_amd64.so
```

The `service.yml` fields are similar to the Windows package, with `Command: go-driver-mqtt-demo`. Use target environment values in `config.yaml`; do not ship development `serviceId`, `project`, or platform addresses. Package the files at the archive root:

```bash
tar -C build/driver-binary -czf go-driver-mqtt-demo-linux-x86_64-binary.tar.gz service.yml go-driver-mqtt-demo etc
```

`-C` only changes directories; the three following paths select archive contents. Add `lib` at the end when the library is needed. For macOS, use `GOOS=darwin`, the appropriate `GOARCH`, a `.dylib`, and a `darwin-...-binary.tar.gz` package name.

### License Library

When the platform sends start configuration over gRPC, the SDK calls `Driver.Start` directly and does **not** load the license library. Local startup with `dataFile.enable=true` and the SDK HTTP start endpoint load `license_core` before calling `Driver.Start`. The `license` setting is a license directory path.

| Target | Library name |
| --- | --- |
| Windows amd64 / arm64 | `license_core_windows_amd64.dll` / `license_core_windows_arm64.dll` |
| Linux amd64 / arm64 / loong64 | `license_core_linux_amd64.so` / `license_core_linux_arm64.so` / `license_core_linux_loong64.so` |
| macOS amd64 / arm64 | `license_core_darwin_amd64.dylib` / `license_core_darwin_arm64.dylib` |

For startup modes that require it, place the library in `lib/` next to the executable. On Linux/macOS, the SDK also searches the working directory, `gtsiot/lib/driver/`, and corresponding locations near the executable. On Windows it also searches `license/lib/`. Without a loadable library, these startup paths report `load driver license library failed` and do not successfully call `Driver.Start`. With `dataFile.enable=true`, the process may remain running after logging the error, so a running process alone does not prove the driver started.

#### Download and Package the Library

For local `dataFile` or SDK HTTP startup, download the library matching the target operating system and CPU architecture:

| Target | Download URL |
| --- | --- |
| Windows amd64 | `https://d.gtsiot.cn/driverjs/licenselib/license_core_windows_amd64.dll` |
| Linux amd64 | `https://d.gtsiot.cn/driverjs/licenselib/license_core_linux_amd64.so` |
| Linux arm64 | `https://d.gtsiot.cn/driverjs/licenselib/license_core_linux_arm64.so` |
| macOS arm64 | `https://d.gtsiot.cn/driverjs/licenselib/license_core_darwin_arm64.dylib` |

Put the downloaded file in the driver's `lib/` directory. On Windows amd64, run this from the project root:

```powershell
New-Item -ItemType Directory -Force .\lib | Out-Null
Invoke-WebRequest 'https://d.gtsiot.cn/driverjs/licenselib/license_core_windows_amd64.dll' -OutFile '.\lib\license_core_windows_amd64.dll'
```

On Linux amd64:

```bash
mkdir -p lib
curl -fL --retry 3 -o lib/license_core_linux_amd64.so https://d.gtsiot.cn/driverjs/licenselib/license_core_linux_amd64.so
```

Verify the file exists, then include `lib/` in a native package or add the optional Dockerfile `COPY` step above. Contact your platform provider for targets not listed in the download table, such as Windows arm64, Linux loong64, and macOS amd64. Skip the download when using only the platform gRPC start flow.

### Install and Verify

1. In the operations UI, use **offline driver upload** to add the **outer** ZIP or `.tar.gz` package to the driver repository. The upload action may move between platform versions. You may also use your environment's offline repository publishing process.
2. Check `GroupName: driver`, `Name`, `Version`, and the package contents. Place `service.yml` at the archive root.
3. Select the uploaded driver in a project and install or create a driver instance. Check installation and driver logs to confirm image import or process startup, then verify connection, configuration delivery, and data reporting.
4. If installation fails, check the package type and architecture, `Command` or `Service`/`Ports`, and config paths. For local `dataFile` or HTTP startup, also check that the license library and its dependencies are available.

## FAQ

### Why is `healthRequestTime` not working?

Use nested keys:

- `driverGrpc.health.requestTime`
- `driverGrpc.health.retry`

### When is the license library required?

The platform gRPC start flow does not need the library. Local `dataFile` startup and the SDK HTTP start endpoint require a `license_core` library for the target platform. See [License Library](#license-library) for download and packaging instructions.

### What is `dataFile` used for?

Local driver runtime config loading and hot reload of `data.json`.

## Requirements

- Go `>= 1.25` (as specified by `go.mod`)

## License

See repository license file.
