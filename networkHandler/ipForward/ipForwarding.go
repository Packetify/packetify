package ipForward

import (
	"errors"
	"fmt"
	"github.com/Packetify/packetify/networkHandler"
	"io/ioutil"
	"net"
	"os/exec"
	"regexp"
	"strings"
)


//enables ip forwarding via sysctl
func EnableIpForwarding() {
	//do nothing if enabled
	if IpForwardingStatus() {
		return
	}
	if _, err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Output(); err != nil {
		panic(errors.New(" error during enabling ip forwarding"))
	}
}

//enable ip forwarding for iface and system
func EnableIpForwardingIface(iface net.Interface) {
	if !networkHandler.IsNetworkInterface(iface.Name) {
		panic(errors.New("cant enable ip forwarding "))
	}
	EnableIpForwarding()

	//do nothing if enabled
	if IpForwardingStatusIface(iface) {
		return
	}
	devPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", iface.Name)
	enbit := []byte("1")
	err := ioutil.WriteFile(devPath, enbit, 0644)
	if err != nil {
		panic(err)
	}
}

//returns ip forwarding status of specified device iface
func IpForwardingStatusIface(iface net.Interface) bool {
	if !networkHandler.IsNetworkInterface(iface.Name) {
		panic(errors.New("specified device is not a network interface"))
	}
	devPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", iface.Name)
	devStatus, _ := ioutil.ReadFile(devPath)
	result := strings.Replace(string(devStatus), "\n", "", -1)
	if result == "1" {
		return true
	}
	return false
}

//checks ip forwarding status via sysctl returns true if enable and false if not
func IpForwardingStatus() bool {
	output, _ := exec.Command("sysctl", "net.ipv4.ip_forward").Output()
	r, _ := regexp.Compile("= [01]")
	result := strings.Replace(r.FindString(string(output)), "= ", "", -1)
	if result == "0" {
		return false
	} else if result == "1" {
		return true
	}
	return false
}

//disable ip forwarding via sysctl
func DisableIpForwarding() {
	//do nothing if disabled
	if !IpForwardingStatus() {
		return
	}
	if _, err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=0").Output(); err != nil {
		panic(errors.New(" error during disabling ip forwarding"))
	}
}

//disable ip forwarding for iface and system
func DisableIpForwardingIface(iface net.Interface) {
	if !networkHandler.IsNetworkInterface(iface.Name) {
		panic(errors.New("cant enable ip forwarding "))
	}
	DisableIpForwarding()

	//do nothing if disabled
	if !IpForwardingStatusIface(iface) {
		return
	}
	devPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", iface.Name)
	disbit := []byte("1")
	err := ioutil.WriteFile(devPath, disbit, 0644)
	if err != nil {
		panic(err)
	}
}
