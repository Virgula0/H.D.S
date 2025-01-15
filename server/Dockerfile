FROM golang:1.23.4-bookworm

COPY server /app/server
COPY proto-definitions /app/proto-definitions
COPY proto.sh /app/proto.sh

RUN apt update -y && \
    apt install protobuf-compiler -y

WORKDIR /app/server

RUN useradd -ms /bin/bash server && \
    chown -R server:server /app/server

USER server

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

RUN make proto && \
    go mod verify && \
    go mod tidy


