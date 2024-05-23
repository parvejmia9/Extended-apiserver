# Extended-apiserver
Kubernetes Extended APIServer using net/http library

```console
# Terminal 1
$ go run apiserver/main.go
2019/01/02 18:09:41 listening on 127.0.0.1:8989

# Terminal 2
export APISERVER_ADDR=127.0.0.1:8989

$ curl -k https://${APISERVER_ADDR}/core/pods
Resource: pods
```

```console
# Terminal 3
$ go run database-apiserver/main.go
2019/01/02 18:09:45 listening on 127.0.0.2:8989

# Terminal 2
export EAS_ADDR=127.0.0.2:8989

$ curl -k https://${EAS_ADDR}/database/postgress
Resource: postgress requested by user [-]=system:anonymous
```

```console
# Terminal 1
$ go run apiserver/main.go --send-proxy-request
2019/01/02 18:09:41 listening on 127.0.0.1:8989
forwarding request to https://127.0.0.2:8989/database/postgress

# Terminal 3
$ go run database-apiserver/main.go --receive-proxy-request
2019/01/02 18:09:45 listening on 127.0.0.2:8989

# Terminal 2

$ curl -k https://${APISERVER_ADDR}/core/pods
Resource: pods

$ curl -k https://${APISERVER_ADDR}/database/postgress
Resource: postgress requested by user [Client-Cert-CN]=apiserver.parvejmia9

```