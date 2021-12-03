package wifi

import (
	"errors"
	"fmt"
	"github.com/Packetify/ipcalc/ipv4calc"
	"github.com/Packetify/packetify/networkHandler"
	"log"
	"math/rand"
	"net"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type WifiDevice struct {
	Phy        string
	VirtIfaces []net.Interface
	Modes      []string
	net.Interface
}

type Frequency struct {
	Channel int
	Freq    string
}

var (
	WifiDevicesList []WifiDevice
)

func init() {
	for _, cmd := range []string{"iw", "iwlist", "ip"} {
		if !networkHandler.IsCommandAvailable(cmd) {
			log.Fatalf("command %s requierd by wifi package but is not available.", cmd)
		}
	}
	WifiDevicesList = getWifiDevices()
}

//creates a new instance of wifiDevice struct
func New(iface string) *WifiDevice {
	if !networkHandler.IsNetworkInterface(iface) || !IsWifiDevice(iface) {
		panic(errors.New("can't create wifi instance, iface is not a wifi device"))
	}
	for _, dev := range WifiDevicesList {
		if dev.Name == iface {
			return &dev
		}
	}
	return nil
}

func IsWifiDevice(iface string) bool {
	for _, dev := range WifiDevicesList {
		if dev.Name == iface {
			return true
		}
	}
	return false
}

// returns a valid interfaceName for virtual network interface
func GetValidVirtIfaceName(word string) string {

	getRandName := func() string {
		src := rand.NewSource(time.Now().UnixNano())
		rand1 := rand.New(src)
		return fmt.Sprintf("%s%d", word, rand1.Intn(100))
	}
	for {
		randName := getRandName()
		if networkHandler.IsNetworkInterface(randName) {
			continue
		}
		return randName
	}
}

// validate interface name,often used for virtual iface name validation
func ValidateIfaceName(iface string) error {
	if len(iface) == 0 {
		return errors.New("iface name could not be empty")
	}
	r, _ := regexp.Compile("[a-zA-Z][a-zA-Z0-9]*")
	if r.FindString(iface) == iface {
		return nil
	}
	return errors.New("given network Iface name is invalid")
}

// creates a new virtual interface for access point on top of wifi interface via iw
func (wifiDev *WifiDevice) CreateVirtualIface(virtIface string) error {
	if err := ValidateIfaceName(virtIface); err != nil {
		return err
	}

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
			wifiDev.VirtIfaces = append(wifiDev.VirtIfaces, iface)
		}
	}
	return nil
}

// deletes virtual network interface if exist
func (wifiDev *WifiDevice) DeleteVirtualIface(virtIface string) error {

	if wifiDev.IsVirtInterfaceAdded(virtIface) {
		cmd := exec.Command("iw", "dev", virtIface, "del")
		if err := cmd.Run(); err != nil {
			return err
		}

		virtListTemp := make([]net.Interface, 0)
		for index, vif := range wifiDev.VirtIfaces {
			if vif.Name == virtIface {
				virtListTemp = append(virtListTemp, wifiDev.VirtIfaces[:index]...)
				wifiDev.VirtIfaces = append(virtListTemp, wifiDev.VirtIfaces[index+1:]...)
			}
		}
		return nil
	}
	return errors.New("virtual iface is not exist")
}

// returns true if virtual interface created before
func (wifiDev WifiDevice) IsVirtInterfaceAdded(iface string) bool {
	for _, virtIF := range wifiDev.VirtIfaces {
		if iface == virtIF.Name {
			return true
		}
	}
	return false
}

// assigns ip to virtual interface
func (wifiDev WifiDevice) SetupIpToVirtIface(gatewayIP *net.IPNet, virtIface string) error {
	if !wifiDev.IsVirtInterfaceAdded(virtIface) {
		return errors.New("ip assign failed, virtual iface not created before")
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

// returns phy of wifi devices by iface
// returns empty string if iface wasn't 80211 or not exist
func GetPhyOfDevice(iface string) (string, error) {
	if !networkHandler.IsNetworkInterface(iface) {
		return "", errors.New("error unkown iface can't find phy address")
	}
	devicesList := WifiDevicesList
	for _, dev := range devicesList {
		if dev.Name == iface {
			return dev.Phy, nil
		}
	}
	return "", nil
}

// returns a slice of wifi devices available in your machine
func getWifiDevices() []WifiDevice {
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
					wifidev.Modes = wifidev.getModes()
					deviceList = append(deviceList, wifidev)
				}
			}
		}
	}
	return deviceList
}

// returns adapter info by iface
func (wifiDev WifiDevice) GetAdapterInfo() (string, error) {
	cmdOut, _ := exec.Command("iw", "phy", wifiDev.Phy, "info").Output()
	return string(cmdOut), nil
}

// returns true if iface has AP ability
func (wifiDev WifiDevice) HasAPAndVirtIfaceMode() bool {
	count := 0
	for _, mode := range wifiDev.Modes {
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

// GetAdapterModes returns a slice of wifi adapter supported modes
func (WifiDevice WifiDevice) GetAdapterModes() []string {
	return WifiDevice.Modes
}

// GetVirtIfaces returns a slice of virtual interfaces created before
func (WifiDevice WifiDevice) GetVirtIfaces() []net.Interface {
	return WifiDevice.VirtIfaces
}

// GetSupportedFreq returns all frequencies wifi iface supports using iwlist owned by 'wireless_tools' package in arch
func (wifiDev WifiDevice) GetSupportedFreq() []Frequency {
	freqList := make([]Frequency, 0)
	cmdOut, _ := exec.Command("iwlist", wifiDev.Name, "freq").Output()
	r, _ := regexp.Compile("(\\d*)\\s:(\\s\\d*\\.\\d*.*)")
	allFreqs := r.FindAllString(string(cmdOut), -1)
	for _, fq := range allFreqs {
		tmp := strings.Split(fq, ":")
		channel, _ := strconv.Atoi(strings.Trim(tmp[0], " "))
		frequency := strings.Trim(tmp[1], " ")
		freqList = append(freqList, Frequency{Channel: channel, Freq: frequency})
	}
	return freqList
}

// checks if specified channel supported by network card
// for example:
// 2.4GHz = channel range 1-14
// 5GHz = channel tange 36-140
func (wifiDev WifiDevice) IsSupportedChannel(channel int) bool {
	allFreqs := wifiDev.GetSupportedFreq()
	for _, frq := range allFreqs {
		if frq.Channel == channel {
			return true
		}
	}
	return false
}

// deletes wifi/virtual interface if exist and returns error if not exist
func DeleteInterface(iface string) error {
	cmd := exec.Command("iw", "dev", iface, "del")
	if err := cmd.Run(); err != nil {
		return errors.New("can't delete interface cause it doesn't exist ")
	}
	return nil
}
