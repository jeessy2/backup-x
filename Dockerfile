# build stage
FROM golang:1.17 AS builder

WORKDIR /app
COPY . .
RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && make clean test build

# build s3sync
FROM golang:1.17 AS s3sync

WORKDIR /src/
RUN git clone --branch 2.33 https://github.com/larrabee/s3sync.git

WORKDIR /src/s3sync
ENV CGO_ENABLED 0
COPY . ./
RUN go mod vendor && \
    go build -o s3sync ./cli

# final stage
FROM debian:stable-slim

LABEL name=backup-x
LABEL url=https://github.com/jeessy2/backup-x

RUN apt-get -y update  \
    && apt-get install -y ca-certificates curl  \
    && apt-get install -y postgresql-client \
    && apt-get install -y default-mysql-client

WORKDIR /app

VOLUME /app/backup-x-files
ENV TZ=Asia/Shanghai
COPY --from=builder /app/backup-x /app/backup-x
COPY --from=s3sync /src/s3sync/s3sync /usr/local/bin/s3sync
EXPOSE 9977
ENTRYPOINT ["/app/backup-x"]
