package networkHandler

import (
	"fmt"
	"net"
	"testing"
)

var nm = NewNetworkManager("packetify.conf")

func TestGetWifiDevicesInfo(t *testing.T) {
	fmt.Println(GetWifiDevicesInfo())
}

func TestNetworkManager_KnownIface(t *testing.T) {

	fmt.Println(nm.KnownIface())
}

func TestNetworkManager_IsUnmanaged(t *testing.T) {
	fmt.Println(nm.IsUnmanaged(net.Interface{Name: "loasd"}))
}

func TestIsWifiEnabled(t *testing.T) {
	fmt.Println("wifi Enabled: ", IsWifiEnabled())
}

func TestTurnWifiOn(t *testing.T) {
	if !IsWifiEnabled() {
		fmt.Println("wifi is off let's turn it on")
	}
	if err := TurnWifiOn(); err != nil {
		panic(err)
	}
}
