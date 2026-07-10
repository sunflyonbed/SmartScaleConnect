# SmartScaleConnect 中文说明

SmartScaleConnect 用于在不同体脂秤生态之间同步体重、BMI、体脂率等数据。本仓库在原项目基础上增加了 Picooc 中国区、Home Assistant add-on、本地 SQLite 去重存储、MQTT 推送等能力。

## 快速开始

配置文件默认名为 `scaleconnect.yaml`，可以放在当前执行目录，也可以通过 `-c` 指定：

```bash
./scaleconnect -c scaleconnect.yaml
```

示例：

```yaml
sync_picooc_garmin:
  from: picooc {phone} {password}
  to: garmin {email} {password}
```

如果 Picooc 账号下有多个角色，可以在密码后追加 `rolename`：

```yaml
sync_picooc_garmin:
  from: picooc {phone} {password} {rolename}
  to: garmin {email} {password}
```

启动登录后会打印可用的 `roleid` 和 `rolename`。配置中带 `rolename` 时必须和打印出的名称匹配；不带 `rolename` 时默认使用第一个角色。

## SQLite 去重存储

程序默认开启 DB 存储。每个 sync 会根据 sync 名称和配置计算一个稳定的 `sync_id`，数据库按：

```text
sync_id + 时间戳
```

做去重。相同 sync 下同一时间的数据不会重复写入；不同 sync 之间互不影响。

常用参数：

```bash
./scaleconnect -db -db-path ./data -c scaleconnect.yaml
```

说明：

- `-db`：开启 SQLite 存储，当前默认开启。
- `-db=false`：关闭 SQLite 存储。
- `-db-path ./data`：指定 DB 目录，实际文件为 `./data/scaleconnect.db`。
- `-db-path ./data/scaleconnect.db`：直接指定 DB 文件。

每次同步会打印类似日志：

```text
scaleconnect config: 1 sync(s)
sync_picooc: sync_id=8f3a... synced=123 new=5
sync_picooc: OK
```

其中：

- `scaleconnect config` 只在服务启动读取配置时打印一次。
- `synced` 是本次抓取并进入 DB 逻辑的数据条数。
- `new` 是本次相对 DB 新增的数据条数。

## Home Assistant Add-on

本地 add-on 配置位于：

```text
picooc_scaleconnect/
```

关键配置：

```yaml
startup: system
init: false
stdin: true

options:
  db: true
  db_path: ""
  sleep: 24h

schema:
  db: bool
  db_path: "str"
  sleep: "str"

map: [ "config:rw" ]
```

add-on 会读取：

```text
/config/scaleconnect.yaml
```

默认 DB 存储位置跟随运行目录。add-on 启动脚本会优先进入 `/data`，所以默认数据库为：

```text
/data/scaleconnect.db
```

如果需要指定 DB 位置，可以在 add-on 配置中设置：

```yaml
db_path: /config/scaleconnect.db
```

## Home Assistant MQTT 输出

新增 `ha_mqtt` 输出模式：

```yaml
sync_picooc_mqtt:
  from: picooc {phone} {password}
  to: ha_mqtt {ha_addr} {username} {password} {chconfig} {chreset}
```

参数含义：

- `ha_addr`：MQTT broker 地址，例如 `ha`、`ha:1883`、`192.168.1.10:1883` 或 `tcp://192.168.1.10:1883`；不写端口时默认使用 `1883`。
- `username` / `password`：MQTT 用户名和密码。
- `chconfig`：Home Assistant MQTT Discovery sensor config topic，程序会原样往这个 topic 写配置。建议使用 `homeassistant/sensor/scaleconnect_picooc/config` 这种格式。
- `chreset`：重置/触发 topic。

行为：

- 程序会自动发布 Home Assistant MQTT Discovery 配置，不需要再手写 HA `configuration.yaml`。
- Discovery 配置 topic 使用配置中的 `chconfig`，程序不会改写它，payload 会 retained 保存。
- `chconfig` 需要是 HA MQTT Discovery 的 sensor topic，例如 `homeassistant/sensor/scaleconnect_picooc/config`；如果写成普通 topic，HA 能收到 MQTT 消息但不会创建实体。
- 一个 `chconfig` 对应一个 HA 实体；多个 sync 请配置不同的 `chconfig`，否则 retained config 会互相覆盖。
- 这个 config 里只声明一个体重 sensor；实体状态来自 `Weight`。
- 设备名和实体名会带上短 `sync_id`，例如 `sync_picooc 1a2b3c4d`，方便多个 sync 同时存在时区分。
- `BMI`、`BodyFat`、`BodyWater`、`BoneMass`、`MetabolicAge`、`VisceralFat`、`BasalMetabolism`、`SkeletalMuscleMass`、`Source` 等完整 JSON 字段会作为这个实体的 attributes 保存。
- `ha_mqtt` 模式必然使用 DB 去重。
- 每次抓取后，只把 DB 中不存在的新数据推送到自动生成的 state topic。
- state topic 会基于 `chconfig` 生成，并带上 `sync_id`，例如 `homeassistant/sensor/scaleconnect_picooc/state/{sync_id}`。
- discovery config 和 state 都会使用 retained 消息，方便 HA 重启或后开始监听时读取最后状态。
- 日志会打印实际发布的 `chconfig` 和 `state_topic`，如果 `new=0` 会跳过 state 发布。
- 新数据会按 `Date` 从小到大逐条推送。
- 每条 MQTT payload 是单条体重数据的 JSON。
- 程序启动后会监听 `chreset`。
- `chreset` 收到非空 payload 后，会立即重新执行该 sync 的拉取逻辑；如果发现新数据，再推送到 state topic。

示例发布内容：

```json
{"Date":"2026-07-09T08:00:00Z","Weight":70.5,"BMI":22.1}
```

## 命令行参数

```text
-c, --config       配置文件路径，或直接传 JSON 配置
-db                开启 SQLite 存储，默认开启
-db-path           SQLite 数据库文件或目录
-i, --interactive  保持 stdin 打开，允许从 stdin 接收配置
-r, --repeat       按间隔重复执行，例如 24h、30m
```

示例：

```bash
./scaleconnect -c scaleconnect.yaml -r 24h
./scaleconnect -c scaleconnect.yaml -db-path ./data
./scaleconnect -c scaleconnect.yaml -db=false
```

## Makefile

新增 `Makefile`，常用命令：

```bash
make tidy
make fmt
make test
make build
make build-linux-amd64
make build-linux-arm64
make addon-bin
make addon-bin-amd64
make addon-bin-arm64
make addon-image
make clean
```

说明：

- `make build` 输出到 `bin/scaleconnect`。
- `make addon-bin` 会生成 add-on 使用的静态二进制：`picooc_scaleconnect/scaleconnect`。
- `make addon-image` 会基于 `picooc_scaleconnect/Dockerfile` 构建本地 add-on 镜像。

## 注意

新增 MQTT 功能依赖：

```text
github.com/eclipse/paho.mqtt.golang
```

如果依赖文件不完整，重新编译前需要执行：

```bash
go mod tidy
```
