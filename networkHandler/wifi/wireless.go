package wifi

import (
	"errors"
	"fmt"
	"github.com/Packetify/ipcalc/ipv4calc"
	"github.com/Packetify/packetify/networkHandler"
	"net"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type WifiDevice struct {
	Iface string
	Phy   string
}

//validate interface name, returns modified name if invalid, panic if empty
func GetValidIfName(iface net.Interface) string {
	if len(iface.Name) == 0 {
		panic(errors.New("iface name could not be empty"))
	}
	r, _ := regexp.Compile("[[:alnum:]:;,-]*")
	iface.Name = strings.Join(r.FindAllString(iface.Name, -1), "")
	if _, err := strconv.Atoi(string(iface.Name[0])); err == nil {
		iface.Name = "ap" + iface.Name
	}
	return iface.Name
}

// creates a new virtual interface for access point on top of wifi interface via iw
func CreateVirtualIface(wifiIface net.Interface, virtIface net.Interface) error {
	virtIface.Name = GetValidIfName(virtIface)
	if !networkHandler.IsNetworkInterface(wifiIface) {
		return errors.New(fmt.Sprintf("%s is not a network interface", wifiIface))
	}

	if networkHandler.IsNetworkInterface(virtIface) {
		return errors.New("interface already exists")
	}

	cmd := exec.Command("iw", "dev", wifiIface.Name, "interface", "add", virtIface.Name, "type", "__ap")
	if err := cmd.Run(); err != nil {
		return errors.New("error during create new avirtual iface")
	}
	return nil
}

//deletes virtual network interface
func DeleteVirtualIface(virtIface net.Interface) error {
	if !networkHandler.IsNetworkInterface(virtIface) {
		return errors.New("error while removing virt interface because not exists")
	}
	cmd := exec.Command("iw", "dev", virtIface.Name, "del")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func SetupIpToIface(iface net.Interface, gatewayIP *net.IPNet) error {
	if !networkHandler.IsNetworkInterface(iface) {
		return errors.New("error network iface not exists for setup IP")
	}

	ipcalc := ipv4calc.New(gatewayIP)
	brodcastIP := ipcalc.GetBroadCastIP().String()
	cidrIP := gatewayIP.String()
	setDown := exec.Command("ip", "link", "set", "down", "dev", iface.Name)
	flush := exec.Command("ip", "addr", "flush", iface.Name)
	setUp := exec.Command("ip", "link", "set", "up", "dev", iface.Name)
	addIP := exec.Command("ip", "addr", "add", cidrIP, "broadcast", brodcastIP, "dev", iface.Name)

	commandList := []*exec.Cmd{setDown, flush, setUp, addIP}
	for _, command := range commandList {
		command.Run()
	}
	return nil
}

func GetAdapterKernelModule(iface net.Interface) string {
	modulePath, _ := exec.Command("readlink", "-f", fmt.Sprintf("/sys/class/net/%s/device/driver/module", iface.Name)).Output()
	modName := strings.Split(string(modulePath), "/")
	return modName[len(modName)-1]
}

//returns phy of wifi devices by iface
//returns empty string if iface wasn't 80211 or not exist
func GetPhyOfDevice(iface net.Interface) (string, error) {
	if !networkHandler.IsNetworkInterface(iface) {
		return "", errors.New("error unkown iface can't find phy address")
	}
	devicesList := GetWifiDevices()
	for _, dev := range devicesList {
		if dev.Iface == iface.Name {
			return dev.Phy, nil
		}
	}
	return "", nil
}

//returns a list of wifi devices struct with iface and phy fields
func GetWifiDevices() []WifiDevice {
	var deviceList []WifiDevice
	phyDevices, _ := filepath.Glob("/sys/class/ieee80211/*")
	for _, phy := range phyDevices {
		ifaceList, _ := filepath.Glob(phy + "/device/net/*")
		for _, ifacePath := range ifaceList {
			ifacePhy := strings.Split(phy, "/")
			iface := strings.Split(ifacePath, "/")
			deviceList = append(deviceList, WifiDevice{Phy: ifacePhy[len(ifacePhy)-1], Iface: iface[len(iface)-1]})
		}
	}
	return deviceList
}

//returns adapter info by iface
func GetAdapterInfo(iface net.Interface) (string, error) {
	if !networkHandler.IsNetworkInterface(iface) {
		return "", errors.New("unkown iface can't show adapter info")
	}
	ifacePhy, _ := GetPhyOfDevice(iface)
	cmdOut, _ := exec.Command("iw", "phy", ifacePhy, "info").Output()
	return string(cmdOut), nil
}

//returns true if iface has AP ability
func CanBeAP(iface net.Interface) (bool, error) {
	if !networkHandler.IsNetworkInterface(iface) {
		return false, errors.New("unkown iface can't be AP")
	}
	r, _ := regexp.Compile("\\* AP")
	adapterInfo, _ := GetAdapterInfo(iface)
	return r.MatchString(adapterInfo), nil
}
