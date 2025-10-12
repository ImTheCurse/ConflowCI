package grpc

import (
	"flag"
)

var TlsFlag *bool

func DefineFlags() {
	TlsFlag = flag.Bool("tls", true, "if transport client uses TLS, by default true.")
}
