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
	HostapdBasic = hstapdBase
}

type HostapdBase struct {
	Driver         string `Hostapd:"driver"`
	Ssid           string `Hostapd:"ssid"`
	Interface      string `Hostapd:"interface"`
	BeaconInterval int    `Hostapd:"beacon_int"`
	Channel        int    `Hostapd:"channel"`
	Ignorebrodcast int    `Hostapd:"ignore_broadcast_ssid"`
	APIsolate      int    `Hostapd:"ap_isolate"`
	HwMode         string `Hostapd:"hw_mode"`
	WPA            int    `Hostapd:"wpa"`
	WPA_PassPhrase string `Hostapd:"wpa_passphrase"`
	WPA_KeyMgmt    string `Hostapd:"wpa_key_mgmt"`
	WPA_Pairwise   string `Hostapd:"wpa_pairwise"`
	RSN_Pairwise   string `Hostapd:"rsn_pairwise"`
	CtrlInterface  string `Hostapd:"ctrl_interface"`
}

func New(iface string, ssid string, password string, wpa int) HostapdBase {
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
	resMap := make(map[string]interface{})
	hstType := reflect.TypeOf(hst).Elem()
	hstVal := reflect.ValueOf(hst).Elem()

	for index := 0; index < hstType.NumField(); index++ {
		tag := hstType.Field(index).Tag.Get("Hostapd")
		if tag != "" {
			resMap[tag] = hstVal.Field(index).Interface()
		}
	}
	return resMap
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
