
docker network create --driver=cilium:dev --ipam-driver=cilium:dev --attachable=true --scope=swarm ctest
docker network create --driver=cilium:dev --attachable=true ctest

docker plugin disable --force cilium:dev && docker plugin set cilium:dev plugin-args="--debug" && docker plugin enable cilium:dev
