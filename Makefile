export PATH := $(PATH):$(shell go env GOPATH)/bin

PROTOC_GEN_FLAGS = \
  --go_out=. \
  --go-grpc_out=require_unimplemented_servers=false:. \
  --go_opt=module=github.com/ImTheCurse/ConflowCI \
  --go-grpc_opt=module=github.com/ImTheCurse/ConflowCI

provider-proto:
	protoc $(PROTOC_GEN_FLAGS) proto/provider/provider.proto

sync-proto:
	protoc $(PROTOC_GEN_FLAGS) --proto_path=proto proto/sync/exec.proto proto/sync/build.proto

mq-proto:
	protoc $(PROTOC_GEN_FLAGS) proto/mq/consume.proto
