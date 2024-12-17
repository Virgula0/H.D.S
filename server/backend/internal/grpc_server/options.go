package grpc_server

import "time"

type Option struct {
	Debug           bool
	GrpcURL         string
	GrpcConnTimeout time.Duration
}
