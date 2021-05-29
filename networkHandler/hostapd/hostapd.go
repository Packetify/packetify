package hostapd

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type HostapdBase struct {
	Driver         string `Hostapd:"driver" Default:"nl80211"`
	Ssid           string `Hostapd:"ssid" Default:"Packetify"`
	Interface      string `Hostapd:"interface"`
	BeaconInterval int    `Hostapd:"beacon_int" Default:"100" min:"15" max:"65535"`
	Channel        int    `Hostapd:"channel" Default:"1"`
	Ignorebrodcast int    `Hostapd:"ignore_broadcast_ssid" Default:"0" min:"1" max:"2"`
	APIsolate      int    `Hostapd:"ap_isolate" Default:"0"`
	HwMode         string `Hostapd:"hw_mode" Default:"g"`
	WPA            int    `Hostapd:"wpa" Default:"1"`
	WPA_PassPhrase string `Hostapd:"wpa_passphrase"`
	WPA_KeyMgmt    string `Hostapd:"wpa_key_mgmt" Default:"WPA-PSK"`
	WPA_Pairwise   string `Hostapd:"rsn_pairwise" Default:"CCMP"`
	CtrlInterface  string `Hostapd:"ctrl_interface" Default:"/var/run/hostapd"`
}

//writes structs fields to given config file path by looking fields tags
func WriteToCfgFile(hostapdCFG interface{}, fpath string) {
	var reSTR string
	var configure func(hstapd interface{})

	configure = func(hstapd interface{}) {
		cfgType := reflect.TypeOf(hstapd)
		cfgVal := reflect.ValueOf(hstapd)
		if cfgType.Kind() != reflect.Struct {
			panic("to write Hostapd config file, struct type needed")
		}

		for index := 0; index < cfgType.NumField(); index++ {
			if cfgVal.Field(index).Kind() == reflect.Struct {
				configure(cfgVal.Field(index).Interface())
			}

			tgs := cfgType.Field(index).Tag
			if tgs.Get("Hostapd") != "" {
				reSTR += fmt.Sprintf("%s=%v\n", tgs.Get("Hostapd"), cfgVal.Field(index).Interface())
			}
		}
		return
	}
	configure((hostapdCFG))
	ioutil.WriteFile(fpath, []byte(reSTR), 0644)
}

//checks struct fields type if they wasn't declared it will assign default value by looking Default tag
func (hst *HostapdBase) FillByDefault() {
	cfgType := reflect.TypeOf(hst).Elem()
	cfgVal := reflect.ValueOf(hst).Elem()

	for index := 0; index < cfgType.NumField(); index++ {
		tgs := cfgType.Field(index).Tag
		tmpField := cfgVal.Field(index)
		if tmpField.Kind() == reflect.String && tmpField.Interface() == "" {
			tmpField.Set(reflect.ValueOf(tgs.Get("Default")))
		} else if tmpField.Kind() == reflect.Int && tmpField.Interface() == 0 {
			tmp, _ := strconv.Atoi(tgs.Get("Default"))
			tmpField.Set(reflect.ValueOf(tmp))
		}
	}
}

func ReadCfgFileToStruct(hostapd interface{}, fPath string) {
	cfgFile, err := ioutil.ReadFile(fPath)
	if err != nil {
		panic("can't read hostapd config file")
	}

	hostapdType := reflect.TypeOf(hostapd).Elem()
	hostapdVal := reflect.ValueOf(hostapd).Elem()

	if hostapdType.Kind() != reflect.Struct {
		panic("readCfgFile() just accepts struct type")
	}

	for index := 0; index < hostapdType.NumField(); index++ {
		tgs := hostapdType.Field(index).Tag
		r, _ := regexp.Compile(fmt.Sprintf("%s=.+", tgs.Get("Hostapd")))
		fileLine := r.FindString(string(cfgFile))

		valstr := strings.Split(fileLine, "=")[1]

		if hostapdVal.Field(index).Kind() == reflect.String {
			hostapdVal.Field(index).Set(reflect.ValueOf(valstr))
		} else if hostapdVal.Field(index).Kind() == reflect.Int {
			tmpint, _ := strconv.Atoi(valstr)
			hostapdVal.Field(index).Set(reflect.ValueOf(tmpint))
		}
	}
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
