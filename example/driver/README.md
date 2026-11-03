# sdk-go-demo

驱动开发包例子

从本目录运行 `go run .`，读取 `./etc/config.yaml` 和 `./data.json`。当前示例配置启用了本地文件启动，需准备授权动态库及设备 MQTT 服务，详见[主文档](../../README.md#配置与运行约定)。`mq.type: local` 只替换 SDK 上报通道，设备 MQTT 连接仍需可用。

## 测试数据

topic: test1/gosdk/gosdk1

```json
[
  {
    "key": "p1",
    "value": 1.234
  },
  {
    "key": "p2",
    "value": 2
  }
]
```

## 解析脚本

```javascript
// 脚本返回值必须为对象数组
// id: 设备编号
// time: 时间戳(毫秒)
// fields: 数据点数据. 该字段为 JSON 对象, key 为数据点标识, value 为数据点的值
function handler(topic, message) {
  console.log("handler message", message)
  try {
    // Go 示例已经将 MQTT JSON 消息解码为数组，再传入此脚本。
    let arr = message
    console.log("handler arr", arr)
    let topics = topic.split("/");
    let field = {}
    arr.forEach(ele => {
      field[ele.key] = ele.value
    })
    return [
      {table: topics[1], id: topics[2], time: Date.now(), fields: field}
    ]
  } catch (e) {
    console.error("handler error", e)
    return []
  }
}
```

## 指令脚本

topic: cmd/#

```javascript
function handler(tableId, deviceId, command) {
    return {"topic": "cmd/" + tableId + "/" + deviceId, "payload": JSON.stringify(command.params)}
  }
```
