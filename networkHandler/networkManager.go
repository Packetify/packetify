package networkHandler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type NMDevice struct {
	DevName    string
	Type       string
	Status     string
	Connection string
}

type NetworkManager struct {
	ConfigName string
	ConfigPath string
	ConfigDir  string
}

func init() {
	//checks network manager exists
	if _, err := GetVersion(); err != nil {
		panic(err)
	}
}

func NewNetworkManager(configName string) *NetworkManager {
	configDir := "/etc/NetworkManager/conf.d"
	if !strings.Contains(configName, ".conf") {
		configName += ".conf"
	}
	nm := NetworkManager{
		ConfigName: configName,
		ConfigDir:  configDir,
	}
	nm.ConfigPath = fmt.Sprintf("%s/%s", configDir, configName)
	return &nm
}

// RemoveConfigFile removes config file if exists
func (nm *NetworkManager) RemoveConfigFile() error {
	//do nothing if file not exist
	if _, err := os.Stat(nm.ConfigPath); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(nm.ConfigPath); err != nil {
		return errors.New("error during removing config file")
	}
	return nil
}

// RemoveUnmanaged removes unmanaged option for given interface from config file
func (nm *NetworkManager) RemoveUnmanaged(iface net.Interface) error {
	if len(iface.Name) == 0 {
		return errors.New("iface name is empty")
	}

	//check if config directory exists
	if _, err := os.Stat(nm.ConfigDir); os.IsNotExist(err) {
		fmt.Println("config directory not exist")
		return nil
	}
	unmanagedIfaces := nm.ReadUnmanaged()

	hasIface := func(ifaceDevices []string) bool {
		for _, nmiface := range ifaceDevices {
			if iface.Name == nmiface {
				return true
			}
		}
		return false
	}
	//do nothing if iface not exist
	if unmanagedIfaces == nil || hasIface(unmanagedIfaces) == false {
		return nil
	}

	configString := "unmanaged-devices="
	for _, nmiface := range unmanagedIfaces {
		if iface.Name == nmiface {
			continue
		}
		temp := fmt.Sprintf("interface-name:%s;", nmiface)
		configString += temp
	}
	//remove last semicolon in unmanaged devices
	configString = strings.TrimSuffix(configString, ";")

	content, _ := ioutil.ReadFile(nm.ConfigPath)
	lines := strings.Split(string(content), "\n")
	for index, line := range lines {
		if strings.Contains(line, "unmanaged-devices=") {
			lines[index] = configString
		}
	}
	output := strings.Join(lines, "\n")
	if err := ioutil.WriteFile(nm.ConfigPath, []byte(output), 0755); err != nil {
		return errors.New("can't write interface to config file")
	}
	return nil
}

// AddUnmanaged adds interface as unmanaged interface by network manager
// and changes will be written into file
func (nm NetworkManager) AddUnmanaged(iface string) error {

	if !MainNetworkService.IsNetworkInterface(iface) {
		return errors.New(fmt.Sprintf("the %s is not a network interface make sure it's availabe or created", iface))
	}

	if err := UnmanageIface(iface); err != nil {
		return err
	}

	//create config directory if not exist
	configDir := "/etc/NetworkManager/conf.d"
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if mkdirErr := os.Mkdir(configDir, 0755); mkdirErr != nil {
			return errors.New("can't create network manager config directory ")
		}
	}

	//add keyFile if not exist
	//creates config file if not exist
	nmAddKeyFile(nm.ConfigPath)

	unmanagedIfaces := nm.ReadUnmanaged()
	configString := "unmanaged-devices="
	if unmanagedIfaces != nil {
		for _, unmIface := range unmanagedIfaces {
			//do nothing if it's unmanaged
			if iface == unmIface {
				return nil
			}
		}

		//create unmanaged text for config file
		//append iface to unmanaged devices
		unmanagedIfaces = append(unmanagedIfaces, iface)
		for _, unmIface := range unmanagedIfaces {
			temp := fmt.Sprintf("interface-name:%s;", unmIface)
			configString += temp
		}

		//remove last semicolon in unmanaged devices
		configString = strings.TrimSuffix(configString, ";")

		content, _ := ioutil.ReadFile(nm.ConfigPath)
		lines := strings.Split(string(content), "\n")
		for index, line := range lines {
			if strings.Contains(line, "unmanaged-devices=") {
				lines[index] = configString
			}
		}
		output := strings.Join(lines, "\n")
		if err := ioutil.WriteFile(nm.ConfigPath, []byte(output), 0755); err != nil {
			return errors.New("can't write interface to config file")
		}
	} else {
		configString := configString + "interface-name:" + iface + "\n"
		f, err := os.OpenFile(nm.ConfigPath, os.O_APPEND|os.O_WRONLY, 0755)
		if err != nil {
			return errors.New("can't write interface to config file")
		}
		defer f.Close()
		f.WriteString(configString)
	}
	return nil
}

