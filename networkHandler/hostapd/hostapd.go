package hostapd

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"os/exec"
)

type HstapdOption string

const (
	Driver            HstapdOption = "driver"
	Ssid              HstapdOption = "ssid"
	Interface         HstapdOption = "interface"
	BeaconInterval    HstapdOption = "beacon_int"
	Channel           HstapdOption = "channel"
	Ignorebrodcast    HstapdOption = "ignore_broadcast_ssid"
	APIsolate         HstapdOption = "ap_isolate"
	HwMode            HstapdOption = "hw_mode"
	WPA               HstapdOption = "wpa"
	WPA_PassPhrase    HstapdOption = "wpa_passphrase"
	WPA_KeyMgmt       HstapdOption = "wpa_key_mgmt"
	WPA_Pairwise      HstapdOption = "wpa_pairwise"
	RSN_Pairwise      HstapdOption = "rsn_pairwise"
	CtrlInterface     HstapdOption = "ctrl_interface"
	Auth_algs         HstapdOption = "auth_algs"
	EAP_server        HstapdOption = "eap_server"
	IEEE8021x         HstapdOption = "ieee8021x"
	IEEE80211n        HstapdOption = "ieee80211n"
	OKC               HstapdOption = "okc"
	AuthServer_addr   HstapdOption = "auth_server_addr"
	AuthServer_port   HstapdOption = "auth_server_port"
	AuthServer_secret HstapdOption = "auth_server_shared_secret"
)

func New(iface string, ssid string, password string, wpa int8) map[HstapdOption]interface{} {
	hstapd := make(map[HstapdOption]interface{})
	hstapd[Interface] = iface
	hstapd[Driver] = "nl80211"
	hstapd[Channel] = 6
	hstapd[BeaconInterval] = 100
	hstapd[Ignorebrodcast] = 0
	hstapd[APIsolate] = 0
	hstapd[HwMode] = "g"
	hstapd[WPA_KeyMgmt] = "WPA-PSK"
	hstapd[WPA_Pairwise] = "TKIP CCMP"
	hstapd[RSN_Pairwise] = "CCMP"
	hstapd[Ssid] = ssid
	hstapd[WPA_PassPhrase] = password
	hstapd[WPA] = wpa
	hstapd[CtrlInterface] = "/var/run/hostapd"
	return hstapd
}

func ReadCfg(path string) (map[string]interface{}, error) {
	cfgFile := viper.New()
	cfgFile.SetConfigFile(path)
	cfgFile.SetConfigType("properties")
	if err := cfgFile.ReadInConfig(); err != nil {
		return nil, err
	}
	return cfgFile.AllSettings(), nil
}

func WriteCfg(path string, cfgData map[HstapdOption]interface{}) error {
	configContent := ""
	for k, v := range cfgData {
		configContent += fmt.Sprintf("%v=%v\n", k, v)
	}
	if err := ioutil.WriteFile(path, []byte(configContent), 0644); err != nil {
		return err
	}
	return nil
}

func Run(cfgPath string, daemon bool) (*exec.Cmd, error) {
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, errors.New("config file not exist can't run hostapd")
	}
	if daemon {
		hstapdDaemonCmd := exec.Command("hostapd", "-B", cfgPath)
		if err := hstapdDaemonCmd.Run(); err != nil {
			return nil, err
		}
		return hstapdDaemonCmd, nil
	}

	hstapdCmd := exec.Command("hostapd", cfgPath)
	if err := hstapdCmd.Start(); err != nil {
		return nil, err
	}
	go func() {
		hstapdCmd.Wait()
	}()
	return hstapdCmd, nil
}
