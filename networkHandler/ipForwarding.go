package networkHandler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strings"
)

// EnableIpForwarding enables ip forwarding via sysctl
func (ns *NetworkService) EnableIpForwarding() error {
	//do nothing if enabled
	if ns.IpForwardingStatus() {
		return nil
	}
	if _, err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Output(); err != nil {
		return fmt.Errorf(" Error Enable IP forwarding %v", err)
	}
	return nil
}

// EnableIpForwardingIface enables ip forwarding for iface and system
func (ns *NetworkService) EnableIpForwardingIface(iface net.Interface) error {
	if !ns.IsNetworkInterface(iface.Name) {
		return errors.New("Error Enable IPForwarding " + iface.Name + " is not iface")
	}
	if err := ns.EnableIpForwarding(); err != nil {
		return err
	}

	//do nothing if enabled
	if ns.IpForwardingStatusIface(iface) {
		return nil
	}
	devPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", iface.Name)
	enbit := []byte("1")
	err := ioutil.WriteFile(devPath, enbit, 0644)
	if err != nil {
		return err
	}
	return nil
}

// IpForwardingStatusIface returns ip forwarding status of specified network iface
func (ns *NetworkService) IpForwardingStatusIface(iface net.Interface) bool {
	if !ns.IsNetworkInterface(iface.Name) {
		log.Fatalf("Error: IpForwardingStatusIface(%s)", iface.Name)
	}
	devPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", iface.Name)
	devStatus, _ := ioutil.ReadFile(devPath)
	result := strings.Replace(string(devStatus), "\n", "", -1)
	if strings.Contains(result, "1") {
		return true
	}
	return false
}

// IpForwardingStatus checks ip forwarding status via sysctl returns true if enable and false if not
func (ns *NetworkService) IpForwardingStatus() bool {
	cmdString := "sysctl net.ipv4.ip_forward | cut -d= -f2"
	cmd := exec.Command("bash", "-c", cmdString)
	out, err := cmd.Output()
	log.Println(cmd.String())
	if err != nil {
		log.Fatalf("error during getting ip forwarding status: %s", err)
		return false
	}

	if strings.Contains(string(out), "1") {
		return true
	}
	return false
}

// DisableIpForwarding disable ip forwarding via sysctl
func (ns *NetworkService) DisableIpForwarding() error {
	//do nothing if disabled
	if !ns.IpForwardingStatus() {
		return nil
	}
	if _, err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=0").Output(); err != nil {
		return err
	}
	return nil
}

// DisableIpForwardingIface disable ip forwarding for iface and system
func (ns *NetworkService) DisableIpForwardingIface(iface net.Interface) error {
	if !ns.IsNetworkInterface(iface.Name) {
		return fmt.Errorf("Error:  Disable IPForwarding %v", iface.Name)
	}
	if err := ns.DisableIpForwarding(); err != nil {
		return err
	}

	//do nothing if disabled
	if !ns.IpForwardingStatusIface(iface) {
		return nil
	}
	devPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", iface.Name)
	disbit := []byte("1")
	err := ioutil.WriteFile(devPath, disbit, 0644)
	if err != nil {
		panic(err)
	}
	return nil
}
