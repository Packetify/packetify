package networkManager

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"../networkHandler"
)

type NetworkManager struct {
	ConfigName string
	ConfigPath string
	ConfigDir  string
}

func NewNetworkManager(configName string) *NetworkManager {
	configDir := "/etc/NetworkManager/conf.d"
	if !strings.Contains(configName, ".conf") {
		configName += ".conf"
	}
	nm := NetworkManager{ConfigName: configName, ConfigDir: configDir}
	nm.ConfigPath = fmt.Sprintf("%s/%s", configDir, configName)
	return &nm
}

//removes config file if exists
func (nm NetworkManager) RemoveConfigFile() error {
	//do nothing if file not exist
	if _, err := os.Stat(nm.ConfigPath); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(nm.ConfigPath); err != nil {
		return errors.New("error during removing config file")
	}
	return nil
}

func (nm NetworkManager) RealodProcess() {
	nmpid, err := exec.Command("pidof", "NetworkManager").Output()
	if err != nil {
		panic(err)
	}
	if len(nmpid) != 0 {
		exec.Command("kill", "-HUP", string(nmpid))
	}
}

func (nm NetworkManager) RemoveUnmanaged(iface string) error {
	if len(iface) == 0 {
		return errors.New("iface name is empty")
	}
	//checks network manager exists
	if _, err := nm.GetVersion(); err != nil {
		return errors.New("network Manager not exists")
	}

	//check if config directory exists
	if _, err := os.Stat(nm.ConfigDir); os.IsNotExist(err) {
		fmt.Println("config directory not exist")
		return nil
	}
	unmanagedIfaces := nm.ReadUnmanaged()

	hasIface := func(ifaceDevices []string) bool {
		for _, nmiface := range ifaceDevices {
			if iface == nmiface {
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
		if iface == nmiface {
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
	nm.RealodProcess()
	return nil
}

func (nm NetworkManager) AddUnmanaged(iface string) error {

	//check interface name has a valid name exclude
	iface = networkHandler.GetValidIfName(iface)

	//checks network manager exists
	if _, err := nm.GetVersion(); err != nil {
		return errors.New("network Manager not exists")
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
	nm_add_keyFile(nm.ConfigPath)

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
	nm.RealodProcess()
	return nil
}

func nm_add_keyFile(cfgFile string) {
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

// reads config file and extract unmanaged interface if it was empty it will return nil
func (nm NetworkManager) ReadUnmanaged() []string {
	r, _ := regexp.Compile("unmanaged-devices=[[:alnum:]:;,-]*")
	content, _ := ioutil.ReadFile(nm.ConfigPath)
	unmanaged := r.FindString(string(content))
	if len(unmanaged) > 18 {
		return strings.Split(strings.Replace(strings.Replace(unmanaged, "interface-name:", "", -1), "unmanaged-devices=", "", 1), ";")
	}
	return nil
}

//returns nmcli version or error if not exists
func (nm NetworkManager) GetVersion() (string, error) {
	nmversion, err := exec.Command("nmcli", "--version").Output()
	if err != nil {
		return "", errors.New("nmcli not exist")
	}
	r, _ := regexp.Compile("[0-9]+(\\.[0-9]+)*\\.[0-9]+")
	return r.FindString(string(nmversion)), nil
}

func (nm NetworkManager) IsUnmanaged(iface string) (bool, error) {
	if !nm.IsNetworkInterface(iface) {
		return false, errors.New("passed interface is not network interface")
	}

	if !nm.KnowsIface(iface) {
		return false, errors.New("NetWorkManager doesn't know specified Interface")
	}

	//gets network manager known devices and its status
	cmdOut, err := exec.Command("nmcli", "-t", "-f", "DEVICE,STATE", "d").Output()
	if err != nil {
		return false, err
	}

	r, _ := regexp.Compile(fmt.Sprintf("%s:unmanaged", iface))
	regexRes := r.FindString(string(cmdOut))
	if len(regexRes) != 0 {
		return true, nil
	}
	return false, nil
}

func (nm NetworkManager) IsNetworkInterface(iface string) bool {
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

//returns list of network ifaces known by network manager
func (nm NetworkManager) KnownIface() []string {
	output, err := exec.Command("nmcli", "-t", "-f", "DEVICE", "d").Output()
	if err != nil {
		panic(err)
	}
	return strings.Split(string(output), "\n")
}

//chrcks iiface is in network manager known ifaces list
func (nm NetworkManager) KnowsIface(iface string) bool {
	nmIfaces := nm.KnownIface()
	for _, nmIface := range nmIfaces {
		if nmIface == iface {
			return true
		}
	}
	return false
}
