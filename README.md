# videomerger

`videomerger` 是一个用于合并监控视频文件的工具，支持通过命令行参数指定输入路径、输出路径、日期范围以及定时任务等功能。

## 功能特性

- 合并指定目录下的监控视频文件
- 支持通过日期范围筛选视频文件
- 支持设置定时任务定期合并视频文件
- 自动跳过已完成合并的视频

### TODO
- 支持更多的视频文件名格式（目前只支持识别`20240706212250_20240706213600.mp4`）
- 支持自定义视频输出格式（目前固定输出mp4）
- 合并视频的兼容性测试

## 使用方法

### 构建项目

在项目根目录下运行以下命令构建项目：

```sh
make build
```

### 运行项目
构建完成后，可以通过以下命令运行项目：

```sh
./bin/videomerger merge -i <输入路径> -o <输出路径> -r <日期范围> -c <cron表达式>
```

### 命令行参数
* `--input_path, -i`：待合并视频文件的目录（必填）
* `--output_path, -o`：合并后视频的输出路径，默认为当前路径
* `--cron_spec, -c`：cron 表达式，默认为空，即只运行一次。cron语法参考：https://en.m.wikipedia.org/wiki/Cron
* `--date_range, -r`：日期范围，如 "20060102-20060202"，开始日期和结束日期都可以为空，默认为空，即合并所有文件

### 示例

合并指定目录下的所有视频文件：

```sh
./bin/videomerger merge -i /path/to/input -o /path/to/output
```

合并指定日期范围内的视频文件：

```sh
./bin/videomerger merge -i /path/to/input -o /path/to/output -r 20230101-20230631
```

设置定时任务，每天凌晨 2 点合并视频文件：

```sh
nohup ./bin/videomerger merge -i /path/to/input -o /path/to/output -c "0 2 * * *" > output.log 2>&1 &
```

### Docker运行

本地构建Docker镜像

```sh
# 创建多平台构建builder，如果只需要本地运行可跳过该步骤
make multi_builder

# 构建本地docker镜像
make image
```

单次运行

```sh
docker run --rm -v /path/to/input:/app/videos -v /path/to/output:/app/output github.com/7byte/videomerger:latest
```

后台运行，每天凌晨 2 点合并视频文件

```sh
docker run -d \
--name=videomerger \
--restart always \
-e TZ=Asia/Shanghai \
-v /path/to/input:/app/videos \
-v /path/to/output:/app/output \
github.com/7byte/videomerger:latest -c "0 2 * * *"
```