package networkHandler

import (
	"fmt"
	"log"
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

func EnableDnsServer(ipRange net.IPNet, port uint16) error {

	commands := []string{
		fmt.Sprintf("-w -I INPUT -p tcp -m tcp --dport %d -j ACCEPT", port),
		fmt.Sprintf("-w -I INPUT -p udp -m udp --dport %d -j ACCEPT", port),
		fmt.Sprintf("-w -t nat -I PREROUTING -s %s -d %s -p tcp -m tcp --dport 53 -j REDIRECT --to-ports %d",
			ipRange.String(), ipRange.IP.String(), port),
		fmt.Sprintf("-w -t nat -I PREROUTING -s %s -d %s -p udp -m udp --dport 53 -j REDIRECT --to-ports %d",
			ipRange.String(), ipRange.IP.String(), port),
	}

	for _, command := range commands {
		cmd := exec.Command("iptables", strings.Split(command, " ")...)
		log.Println(cmd.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil

}

func EnableInternetSharing(iface string, netSahreIface string, ipRange net.IPNet, flush bool) error {
	if flush {
		if err := IPTablesFlash(); err != nil {
			return err
		}
	}
	commands := []string{
		fmt.Sprintf("-w -t nat -I POSTROUTING -s %s ! -o %s -j MASQUERADE", ipRange.String(), iface),
		fmt.Sprintf("-w -I FORWARD -i %s -s %s -j ACCEPT", iface, ipRange.String()),
		fmt.Sprintf("-w -I FORWARD -i %s -d %s -j ACCEPT", netSahreIface, ipRange.String()),
		fmt.Sprintf("-w -I INPUT -p udp -m udp --dport 67 -j ACCEPT"),
	}

	for _, command := range commands {
		cmd := exec.Command("iptables", strings.Split(command, " ")...)
		log.Println(cmd.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func IPTablesFlash() error {

	commandArgs := []string{
		"-t nat -F",
		"-F",
	}
	for _, command := range commandArgs {
		if err := exec.Command("iptables", strings.Split(command, " ")...).Run(); err != nil {
			return err
		}
	}
	return nil
}

func IsCommandAvailable(command string) bool {
	_, err := exec.Command("which", command).Output()
	if err == nil {
		return true
	}
	return false
}
