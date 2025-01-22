FROM golang:1.23.4-bookworm

RUN apt update && \
    apt install -y libpcap0.8-dev

COPY . /app/rsp

WORKDIR /app/rsp

RUN dbus-uuidgen > /etc/machine-id && \
    useradd -ms /bin/bash rsp && \
    chown -R rsp:rsp /app/rsp

USER rsp

RUN make build

