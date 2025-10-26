SHELL := /bin/bash
HOME_PATH := $(HOME)
export PATH := $(PATH):$(shell go env GOPATH)/bin

PROTOC_GEN_FLAGS = \
  --go_out=. \
  --go-grpc_out=require_unimplemented_servers=false:. \
  --go_opt=module=github.com/ImTheCurse/ConflowCI \
  --go-grpc_opt=module=github.com/ImTheCurse/ConflowCI

build:
	apt install -y --no-install-recommends curl vim gpg ca-certificates && \
    curl -fsSL https://packages.smallstep.com/keys/apt/repo-signing-key.gpg -o /etc/apt/trusted.gpg.d/smallstep.asc && \
    echo 'deb [signed-by=/etc/apt/trusted.gpg.d/smallstep.asc] https://packages.smallstep.com/stable/debian debs main' \
    | tee /etc/apt/sources.list.d/smallstep.list && apt-get -y install step-cli step-ca && \

# TODO: add the functionality to add a custom address and not a local one, this is currently for testing purpose.
CA:
	sudo mkdir -p /etc/step-ca && \
	cd /etc/step-ca
	step ca init \
  --name "Conflow Internal CA" \
  --dns "127.0.0.1" \
  --address ":$(PORT)" \
  --provisioner "admin@conflow.internal" \
  --provisioner-password-file <(echo "$(KEY)")

start-ca:
	sudo step-ca $(HOME_PATH)/.step/config/ca.json

get-cert-fingerprint:
	sudo step certificate fingerprint $(HOME_PATH)/.step/certs/root_ca.crt

bootstrap-CA:
	step ca bootstrap --ca-url "https://$(HOST):$(PORT)" --fingerprint $(FINGERPRINT)

# TODO: add functionality to add custom SAN.
server-cert:
	sudo step ca certificate "server.internal" \
    /etc/ssl/certs/conflow_server_cert.pem \
    /etc/ssl/certs/conflow_server_key.pem \
    --san "server.internal" \
    --san "localhost" \
    --san "127.0.0.1" && \
   sudo chmod 644  /etc/ssl/certs/conflow_server_cert.pem /etc/ssl/certs/conflow_server_key.pem

client-cert:
	sudo step ca certificate $(CLIENT_NAME)\
    /etc/ssl/certs/conflow_client_cert.pem \
    /etc/ssl/certs/conflow_client_key.pem && \
    sudo chmod 644 /etc/ssl/certs/conflow_client_cert.pem /etc/ssl/certs/conflow_client_key.pem


provider-proto:
	protoc $(PROTOC_GEN_FLAGS) proto/provider/provider.proto

sync-proto:
	protoc $(PROTOC_GEN_FLAGS) --proto_path=proto proto/sync/exec.proto proto/sync/build.proto

mq-proto:
	protoc $(PROTOC_GEN_FLAGS) proto/mq/consume.proto
