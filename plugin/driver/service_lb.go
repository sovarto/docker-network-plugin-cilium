package driver

// Source of truth:
//	- Container IPs: Cilium / etcd
//  - Service IPs: Docker
//  - Connection between containers and services: Docker / Container labels
// On Docker manager nodes:
//  - Upon event about new service: Write service ID and VIP to etcd
// On all nodes:
//  - Subscribe to etcd and docker events
//  - Upon call to CreateNetwork (TODO: Verify this is also called when restarting docker
//    and the network already exists) create new interface of type dummy - or bridge? - in
//    namespace of network (???), if it doesn't exist already. Record interface ID in etcd or derive
//    it in a way from the network ID that it is clear from the ID / name itself for which network
//    the interface is.
//    For all existing services on the network:
//    - Create new ipvs entry for a new fwmark
//    - Add known container IPs to it
//    - Create IP table rules for the IP of the interface of the service and the fwmark
//  - Upon new container IP in etcd: Get service for the container, add IP as backend to ipvs

type ServiceLoadBalancer struct {
}

func (lb *ServiceLoadBalancer) updateServiceVip()           {}
func (lb *ServiceLoadBalancer) removeService()              {}
func (lb *ServiceLoadBalancer) addContainerToService()      {}
func (lb *ServiceLoadBalancer) removeContainerFromService() {}
func (lb *ServiceLoadBalancer) createServiceLoadBalancer()  {}
func (lb *ServiceLoadBalancer) removeServiceLoadBalancer()  {}
