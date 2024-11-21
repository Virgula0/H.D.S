# Use Ubuntu as the base image
FROM golang:1.23.3-bookworm

# Install required dependencies
RUN apt update -y && \
    apt install -y --no-install-recommends \
    libminizip-dev \
    ocl-icd-libopencl1 \
    opencl-headers \
    pocl-opencl-icd \
    build-essential \
    wget \
    git \
    dumb-init \
    ca-certificates \
    libz-dev

# Set ENV
ENV HASHCAT_SRC_PATH=/app/hashcat
ENV CGO_CFLAGS="-I$HASHCAT_SRC_PATH/OpenCL -I$HASHCAT_SRC_PATH/deps/LZMA-SDK/C -I$HASHCAT_SRC_PATH/deps/zlib -I$HASHCAT_SRC_PATH/deps/zlib/contrib -I$HASHCAT_SRC_PATH/deps/OpenCL-Headers $CGO_CFLAGS"
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib

WORKDIR /app/hashcat

# Clone and install Hashcat 6.1.1
RUN git clone --branch v6.1.1 https://github.com/hashcat/hashcat . && \
    make install SHARED=1 ENABLE_BRAIN=0 && \
    cp deps/LZMA-SDK/C/LzmaDec.h /usr/local/include/hashcat/ && \
    cp deps/LZMA-SDK/C/7zTypes.h /usr/local/include/hashcat/ && \
    cp deps/LZMA-SDK/C/Lzma2Dec.h /usr/local/include/hashcat/ && \
    cp -r OpenCL/inc_types.h /usr/local/include/hashcat/ && \
    cp -r deps/zlib/contrib /usr/local/include/hashcat && \
    # Important needed symlinks to the hashcat library
    ln -s /usr/local/lib/libhashcat.so.6.1.1 /usr/local/lib/libhashcat.so &&\
    ln -s /usr/local/lib/libhashcat.so.6.1.1 /usr/lib/libhashcat.so.6.1.1

WORKDIR /app/gocat

# Clone the repository for the Go application
RUN git clone https://github.com/mandiant/gocat . && \
    go test -c && \
    cp gocat.test /usr/local/share/hashcat && \
    cp -r testdata /usr/local/share/hashcat

# EVENTUALLY YOU CAN SKIP TESTS
RUN /usr/local/share/hashcat/gocat.test 

COPY . /app/client
WORKDIR /app/client

RUN useradd -ms /bin/bash client && \
    chown -R client:client /app/client && \
    chown -R client:client /usr/local/share/hashcat && \
    chown -R client:client /go 

USER client

RUN go mod verify && \
    go mod tidy && \
    go build main.go

ENTRYPOINT ["tail", "-f", "/dev/null"]