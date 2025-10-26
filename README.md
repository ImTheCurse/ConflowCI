# Conflow-CI


## Build
### Creating Certificates
first, build the Certificate dependecies with:
```bash
# Build the Certificate dependencies
make build
```
then, initialize the CA(Certificate Authority) server using the following:
```bash
# Specify port and key which we will use when creating the server and client certificates using
# the CA.
make make CA PORT=<your_port_number> KEY=<your_key>
```

afterwards, we can start our CA server.

**NOTE**: the CA server should be run on the same computer the orchestrator runs on(the computer that orchestrates and sends the tasks to remote endpoints.)
```bash
make start-ca
```
and on the same device with a different terminal session, run:
```bash
make get-cert-fingerprint
```
this will give the CA fingerprint which we will use when bootstraping our remote workers and orchestrator certificates.

then, we can bootstrap our workers and orchestrator.

**NOTE**: you need to bootstrap each remote endpoint you use, for example, if you have 3 remote endpoints - use the bootstrap command for each of the remote endpoints and for the device running the orchestrator.

```bash
# We specify the CA server host and port and our previously obtained fingerprint from the previous command.
make bootstrap-CA HOST=<CA_server_host> PORT=<CA_server_port> FINGERPRINT=<cert-fingerprint>
```

next, for the orchestrator we use:
```bash
make server-cert
```

and for the workers we use:
```bash
make client-cert CLIENT_NAME=<client_name>
```
