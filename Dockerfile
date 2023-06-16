# build stage
FROM golang:1.20 AS builder

WORKDIR /app
COPY . .
RUN go env -w GO111MODULE=on \
    && make clean test build

# build s3sync
FROM golang:1.20 AS s3sync

WORKDIR /src/
RUN git clone --branch 2.55 https://github.com/larrabee/s3sync.git

WORKDIR /src/s3sync
ENV CGO_ENABLED 0
COPY . ./
RUN go mod tidy && \
    go build -o s3sync ./cli

# final stage
FROM debian:stable-slim

LABEL name=backup-x
LABEL url=https://github.com/jeessy2/backup-x

RUN apt-get -y update \
    && apt-get install -y wget curl gpg gnupg2 software-properties-common apt-transport-https lsb-release ca-certificates

# https://www.postgresql.org/download/linux/debian/
RUN sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list' \
    && wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
    && apt-get -y update

RUN apt-get install -y postgresql-client-14 \
    && apt-get install -y default-mysql-client

WORKDIR /app

VOLUME /app/backup-x-files
ENV TZ=Asia/Shanghai
COPY --from=builder /app/backup-x /app/backup-x
COPY --from=s3sync /src/s3sync/s3sync /usr/local/bin/s3sync
EXPOSE 9977
ENTRYPOINT ["/app/backup-x"]
