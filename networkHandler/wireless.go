package networkHandler

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

//validate interface name, returns modified name if invalid, panic if empty
func GetValidIfName(iface string) string {
	if len(iface) == 0 {
		panic(errors.New("iface name could not be empty"))
	}
	r, _ := regexp.Compile("[[:alnum:]:;,-]*")
	iface = strings.Join(r.FindAllString(iface, -1), "")
	if _, err := strconv.Atoi(string(iface[0])); err == nil {
		iface = "ap" + iface
	}
	return iface
}

// creates a new virtual interface for access point on top of wifi interface via iw
func CreateVirtualIface(wifiIface string, ifaceName string) error {
	ifaceName = GetValidIfName(ifaceName)
	if !IsNetworkInterface(wifiIface) {
		return errors.New(fmt.Sprintf("%s is not a network interface", wifiIface))
	}

	if IsNetworkInterface(ifaceName) {
		return errors.New("interface already exists")
	}

	cmd := exec.Command("iw", "dev", wifiIface, "interface", "add", ifaceName, "type", "__ap")
	if err := cmd.Run(); err != nil {
		return errors.New("error during create new avirtual iface")
	}
	return nil
}

//deletes virtual network interface
func DeleteVirtualIface(ifaceName string) error {
	if !IsNetworkInterface(ifaceName) {
		return errors.New("error while removing virt interface because not exists")
	}
	cmd := exec.Command("iw", "dev", ifaceName, "del")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func SetupIpToIface(iface string, gatewayIP *net.IPNet) error {
	if !IsNetworkInterface(iface) {
		return errors.New("error network iface not exists for setup IP")
	}
	brodcastIP := GetBroadCastIP(gatewayIP).String()

	cidrIP := gatewayIP.String()
	setDown := exec.Command("ip", "link", "set", "down", "dev", iface)
	flush := exec.Command("ip", "addr", "flush", iface)
	setUp := exec.Command("ip", "link", "set", "up", "dev", iface)
	addIP := exec.Command("ip", "addr", "add", cidrIP, "broadcast", brodcastIP, "dev", iface)

	commandList := []*exec.Cmd{setDown, flush, setUp, addIP}
	for _, command := range commandList {
		command.Run()
	}
	return nil
}
