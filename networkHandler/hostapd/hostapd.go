package hostapd

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
)

var (
	HostapdBasic *viper.Viper
)

func init() {
	hstapdBase := viper.New()
	hstapdBase.SetDefault("driver", "nl80211")
	hstapdBase.SetDefault("beacon_int", 100)
	hstapdBase.SetDefault("channel", 1)
	hstapdBase.SetDefault("ignore_broadcast_ssid", 0)
	hstapdBase.SetDefault("ap_isolate", 0)
	hstapdBase.SetDefault("hw_mode", "g")
	hstapdBase.SetDefault("wpa", 1)
	hstapdBase.SetDefault("wpa_key_mgmt", "WPA-PSK")
	hstapdBase.SetDefault("rsn_pairwise", "CCMP")
	hstapdBase.SetDefault("wpa_pairwise", "TKIP CCMP")
	hstapdBase.SetDefault("ctrl_interface", "/var/run/hostapd")
	hstapdBase.SetDefault("auth_algs", 3)
	hstapdBase.SetDefault("eap_server", 0)
	hstapdBase.SetDefault("ieee8021x", 0)
	hstapdBase.SetDefault("ieee80211n", 0)
	hstapdBase.SetDefault("okc", 0)
	hstapdBase.SetDefault("disable_pmksa_caching", 0)
	HostapdBasic = hstapdBase
}

type HostapdBase struct {
	Driver         string `Hostapd:"driver"`
	Ssid           string `Hostapd:"ssid"`
	Interface      string `Hostapd:"interface"`
	BeaconInterval uint16 `Hostapd:"beacon_int"`
	Channel        int8   `Hostapd:"channel"`
	Ignorebrodcast int8   `Hostapd:"ignore_broadcast_ssid"`
	APIsolate      int8   `Hostapd:"ap_isolate"`
	HwMode         string `Hostapd:"hw_mode"`
	WPA            int8   `Hostapd:"wpa"`
	WPA_PassPhrase string `Hostapd:"wpa_passphrase"`
	WPA_KeyMgmt    string `Hostapd:"wpa_key_mgmt"`
	WPA_Pairwise   string `Hostapd:"wpa_pairwise"`
	RSN_Pairwise   string `Hostapd:"rsn_pairwise"`
	CtrlInterface  string `Hostapd:"ctrl_interface"`
	Auth_algs      int8   `Hostapd:"auth_algs"`
	EAP_server     int8   `Hostapd:"eap_server"`
	IEEE8021x      int8   `Hostapd:"ieee8021x"`
	IEEE80211n     int8   `Hostapd:"ieee80211n"`
	OKC            int8   `Hostapd:"okc"`
}

func New(iface string, ssid string, password string, wpa int8) HostapdBase {
	var hostapdStruct HostapdBase
	config := &mapstructure.DecoderConfig{
		TagName: "Hostapd",
		Result:  &hostapdStruct,
	}

	hstDecoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		panic(err)
	}
	if err := hstDecoder.Decode(HostapdBasic.AllSettings()); err != nil {
		panic(err)
	}
	hostapdStruct.Ssid = ssid
	hostapdStruct.WPA_PassPhrase = password
	hostapdStruct.WPA = wpa
	hostapdStruct.Interface = iface
	return hostapdStruct
}

func (hst *HostapdBase) ToMap() map[string]interface{} {
	return StructToMap(*hst)
}

func StructToMap(strct interface{}) map[string]interface{} {
	var res map[string]interface{}
	res = make(map[string]interface{})
	var Mapping func(hstapdSTRCT interface{})

	Mapping = func(hstapdSTRCT interface{}) {
		hstType := reflect.TypeOf(hstapdSTRCT)
		hstVal := reflect.ValueOf(hstapdSTRCT)

		if hstType.Kind() != reflect.Struct {
			panic("can't convert non struct type to map")
		}
		for index := 0; index < hstType.NumField(); index++ {
			if hstVal.Field(index).Kind() == reflect.Struct {
				Mapping(hstVal.Field(index).Interface())
			}

			tgs := hstType.Field(index).Tag
			if tgs.Get("Hostapd") != "" {
				res[tgs.Get("Hostapd")] = hstVal.Field(index).Interface()
			}
		}
		return
	}
	Mapping(strct)
	return res
}

func (hst *HostapdBase) WriteCfg(fpath string) error {
	if err := WriteCfg(fpath, hst.ToMap()); err != nil {
		return err
	}
	return nil
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

func WriteCfg(path string, cfgData map[string]interface{}) error {
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
