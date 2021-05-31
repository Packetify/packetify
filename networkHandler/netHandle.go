package networkHandler

import (
	"fmt"
	"github.com/coreos/go-iptables/iptables"
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

func EnableInternetSharing(iface string, netSahreIface string, ipRange net.IPNet, flush bool) error {
	ipt, _ := iptables.New()
	if flush {
		if err := ipt.ClearAll(); err != nil {
			return err
		}
	}
	if err := ipt.Insert("nat", "POSTROUTING", 1, "-w", "-s", ipRange.String(), "!", "-o", iface, "-j", "MASQUERADE"); err != nil {
		return err
	}

	if err := ipt.Insert("filter", "FORWARD", 1, "-i", iface, "-s", ipRange.String(), "-j", "ACCEPT"); err != nil {
		return err
	}
	if err := ipt.Insert("filter", "FORWARD", 1, "-i", netSahreIface, "-d", ipRange.String(), "-j", "ACCEPT"); err != nil {
		return err
	}
	return nil
}

func IPTablesFlash() error {
	ipt, _ := iptables.New()
	if err := ipt.ClearAll(); err != nil {
		return err
	}
	return nil
}
