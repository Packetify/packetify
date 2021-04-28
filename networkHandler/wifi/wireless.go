package wifi

import (
	"errors"
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
	Phy        string
	virtIfaces []net.Interface
	modes      []string
	net.Interface
}

//creates a new instance of wifiDevice struct
func New(iface string) *WifiDevice {
	if !networkHandler.IsNetworkInterface(iface) || !IsWifiDevice(iface) {
		panic(errors.New("can't create wifi instance, iface is not a wifi device"))
	}
	for _, dev := range GetWifiDevices() {
		if dev.Name == iface {
			return &dev
		}
	}
	return nil
}

func IsWifiDevice(iface string) bool {
	for _, dev := range GetWifiDevices() {
		if dev.Name == iface {
			return true
		}
	}
	return false
}

//validate interface name, returns modified name if invalid, panic if empty
func getValidIfName(iface string) string {
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
func (wifiDev *WifiDevice) CreateVirtualIface(virtIface string) error {
	virtIface = getValidIfName(virtIface)

	if networkHandler.IsNetworkInterface(virtIface) {
		return errors.New("interface already exists")
	}
	cmd := exec.Command("iw", "dev", wifiDev.Name, "interface", "add", virtIface, "type", "__ap")
	if err := cmd.Run(); err != nil {
		return errors.New("error during create new avirtual iface")
	}
	allInterfaces, _ := net.Interfaces()
	for _, iface := range allInterfaces {
		if iface.Name == virtIface {
			wifiDev.virtIfaces = append(wifiDev.virtIfaces, iface)
		}
	}
	return nil
}

//deletes virtual network interface if exist
func (wifiDev *WifiDevice) DeleteVirtualIface(virtIface string) error {

	if wifiDev.IsVirtInterface(virtIface) {
		cmd := exec.Command("iw", "dev", virtIface, "del")
		if err := cmd.Run(); err != nil {
			return err
		}

		virtListTemp := make([]net.Interface, 0)
		for index, vif := range wifiDev.virtIfaces {
			if vif.Name == virtIface {
				virtListTemp = append(virtListTemp, wifiDev.virtIfaces[:index]...)
				wifiDev.virtIfaces = append(virtListTemp, wifiDev.virtIfaces[index+1:]...)
			}
		}
		return nil
	}
	return errors.New("virtual iface is not exist")
}

//returns true if virtual interface created before
func (wifiDev WifiDevice) IsVirtInterface(iface string) bool {
	for _, virtIF := range wifiDev.virtIfaces {
		if iface == virtIF.Name {
			return true
		}
	}
	return false
}

//assigns ip to virtual interface
func (wifiDev WifiDevice) SetupIpToVirtIface(gatewayIP *net.IPNet, virtIface string) error {
	if !wifiDev.IsVirtInterface(virtIface) {
		return errors.New("ip assign failed virtual iface not exists")
	}
	ipcalc := ipv4calc.New(gatewayIP)
	brodcastIP := ipcalc.GetBroadCastIP().String()
	cidrIP := gatewayIP.String()
	setDown := exec.Command("ip", "link", "set", "down", "dev", virtIface)
	flush := exec.Command("ip", "addr", "flush", virtIface)
	setUp := exec.Command("ip", "link", "set", "up", "dev", virtIface)
	addIP := exec.Command("ip", "addr", "add", cidrIP, "broadcast", brodcastIP, "dev", virtIface)

	commandList := []*exec.Cmd{setDown, flush, setUp, addIP}
	for _, command := range commandList {
		command.Run()
	}
	return nil
}

//returns phy of wifi devices by iface
//returns empty string if iface wasn't 80211 or not exist
func GetPhyOfDevice(iface string) (string, error) {
	if !networkHandler.IsNetworkInterface(iface) {
		return "", errors.New("error unkown iface can't find phy address")
	}
	devicesList := GetWifiDevices()
	for _, dev := range devicesList {
		if dev.Name == iface {
			return dev.Phy, nil
		}
	}
	return "", nil
}

//returns a slice of wifi devices available in your machine
func GetWifiDevices() []WifiDevice {
	var deviceList []WifiDevice
	allInterfaces, _ := net.Interfaces()
	phyDevices, _ := filepath.Glob("/sys/class/ieee80211/*")
	for _, phy := range phyDevices {
		ifaceList, _ := filepath.Glob(phy + "/device/net/*")
		ifacePhy := strings.Split(phy, "/")
		phyName := ifacePhy[len(ifacePhy)-1]
		for _, ifacePath := range ifaceList {
			iface := strings.Split(ifacePath, "/")
			ifaceName := iface[len(iface)-1]
			for _, dev := range allInterfaces {
				if dev.Name == ifaceName {
					wifidev := WifiDevice{Phy: phyName, Interface: dev}
					wifidev.modes = wifidev.getModes()
					deviceList = append(deviceList, wifidev)
				}
			}
		}
	}
	return deviceList
}

//returns adapter info by iface
func (wifiDev WifiDevice) GetAdapterInfo() (string, error) {
	cmdOut, _ := exec.Command("iw", "phy", wifiDev.Phy, "info").Output()
	return string(cmdOut), nil
}

//returns true if iface has AP ability
func (wifiDev WifiDevice) HasAPAndVirtIfaceMode() bool {
	count := 0
	for _, mode := range wifiDev.modes {
		if mode == "AP" || mode == "AP/VLAN" {
			count++
		}
	}
	if count == 2 {
		return true
	}
	return false
}

//returns wifi adaptor supported modes
func (wifiDev WifiDevice) getModes() []string {
	var modeList []string
	cardInfo, _ := wifiDev.GetAdapterInfo()
	r, _ := regexp.Compile("Supported interface modes:\\n(\\t\\t\\s\\*\\s([A-Za-z-/0-9]*)\\n)*")
	adaptorModes := strings.Split(r.FindString(cardInfo), "\n")
	for _, mode := range adaptorModes {
		if strings.Contains(mode, "*") {
			mode = strings.ReplaceAll(mode, "\t\t * ", "")
			modeList = append(modeList, mode)
		}
	}
	return modeList
}

//returns a slice of wifi adapter supported modes
func (WifiDevice WifiDevice) GetAdapterModes() []string {
	return WifiDevice.modes
}

//returns a slice of virtual interfaces created before
func (WifiDevice WifiDevice) GetVirtIfaces() []net.Interface {
	return WifiDevice.virtIfaces
}
