package networkHandler

import "net"

func IsNetworkInterface(iface string) bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for _, networkIface := range ifaces {
		if networkIface.Name == iface {
			return true
		}
	}
	return false
}