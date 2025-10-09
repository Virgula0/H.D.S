FROM golang:1.23.4-bookworm

# Install required dependencies
RUN apt update -y && \
    apt install -y --no-install-recommends \
    libminizip-dev \
    ocl-icd-libopencl1 \
    opencl-headers \
    pocl-opencl-icd \
    build-essential \
    ca-certificates \
    libz-dev \
    libssl-dev \
    dbus \
    sqlite3 \
    sqlite-utils \
    # Graphic libraries for raylib
    libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev libwayland-dev libxkbcommon-dev \
    protobuf-compiler # install protoc

RUN dbus-uuidgen > /etc/machine-id

WORKDIR /app 

# you need to have dont a git recursive pull first
COPY externals/gocat /app/externals/gocat
COPY externals/hcxtools /app/externals/hcxtools

# this trick needs to be improved, but needed for git commands in 
# gocat Makefile. this copies a ton of GB inside the container
COPY .git /app/.git

# create user
RUN useradd -ms /bin/bash client

WORKDIR /app/externals/gocat
# Install hashcat 6.1.1
RUN make install && \
    make set-user-permissions USER=client && \
    chown client:client /app/externals/hcxtools && \
    # IMPORTANT! RUN TEST IN ORDER TO LET HASHCAT TO CREATE KERNELS
    make test 

COPY client /app/client
COPY proto-definitions /app/proto-definitions
COPY proto.sh /app/proto.sh

WORKDIR /app/client

RUN chown -R client:client /app && \
    # give permissions to run go on client
    chown -R client:client /go

USER client

# install proto
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

ENV LD_LIBRARY_PATH=/app/client/build:/usr/local/lib

RUN make build && \
    rm -rf /app/.git
