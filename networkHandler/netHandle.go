package networkHandler

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
)

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

func GetAdapterKernelModule(iface net.Interface) string {
	modulePath, _ := exec.Command("readlink", "-f", fmt.Sprintf("/sys/class/net/%s/device/driver/module", iface.Name)).Output()
	modName := strings.Split(string(modulePath), "/")
	return modName[len(modName)-1]
}