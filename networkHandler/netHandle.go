package networkHandler

import "net"

func IsNetworkInterface(iface net.Interface) bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, networkIface := range ifaces {
		if networkIface.Name == iface.Name {
			return true
		}
	}
	return false
}