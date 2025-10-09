export PATH := $(PATH):$(shell go env GOPATH)/bin

PROTOC_GEN_FLAGS = \
  --go_out=. \
  --go-grpc_out=require_unimplemented_servers=false:. \
  --go_opt=module=github.com/ImTheCurse/ConflowCI \
  --go-grpc_opt=module=github.com/ImTheCurse/ConflowCI

github-proto:
	protoc $(PROTOC_GEN_FLAGS) proto/github/provider.proto
