package networkHandler

import (
	"encoding/binary"
	"net"
)

func GetBroadCastIP(gatewayIP *net.IPNet) net.IP {
	resIP := make(net.IP, len(gatewayIP.IP.To4()))
	wildcard := ^binary.BigEndian.Uint32(net.IP(gatewayIP.Mask).To4())
	binIP := binary.BigEndian.Uint32(gatewayIP.IP.To4())

	binary.BigEndian.PutUint32(resIP, binIP|wildcard)
	return resIP
}

func GetNetworkAddress(gatewayIP *net.IPNet) net.IP {
	resIP := make(net.IP, len(gatewayIP.IP.To4()))
	subnetIP := binary.BigEndian.Uint32(net.IP(gatewayIP.Mask).To4())
	binIP := binary.BigEndian.Uint32(gatewayIP.IP.To4())

	binary.BigEndian.PutUint32(resIP, binIP&subnetIP)
	return resIP
}
