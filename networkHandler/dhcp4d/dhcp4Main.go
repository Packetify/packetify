// Example of minimal DHCP server:
package dhcp4d

import (
	"github.com/krolaw/dhcp4"
	"math/rand"
	"net"
	"time"
)

type Lease struct {
	ReqTime  time.Time
	ReqIP    net.IP
	Nic      string    // Client's CHAddr
	Expiry   time.Time // When the lease expires
	HostName string
}

type DHCPHandler struct {
	IP            net.IP        // Server IP to use
	Options       dhcp4.Options // Options to send to DHCP Clients
	Start         net.IP        // Start of IP range to distribute
	LeaseRange    int           // Number of IPs to distribute (starting from start)
	LeaseDuration time.Duration // Lease period
	Leases        map[int]Lease // Map to keep track of leases
}

func (h *DHCPHandler) ServeDHCP(p dhcp4.Packet, msgType dhcp4.MessageType, options dhcp4.Options) (d dhcp4.Packet) {
	switch msgType {

	case dhcp4.Discover:
		free, nic := -1, p.CHAddr().String()
		for i, v := range h.Leases { // Find previous lease
			if v.Nic == nic {
				free = i
				goto reply
			}
		}
		if free = h.FreeLease(); free == -1 {
			return
		}
	reply:
		return dhcp4.ReplyPacket(p, dhcp4.Offer, h.IP, dhcp4.IPAdd(h.Start, free), h.LeaseDuration,
			h.Options.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]))

	case dhcp4.Request:
		if server, ok := options[dhcp4.OptionServerIdentifier]; ok && !net.IP(server).Equal(h.IP) {
			return nil // Message not for this dhcp server
		}
		reqIP := net.IP(options[dhcp4.OptionRequestedIPAddress])

		if reqIP == nil {
			reqIP = net.IP(p.CIAddr())
		}

		if len(reqIP) == 4 && !reqIP.Equal(net.IPv4zero) {
			if leaseNum := dhcp4.IPRange(h.Start, reqIP) - 1; leaseNum >= 0 && leaseNum < h.LeaseRange {
				if l, exists := h.Leases[leaseNum]; !exists || l.Nic == p.CHAddr().String() {
					h.Leases[leaseNum] = Lease{Nic: p.CHAddr().String(), Expiry: time.Now().Add(h.LeaseDuration), ReqTime: time.Now(), ReqIP: reqIP}
					return dhcp4.ReplyPacket(p, dhcp4.ACK, h.IP, reqIP, h.LeaseDuration,
						h.Options.SelectOrderOrAll(options[dhcp4.OptionParameterRequestList]))
				}
			}
		}
		return dhcp4.ReplyPacket(p, dhcp4.NAK, h.IP, nil, 0, nil)

	case dhcp4.Release, dhcp4.Decline:
		nic := p.CHAddr().String()
		for i, v := range h.Leases {
			if v.Nic == nic {
				delete(h.Leases, i)
				break
			}
		}
	}
	return nil
}

func (h *DHCPHandler) FreeLease() int {
	now := time.Now()
	b := rand.Intn(h.LeaseRange) // Try random first
	for _, v := range [][]int{[]int{b, h.LeaseRange}, []int{0, b}} {
		for i := v[0]; i < v[1]; i++ {
			if l, ok := h.Leases[i]; !ok || l.Expiry.Before(now) {
				return i
			}
		}
	}
	return -1
}
