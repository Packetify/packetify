package networkHandler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

// EnableIpForwarding enables ip forwarding via sysctl
func (ns *NetworkService) EnableIpForwarding() {
	//do nothing if enabled
	if ns.IpForwardingStatus() {
		return
	}
	if _, err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Output(); err != nil {
		panic(errors.New(" error during enabling ip forwarding"))
	}
}

// EnableIpForwardingIface enables ip forwarding for iface and system
func (ns *NetworkService) EnableIpForwardingIface(iface net.Interface) {
	if !ns.IsNetworkInterface(iface.Name) {
		panic(errors.New("cant enable ip forwarding "))
	}
	ns.EnableIpForwarding()

	//do nothing if enabled
	if ns.IpForwardingStatusIface(iface) {
		return
	}
	devPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", iface.Name)
	enbit := []byte("1")
	err := ioutil.WriteFile(devPath, enbit, 0644)
	if err != nil {
		panic(err)
	}
}

// IpForwardingStatusIface returns ip forwarding status of specified device iface
func (ns *NetworkService) IpForwardingStatusIface(iface net.Interface) bool {
	if !ns.IsNetworkInterface(iface.Name) {
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

// IpForwardingStatus checks ip forwarding status via sysctl returns true if enable and false if not
func (ns *NetworkService) IpForwardingStatus() bool {
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

// DisableIpForwarding disable ip forwarding via sysctl
func (ns *NetworkService) DisableIpForwarding() {
	//do nothing if disabled
	if !ns.IpForwardingStatus() {
		return
	}
	if _, err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=0").Output(); err != nil {
		panic(errors.New(" error during disabling ip forwarding"))
	}
}

// DisableIpForwardingIface disable ip forwarding for iface and system
func (ns *NetworkService) DisableIpForwardingIface(iface net.Interface) {
	if !ns.IsNetworkInterface(iface.Name) {
		panic(errors.New("cant enable ip forwarding "))
	}
	ns.DisableIpForwarding()

	//do nothing if disabled
	if !ns.IpForwardingStatusIface(iface) {
		return
	}
	devPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", iface.Name)
	disbit := []byte("1")
	err := ioutil.WriteFile(devPath, disbit, 0644)
	if err != nil {
		panic(err)
	}
}
