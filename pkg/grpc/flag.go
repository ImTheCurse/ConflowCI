package grpc

import (
	"flag"
	"sync"
)

var TlsFlag *bool
var once sync.Once

func DefineFlags() {
	once.Do(func() {
		TlsFlag = flag.Bool("tls", true, "enable TLS")
	})
}
