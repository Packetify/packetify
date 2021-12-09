package networkHandler

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type NetworkService struct {
	Devices []*WifiDevice
	NetworkManager *NetworkManager
}

var MainNetworkService = NewNetworkService("packetify.conf")

func NewNetworkService(nmCfgPath string) *NetworkService {
    return &NetworkService{
		NetworkManager: NewNetworkManager(nmCfgPath),
	}
}

func (ns *NetworkService) IsNetworkInterface(iface string) bool {
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

func (ns *NetworkService) GetAdapterKernelModule(iface net.Interface) string {
	cmd := exec.Command(
		"readlink",
		"-f",
		fmt.Sprintf("/sys/class/net/%s/device/driver/module", iface.Name),
	)
	modulePath, _ := cmd.Output()
	modName := strings.Split(string(modulePath), "/")
	return modName[len(modName)-1]
}

func (ns *NetworkService) WhichCommand(command string) bool {
	_, err := exec.Command("which", command).Output()
	if err == nil {
		return true
	}
	return false
}
