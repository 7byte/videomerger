FROM --platform=$BUILDPLATFORM golang:1.22.5 AS go_build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

COPY . .

RUN CGO_ENABLE=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o merge_xiaomi_monitor_video main.go

FROM linuxserver/ffmpeg:7.0.2

WORKDIR /app

COPY --from=go_build /build/merge_xiaomi_monitor_video /app/merge_xiaomi_monitor_video

VOLUME ["/app/videos", "/app/output"]

ENTRYPOINT ["/app/merge_xiaomi_monitor_video", "merge", "-i", "/app/videos", "-o", "/app/output"]
CMD []