FROM golang:1.23.4-bookworm

COPY . /app/server

WORKDIR /app/server

RUN useradd -ms /bin/bash client && \
    chown -R client:client /app/server

USER client

RUN go mod verify && \
    go mod tidy

