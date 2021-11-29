# build stage
FROM golang:1.17 AS builder

WORKDIR /app
COPY . .
RUN go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && make clean build

# final stage
FROM debian:stable-slim

LABEL name=backup-x
LABEL url=https://github.com/jeessy2/backup-x

RUN apt-get -y update  \
    && apt-get install -y ca-certificates curl  \
    && apt-get install -y postgresql-client \
    && apt-get install -y default-mysql-client

RUN useradd -s /bin/bash appuser
RUN mkdir -p /app/backup-x-files \ 
    && chown -R appuser:root /app/backup-x-files
USER appuser
WORKDIR /app

VOLUME /app/backup-x-files
ENV TZ=Asia/Shanghai
COPY --from=builder /app/backup-x /app/backup-x
EXPOSE 9977
ENTRYPOINT ["/app/backup-x"]
