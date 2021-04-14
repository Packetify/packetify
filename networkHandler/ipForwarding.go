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

//enables ip forwarding via sysctl
func EnableIpForwarding(){
	//do nothing if enabled
	if IpForwardingStatus() {
		return
	}
	if _,err := exec.Command("sysctl","-w","net.ipv4.ip_forward=1").Output();err!=nil{
		panic(errors.New(" error during enabling ip forwarding"))
	}
}

//enable ip forwarding for iface and system
func EnableIpForwardingIface(iface string){
	if !IsNetworkInterface(iface){
		panic(errors.New("cant enable ip forwarding "))
	}
	EnableIpForwarding()

	//do nothing if enabled
	if IpForwardingStatusIface(iface) {
		return
	}
	devPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding",iface)
	enbit:=[]byte("1")
	err := ioutil.WriteFile(devPath,enbit,0644)
	if err!=nil{
		panic(err)
	}
}

//returns ip forwarding status of specified device iface
func IpForwardingStatusIface(iface string)bool{
	if !IsNetworkInterface(iface){
		panic(errors.New("specified device is not a network interface"))
	}
	devPath := fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding",iface)
	devStatus,_ := ioutil.ReadFile(devPath)
	result := strings.Replace(string(devStatus),"\n","",-1)
	if result == "1"{
		return true
	}
	return false
}

//checks ip forwarding status via sysctl returns true if enable and false if not
func IpForwardingStatus()bool{
	output ,_ := exec.Command("sysctl", "net.ipv4.ip_forward").Output()
	r,_:=regexp.Compile("= [01]")
	result := strings.Replace(r.FindString(string(output)),"= ","",-1)
	if result == "0"{
		return false
	}else if result == "1"{
		return true
	}
	return false
}