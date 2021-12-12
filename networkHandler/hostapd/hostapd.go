package hostapd

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"os/exec"
)

type HostapdOptionKeys string
type HostapdOption struct {
	Key   HostapdOptionKeys
	Value interface{}
}
type HostapdConfig map[HostapdOptionKeys]interface{}

const (
	Driver            HostapdOptionKeys = "driver"
	Ssid              HostapdOptionKeys = "ssid"
	Interface         HostapdOptionKeys = "interface"
	BeaconInterval    HostapdOptionKeys = "beacon_int"
	Channel           HostapdOptionKeys = "channel"
	Ignorebrodcast    HostapdOptionKeys = "ignore_broadcast_ssid"
	APIsolate         HostapdOptionKeys = "ap_isolate"
	HwMode            HostapdOptionKeys = "hw_mode"
	WPA               HostapdOptionKeys = "wpa"
	WPA_PassPhrase    HostapdOptionKeys = "wpa_passphrase"
	WPA_PSK           HostapdOptionKeys = "wpa_psk"
	WPA_KeyMgmt       HostapdOptionKeys = "wpa_key_mgmt"
	WPA_Pairwise      HostapdOptionKeys = "wpa_pairwise"
	RSN_Pairwise      HostapdOptionKeys = "rsn_pairwise"
	CtrlInterface     HostapdOptionKeys = "ctrl_interface"
	Auth_algs         HostapdOptionKeys = "auth_algs"
	EAP_server        HostapdOptionKeys = "eap_server"
	IEEE8021x         HostapdOptionKeys = "ieee8021x"
	IEEE80211n        HostapdOptionKeys = "ieee80211n"
	OKC               HostapdOptionKeys = "okc"
	AuthServer_addr   HostapdOptionKeys = "auth_server_addr"
	AuthServer_port   HostapdOptionKeys = "auth_server_port"
	AuthServer_secret HostapdOptionKeys = "auth_server_shared_secret"
	CountryCode       HostapdOptionKeys = "country_code"
)

func New(options ...HostapdOption) HostapdConfig {
	hstapd := make(HostapdConfig)
	defaultOptions := []HostapdOption{
		{Driver, "nl80211"},
		{BeaconInterval, 100},
		{Channel, 1},
		{Ignorebrodcast, 0},
		{APIsolate, 0},
		{HwMode, "g"},
		{WPA, 3},
		{WPA_KeyMgmt, "WPA-PSK"},
		{WPA_Pairwise, "TKIP CCMP"},
		{RSN_Pairwise, "CCMP"},
		{CtrlInterface, "/var/run/hostapd"},
	}
	for _, op := range defaultOptions {
		hstapd[op.Key] = op.Value
	}
	for _, op := range options {
		hstapd[op.Key] = op.Value
	}
	return hstapd
}

func ReadCfg(path string) (HostapdConfig, error) {
	cfgFile := viper.New()
	cfgFile.SetConfigFile(path)
	cfgFile.SetConfigType("properties")
	if err := cfgFile.ReadInConfig(); err != nil {
		return nil, err
	}
	hstcfg := make(HostapdConfig)
	for key, val := range cfgFile.AllSettings() {
		hstcfg[HostapdOptionKeys(key)] = val
	}
	return hstcfg, nil
}

func WriteCfg(path string, cfgData HostapdConfig) error {
	configContent := ""
	for k, v := range cfgData {
		configContent += fmt.Sprintf("%v=%v\n", k, v)
	}
	if err := ioutil.WriteFile(path, []byte(configContent), 0644); err != nil {
		return err
	}
	return nil
}

func Run(cfgPath string, daemon bool) (cmd *exec.Cmd, err error) {
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, errors.New("config file not exist can't run hostapd")
	}

	switch daemon {
	case true:
		cmd = exec.Command("hostapd", "-B", cfgPath)
		if err := cmd.Run(); err != nil {
			return nil, err
		}
		return cmd, nil
	case false:
		cmd = exec.Command("hostapd", cfgPath)
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		go func() {
			if err := cmd.Wait(); err != nil {
				fmt.Println(err)
			}
		}()
		return cmd, nil
	}
	return nil, errors.New("hostapd run error")
}

func RemoveConfigFile(path string) error {
	e := os.Remove(path)
	if e != nil {
		return e
	}
	return nil
}
