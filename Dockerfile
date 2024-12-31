FROM gocv/opencv:4.10.0

ARG TARGETARCH

WORKDIR /build

RUN apt-get update && apt-get install -y ffmpeg && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* /var/cache/apt/archives/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH go build -o /app/videomerger main.go

WORKDIR /app

VOLUME ["/app/videos", "/app/output"]

ENTRYPOINT ["/app/videomerger", "-i", "/app/videos", "-o", "/app/output"]
CMD []
