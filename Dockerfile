# build stage
FROM golang:1.20 AS builder

WORKDIR /app
COPY . .
RUN go env -w GO111MODULE=on \
    && make clean test build

# build s3sync
FROM golang:1.20 AS s3sync

WORKDIR /src/
RUN git clone https://github.com/jeessy2/s3sync.git

WORKDIR /src/s3sync
ENV CGO_ENABLED 0
COPY . ./
RUN go mod tidy && \
    go build -o s3sync ./cli

# minio mc
FROM minio/mc:latest AS mc

# final stage
FROM debian:stable-slim

LABEL name=backup-x
LABEL url=https://github.com/jeessy2/backup-x

RUN apt-get -y update \
    && apt-get install -y wget curl gpg gnupg2 software-properties-common apt-transport-https lsb-release ca-certificates rsync lz4 zstd

# https://www.postgresql.org/download/linux/debian/
RUN sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list' \
    && wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
    && apt-get -y update

RUN apt-get install -y postgresql-client-16 \
    && apt-get install -y default-mysql-client

# add RCLone to use directly
RUN mkdir -p /root/.config/rclone/ && \
    apt-get install -y curl wget rclone vim

WORKDIR /app

VOLUME /app/backup-x-files

# config ENV and rclone-default-config-path
ENV TZ=Asia/Shanghai
ENV XDG_CONFIG_HOME=/app/backup-x-files

COPY --from=builder /app/backup-x /app/backup-x
COPY --from=s3sync /src/s3sync/s3sync /usr/local/bin/s3sync
COPY --from=mc /usr/bin/mc /usr/bin/mc

EXPOSE 9977
ENTRYPOINT ["/app/backup-x"]
