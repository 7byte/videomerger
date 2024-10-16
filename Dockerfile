FROM --platform=$BUILDPLATFORM golang:1.22.5 AS go_build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

COPY . .

RUN CGO_ENABLE=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o videomerger main.go

FROM linuxserver/ffmpeg:7.0.2

WORKDIR /app

COPY --from=go_build /build/videomerger /app/videomerger

VOLUME ["/app/videos", "/app/output"]

ENTRYPOINT ["/app/videomerger", "merge", "-i", "/app/videos", "-o", "/app/output"]
CMD []