// UnmanageIface set iface as unmanaged via nmcli
func UnmanageIface(iface string) error {
	cmd := exec.Command("nmcli", "device", "set", iface, "managed", "no")
	if err := cmd.Run(); err != nil {
		return errors.New("nmcli error for add unmanage device")
	}
	return nil
}

func nmAddKeyFile(cfgFile string) {
	keyFileString := []byte("[keyfile]\n")
	r, _ := regexp.Compile("\\[keyfile\\]")
	content, err := ioutil.ReadFile(cfgFile)

	//create cfg file if not exist and write keyfile
	if os.IsNotExist(err) {
		ioutil.WriteFile(cfgFile, keyFileString, 0755)
		return
	}

	//if file and keyfile exist do nothing
	keyFile := r.FindString(string(content))
	if len(keyFile) > 0 {
		return
	}
	//if file exist but is empty write keyfile
	ioutil.WriteFile(cfgFile, keyFileString, 0755)
}

// ReadUnmanaged reads config file and extract unmanaged interface if it was empty it will return nil
func (nm NetworkManager) ReadUnmanaged() []string {
	r, _ := regexp.Compile("unmanaged-devices=[[:alnum:]:;,-]*")
	content, _ := ioutil.ReadFile(nm.ConfigPath)
	unmanaged := r.FindString(string(content))
	if len(unmanaged) > 18 {
		return strings.Split(strings.Replace(strings.Replace(unmanaged, "interface-name:", "", -1), "unmanaged-devices=", "", 1), ";")
	}
	return nil
}

// GetVersion returns nmcli version or error if not exists
func (nm NetworkManager) GetVersion() (string, error) {
	if version, err := GetVersion(); err != nil {
		return "", err
	} else {
		return version, nil
	}

}

// GetVersion returns the version of networkmanager using nmcli
func GetVersion() (string, error) {
	nmversion, err := exec.Command("nmcli", "--version").Output()
	if err != nil {
		return "", errors.New("nmcli(Network Manager) not available")
	}
	r, _ := regexp.Compile("[0-9]+(\\.[0-9]+)*\\.[0-9]+")
	return r.FindString(string(nmversion)), nil
}

// IsUnmanaged returns true if interface is unmanaged by networkmanager
func (nm NetworkManager) IsUnmanaged(iface net.Interface) (bool, error) {
	if !MainNetworkService.IsNetworkInterface(iface.Name) {
		return false, errors.New("passed interface is not network interface")
	}

	if !nm.KnowsIface(iface) {
		return false, errors.New("NetWorkManager doesn't know specified Interface")
	}

	//gets network manager known devices and its status
	for _, dev := range GetWifiDevicesInfo() {
		if dev.DevName == iface.Name && dev.Status == "unmanaged" {
			return true, nil
		}
	}
	return false, nil
}

// KnownIface returns list of network ifaces known by network manager
func (nm NetworkManager) KnownIface() []string {
	var ifaceList []string
	for _, dev := range GetWifiDevicesInfo() {
		ifaceList = append(ifaceList, dev.DevName)
	}
	return ifaceList
}

// KnowsIface checks iface is in network manager known ifaces list
func (nm NetworkManager) KnowsIface(iface net.Interface) bool {
	nmIfaces := nm.KnownIface()
	for _, nmIface := range nmIfaces {
		if nmIface == iface.Name {
			return true
		}
	}
	return false
}

// GetWifiDevicesInfo returns all netwotk interface informations via nmcli
func GetWifiDevicesInfo() (deviceList []NMDevice) {

	output, err := exec.Command("nmcli", "-t", "-f", "DEVICE,TYPE,STATE,CONNECTION", "d").Output()
	if err != nil {
		panic(err)
	}
	devicesInfo := strings.Split(string(output), "\n")
	for _, devline := range devicesInfo[:len(devicesInfo)-1] {
		devTmp := strings.Split(devline, ":")
		deviceList = append(deviceList, NMDevice{devTmp[0], devTmp[1], devTmp[2], devTmp[3]})
	}
	return deviceList
}

// IsWifiEnabled checks if wifi is enable via nmcli
func IsWifiEnabled() bool {
	wifiStat, _ := exec.Command("nmcli", "radio", "wifi").Output()
	if string(wifiStat[:len(wifiStat)-1]) == "enabled" {
		return true
	}
	return false
}

// TurnWifiOn turns wifi on via nmcli
func TurnWifiOn() error {
	if _, err := exec.Command("nmcli", "radio", "wifi", "on").Output(); err != nil {
		return err
	}
	return nil
}

// TurnWifiOff turns wifi off via nmcli
func TurnWifiOff() error {
	if _, err := exec.Command("nmcli", "radio", "wifi", "off").Output(); err != nil {
		return err
	}
	return nil
}
