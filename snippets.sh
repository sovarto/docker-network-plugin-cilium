
docker network create --driver=cilium:dev --ipam-driver=cilium:dev --attachable=true --scope=swarm ctest
docker network create --driver=cilium:dev --attachable=true ctest

docker plugin disable --force cilium:dev && docker plugin set cilium:dev plugin-args="--debug" && docker plugin enable cilium:dev


docker run --rm -e ETCDCTL_API=3 --net=host quay.io/coreos/etcd etcdctl get / --prefix --keys-only


INDEX=3; IP=172.16.0.$(($INDEX + 1)); docker service create --name etcd-$INDEX -e ALLOW_NONE_AUTHENTICATION=yes -e ETCD_ADVERTISE_CLIENT_URLS=http://$IP:2379 -e ETCD_INITIAL_ADVERTISE_PEER_URLS=http://$IP:2380 -e ETCD_LISTEN_PEER_URLS=http://0.0.0.0:2380 -e ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379 -e ETCD_DATA_DIR=/etcd-data -e ETCD_INITIAL_CLUSTER="etcd-1=http://172.16.0.2:2380,etcd-2=http://172.16.0.3:2380,etcd-3=http://172.16.0.4:2380" -e ETCD_INITIAL_CLUSTER_STATE=new -e ETCD_INITIAL_CLUSTER_TOKEN=123token -e ETCD_NAME=etcd-$INDEX --mount=type=bind,source=/etc/etcd/data,target=/etcd-data --network=host --publish published=2379,target=2379,mode=host --publish published=2380,target=2380,mode=host --constraint=node.hostname==docker24-flannel-swarm-manager-$INDEX bitnami/etcd:3.5.16
quayQ
