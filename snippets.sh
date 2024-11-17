
docker network create --driver=cilium:dev --ipam-driver=cilium:dev --attachable=true --scope=swarm ctest
docker network create --driver=cilium:dev --attachable=true ctest

docker plugin disable --force cilium:dev && docker plugin set cilium:dev plugin-args="--debug" && docker plugin enable cilium:dev

docker exec cilium cilium service update --frontend 192.168.65.2:80 --backends=192.168.65.3:80,192.168.65.4:80,192.168.65.5:80,192.168.65.6:80 --id 1

docker run --rm -e ETCDCTL_API=3 --net=host quay.io/coreos/etcd etcdctl get "" --prefix --keys-only
docker exec cilium cilium kvstore get cilium/state/nodes/v1/default/docker24-cilium-swarm-manager-1

INDEX=3; IP=172.16.0.$(($INDEX + 1)); docker service create --name etcd-$INDEX -e ALLOW_NONE_AUTHENTICATION=yes -e ETCD_ADVERTISE_CLIENT_URLS=http://$IP:2379 -e ETCD_INITIAL_ADVERTISE_PEER_URLS=http://$IP:2380 -e ETCD_LISTEN_PEER_URLS=http://0.0.0.0:2380 -e ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379 -e ETCD_DATA_DIR=/etcd-data -e ETCD_INITIAL_CLUSTER="etcd-1=http://172.16.0.2:2380,etcd-2=http://172.16.0.3:2380,etcd-3=http://172.16.0.4:2380" -e ETCD_INITIAL_CLUSTER_STATE=new -e ETCD_INITIAL_CLUSTER_TOKEN=123token -e ETCD_NAME=etcd-$INDEX --mount=type=bind,source=/etc/etcd/data,target=/etcd-data --network=host --publish published=2379,target=2379,mode=host --publish published=2380,target=2380,mode=host --constraint=node.hostname==docker24-flannel-swarm-manager-$INDEX bitnami/etcd:3.5.16
quayQ

SERVICE_NAME=s1
NETWORK_NAME=web
IPS="192.168.65.3 192.168.65.4 192.168.65.5 192.168.65.6"
NETWORK_ID=$(docker network inspect --format '{{.ID}}' $NETWORK_NAME)
VIP=$(docker service inspect --format '{{range .Endpoint.VirtualIPs}}{{if eq .NetworkID "'$NETWORK_ID'"}}{{index (split .Addr "/") 0}}{{end}}{{end}}' $SERVICE_NAME)
FWMARK=377
IFACE=lb_${NETWORK_ID:0:10}
ipvsadm -A -f $FWMARK -s rr
for IP in $IPS; do
  ipvsadm -a -f $FWMARK -r $IP:0 -m
done
iptables -t nat -A POSTROUTING -d $VIP -m mark --mark $FWMARK -j MASQUERADE
iptables -t mangle -A PREROUTING -d $VIP -p udp -j MARK --set-mark $FWMARK
iptables -t mangle -A PREROUTING -d $VIP -p tcp -j MARK --set-mark $FWMARK
modprobe dummy
ip link add $IFACE type dummy
ip addr add $VIP/32 dev $IFACE
ip link set $IFACE up
ip link set $IFACE mtu 1450
curl $VIP:80
curl $VIP:80
curl $VIP:80
curl $VIP:80


docker plugin install sovarto/docker-network-plugin-cilium:local-local --alias cilium:ll --grant-all-permissions --disable
docker plugin install sovarto/docker-network-plugin-cilium:global-local --alias cilium:gl --grant-all-permissions --disable
docker plugin install sovarto/docker-network-plugin-cilium:global-global --alias cilium:gg --grant-all-permissions --disable
docker plugin install sovarto/docker-network-plugin-cilium:local-global --alias cilium:lg --grant-all-permissions --disable

docker plugin disable --force cilium:ll
docker plugin disable --force cilium:lg
docker plugin disable --force cilium:gl
docker plugin disable --force cilium:gg

docker plugin set cilium:ll plugin-args="--debug" && docker plugin enable cilium:ll
docker plugin set cilium:lg plugin-args="--debug" && docker plugin enable cilium:lg
docker plugin set cilium:gl plugin-args="--debug" && docker plugin enable cilium:gl
docker plugin set cilium:gg plugin-args="--debug" && docker plugin enable cilium:gg

PLUGIN_NAME="cilium:gg2"; PLUGIN_ID=$(docker plugin inspect $PLUGIN_NAME --format "{{.ID}}"); journalctl -u docker.service | grep plugin=$PLUGIN_ID | grep -v "Event received" | grep -v "Error while patching" | sed -e 's/\\"/"/g' -e 's/^.\{108\}//' -E -e 's/ subsys=cilium-docker(-driver)?" plugin='$PLUGIN_ID'//'

docker service create --name sgl1 --network gl1 --mode global traefik/whoami
docker service create --name sll2 --network ll2 --mode global traefik/whoami
docker service create --name slg2 --network lg2 --mode global traefik/whoami
docker service create --name sgg1 --network gg1 --mode global traefik/whoami
