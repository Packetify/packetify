package hostapd

import (
	"errors"
	"fmt"
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

func WriteToCfgFile(hostapdCFG interface{}, fpath string) {
	cfgType := reflect.TypeOf(hostapdCFG)
	cfgVal := reflect.ValueOf(hostapdCFG)
	if cfgType.Kind() != reflect.Struct {
		panic("to write Hostapd config file, struct type needed")
	}

	var reSTR string
	for index := 0; index < cfgType.NumField(); index++ {
		tgs := cfgType.Field(index).Tag
		if tgs.Get("Hostapd") != "" {
			reSTR += fmt.Sprintf("%s=%v\n", tgs.Get("Hostapd"), cfgVal.Field(index).Interface())
		}
	}
	ioutil.WriteFile(fpath, []byte(reSTR), 0644)
}

func (hst *HostapdBase) Validate() {
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

func ReadCfgFile(fPath string,) *HostapdBase{
	hstapd := &HostapdBase{}
	cfgFile, err := ioutil.ReadFile(fPath)
	if err != nil {
		panic("can't read hostapd config file")
	}

	hostapdType := reflect.TypeOf(hstapd).Elem()
	hostapdVal := reflect.ValueOf(hstapd).Elem()

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
	return hstapd
}

func Run(cfgPath string)error{
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return errors.New("config file not exist can't run hostapd")
	}
	cmd := exec.Command("hostapd",cfgPath)
	if err := cmd.Run() ;err!=nil{
		return err
	}
	return nil
}