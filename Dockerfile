FROM golang:1.22.5 AS go_build

WORKDIR /build

COPY . .

RUN CGO_ENABLE=0 go build -o merge_xiaomi_monitor_video main.go

FROM ubuntu:24.04

# 更新包列表并安装依赖
RUN apt-get update && apt-get install -y \
    software-properties-common \
    build-essential \
    wget \
    yasm \
    unzip \
    # 安装FFmpeg和依赖
    && apt-get install -y ffmpeg \
    # 清理不再需要的包
    && apt-get autoremove -y && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=go_build /build/merge_xiaomi_monitor_video /app/merge_xiaomi_monitor_video

VOLUME ["/app/videos", "/app/output"]

ENTRYPOINT ["/app/merge_xiaomi_monitor_video", "merge", "-i", "/app/videos", "-o", "/app/output"]
CMD []