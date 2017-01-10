# etcd

### test

Run etcd in a docker container.
```
docker run --rm -p 0.0.0.0:2380:2380 -p 0.0.0.0:2379:2379 --name etcd quay.io/coreos/etcd:v3.1.0-rc.1 etcd -advertise-client-urls http://0.0.0.0:2379 -listen-client-urls http://0.0.0.0:2379 -initial-advertise-peer-urls http://0.0.0.0:2380 -listen-peer-urls http://0.0.0.0:2380
```

Run the integration tests.
```
GOOS=darwin; GOARCH=amd64 go test -tags integration $(glide novendor)
```

Cleanup the keyspace within etcd.
```
docker run --rm -e ETCDCTL_API=3 --net host --name etcdctl quay.io/coreos/etcd:v3.1.0-rc.1 etcdctl del --prefix ""
```
