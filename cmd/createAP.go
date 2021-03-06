package cmd

import (
	"context"
	"github.com/Packetify/ipcalc/ipv4calc"
	"github.com/Packetify/packetify/networkHandler"
	"github.com/Packetify/packetify/networkHandler/dhcp4d"
	"github.com/Packetify/packetify/networkHandler/hostapd"
	"github.com/krolaw/dhcp4"
	"github.com/krolaw/dhcp4/conn"
	"github.com/spf13/cobra"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// TODO: add validation for flags
// TODO: handle daemonization
// TODO: add vpn flag

type AccessPoint struct {
	IfaceName     string
	WifiIface     string
	IPRange       net.IPNet
	Dns           net.IP
	Ssid          string
	Password      string
	HostapdCFG    string
	InternetIface string
}

// startAP represents the start command
var (
	wlanIface     string
	virtIfaceName string
	wlanIPNet     net.IPNet
	ssid          string
	password      string
	dnsServer     net.IP
	netShare      string
	daemon        bool
	driver        string
	isolateClient bool
	channel       int
	countryCode   string
	wpaVersion    int
	hidden        bool
	denyMacFile   string
	acceptMacFile string
	dnsmasq       bool
	dnsblockFile      string
	openvpn string
	powersave bool
	enablevpn bool

	startAP       = &cobra.Command{
		Use:     "createap",
		Short:   "createap packetify access Point",
		Example: "sudo packetify createap -w wlp0s20f3 -n enp7s0 --ssid \"exampleAP\" -p \"12345678\" -d \"1.1.1.1\" --ip \"192.168.214.1/24\"",

		Run: func(cmd *cobra.Command, args []string) {
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)

			validateWlanIface(wlanIface)
			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			wlanIPNet.IP = dhcp4.IPAdd(wlanIPNet.IP, 1)
			myAccessPoint := AccessPoint{
				virtIfaceName,
				wlanIface,
				wlanIPNet,
				dnsServer,
				ssid,
				password,
				"/tmp/hostapd.conf",
				netShare,
			}
			hostapdOptions := []hostapd.HostapdOption{
				{
					hostapd.Interface,
					myAccessPoint.IfaceName,
				},
				{
					hostapd.Ssid,
					myAccessPoint.Ssid,
				},
				{
					hostapd.WPA_PassPhrase,
					myAccessPoint.Password,
				},
				{
					hostapd.WPA,
					wpaVersion,
				},
				{
					hostapd.Driver,
					driver,
				},
				{
					hostapd.Channel,
					channel,
				},
				{
					hostapd.CountryCode,
					countryCode,
				},
			}
			if isolateClient {
				log.Println("isolate clients enabled")
				hostapdOptions = append(hostapdOptions, hostapd.HostapdOption{
					Key:   hostapd.APIsolate,
					Value: "1",
				})
			}
			if hidden {
				log.Println("hidden ssid enabled")
				hostapdOptions = append(hostapdOptions, hostapd.HostapdOption{
					Key:   hostapd.Ignorebrodcast,
					Value: "1",
				})
			}

			if len(denyMacFile) != 0 {
				log.Println("deny mac file enabled")
				hostapdOptions = append(hostapdOptions, hostapd.HostapdOption{
					Key:   hostapd.DenyMacFile,
					Value: denyMacFile,
				})
				hostapdOptions = append(hostapdOptions, hostapd.HostapdOption{
					Key:   hostapd.MacAddrACL,
					Value: "0",
				})
			} else if len(acceptMacFile) != 0 {
				log.Println("accept mac file enabled")
				hostapdOptions = append(hostapdOptions, hostapd.HostapdOption{
					Key:   hostapd.AcceptMacFile,
					Value: acceptMacFile,
				})
				hostapdOptions = append(hostapdOptions, hostapd.HostapdOption{
					Key:   hostapd.MacAddrACL,
					Value: "1",
				})
			}

			go func() {
				wg.Add(1)
				if err := myAccessPoint.CreateAP(ctx, &wg, hostapdOptions); err != nil {
					log.Println("Error creating AP", err)
					panic(err)
				}
			}()
			//time.Sleep(5*time.Second)
			//go packetParser.TotalBandWidthUsage(ctx,myAccessPoint.IfaceName, myAccessPoint.IPRange)
			<-sigs

			cancel()
			wg.Wait()
		},
	}
)

func init() {
	rootCmd.AddCommand(startAP)

	userdomain, _ := os.Hostname()
	_, defaultIPNet, _ := net.ParseCIDR("192.168.100.1/24")

	startAP.Flags().StringVarP(&wlanIface, "wlaniface", "w", "", "set wireless interface name")
	startAP.Flags().StringVarP(&virtIfaceName, "virtiface", "v", networkHandler.GetValidVirtIfaceName("hotSpot"), "set virtual interface name")
	startAP.Flags().IPNetVarP(&wlanIPNet, "ip", "", *defaultIPNet, "set accesspoint ip range")
	startAP.Flags().IPVarP(&dnsServer, "dns", "d", net.IP{1, 1, 1, 1}, "Set DNS returned by DHCP")
	startAP.Flags().StringVarP(&ssid, "ssid", "s", userdomain, "set access point name")
	startAP.Flags().StringVarP(&password, "password", "p", "12345678", "set accesspoint password")
	startAP.Flags().StringVarP(&netShare, "netshare", "n", "false", "enable internet sharing(interface with internet)")
	startAP.Flags().BoolVarP(&daemon, "daemon", "", false, "Run in the background")
	startAP.Flags().StringVarP(&driver, "driver", "", "nl80211", "Choose your WiFi adapter driver")
	startAP.Flags().BoolVarP(&isolateClient, "isolate", "", false, "Disable communication between clients")
	startAP.Flags().IntVarP(&channel, "channel", "", 1, "Channel number")
	startAP.Flags().StringVarP(&countryCode, "country", "", "US", "Set two-letter country code for regularity")
	startAP.Flags().IntVarP(&wpaVersion, "wpa", "", 3, "Use 1 for WPA, use 2 for WPA2, use 1+2 for both")
	startAP.Flags().BoolVarP(&hidden, "hidden", "", false, "Make the Access Point hidden (do not broadcast the SSID)")
	startAP.Flags().StringVarP(&acceptMacFile, "acceptmac", "", "", "Accept lists are read from separate files")
	startAP.Flags().StringVarP(&denyMacFile, "denymac", "", "", "Deny lists are read from separate files")
	startAP.Flags().BoolVarP(&dnsmasq,"dnsmasq","",false,"use dnsmasq as dhcp , dns server")
	startAP.Flags().StringVarP(&dnsblockFile,"dnsblock","","","block dns request of domain/ip in file")
	startAP.Flags().StringVarP(&openvpn,"openvpn","","","run openvpn config pass all traffic throgh vpn")
	startAP.Flags().BoolVarP(&enablevpn,"vpn","",false,"enable clients use vpn")
	startAP.Flags().BoolVarP(&powersave,"powersave","",false,"enable powersaving on interface")


	startAP.MarkFlagRequired("wlaniface")
}

func validateWlanIface(iface string) {
	//New also validates iface existance
	wlandev,err := networkHandler.NewWIFI(iface)
	if err!=nil{
		log.Fatal(err)
	}
	if wlandev.HasAPAndVirtIfaceMode() {
		log.Println("AP mode availabe")
		return
	} else {
		log.Panic("given wifi interface doesn't support AP mode")
	}
}

func (AP *AccessPoint) CreateAP(ctx context.Context, wg *sync.WaitGroup, hostapdOptions []hostapd.HostapdOption) error {

	wifidev,err := networkHandler.NewWIFI(AP.WifiIface)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := networkHandler.MainNetworkService.EnableIpForwardingIface(wifidev.Interface); err != nil {
		log.Println("Error enabling IP forwarding", err)
		return err
	}
	log.Println("Enable IPForwarding for iface", AP.IfaceName)

	err = networkHandler.IWDeleteInterface(AP.IfaceName)
	if err != nil && err != networkHandler.ErrorInterfaceNotExist {
		log.Printf("Error deleting interface %v", err)
		return err
	}
	log.Println("Deleted interface", AP.IfaceName)

	if err := wifidev.IWCreateVirtualIface(AP.IfaceName); err != nil {
		log.Println("Error creating virtual interface", err)
		return err
	}
	log.Println("Created virtual interface", AP.IfaceName)

	if err := networkHandler.UnmanageIface(AP.IfaceName); err != nil {
		log.Println("Error unmanaging interface", err)
		return err
	}

	ipcalc := ipv4calc.New(AP.IPRange)

	//dhcpServer
	handler := &dhcp4d.DHCPHandler{
		IP:            AP.IPRange.IP.To4(),
		LeaseDuration: 5 * time.Hour,
		Start:         ipcalc.GetMinHost(),
		LeaseRange:    ipcalc.GetValidHosts(),
		Leases:        make(map[int]dhcp4d.Lease, 10),
		DevicesChan:   make(chan dhcp4d.DeviceInfo),
		Options: dhcp4.Options{
			dhcp4.OptionSubnetMask:       AP.IPRange.Mask,
			dhcp4.OptionRouter:           AP.IPRange.IP.To4(), // Presuming Server is also your router
			dhcp4.OptionDomainNameServer: AP.Dns.To4(),        // Presuming Server is also your DNS server
		},
	}

	dhcp4PacketConn, _ := conn.NewUDP4BoundListener(AP.IfaceName, ":67")
	go func() {
		if err := dhcp4.Serve(dhcp4PacketConn, handler); err != nil {
			log.Println("dhcp server stoped....")
			close(handler.DevicesChan)
			return
		}
	}()
	go func() {
		for {
			select {
			case dev := <-handler.DevicesChan:
				log.Println(dev.HostName, dev.IPAddr, dev.MacAddr)
			case <-ctx.Done():
				log.Println("Stoping dhcp server and user log")
				return

			}
		}
	}()

	if err := wifidev.SetupIpToVirtIface(&AP.IPRange, AP.IfaceName); err != nil {
		log.Println("Error setting up IP to virtual interface", err)
		return err
	}

	//hostapd
	Hstapd := hostapd.New(hostapdOptions...)
	if err := hostapd.WriteCfg(AP.HostapdCFG, Hstapd); err != nil {
		log.Println("Error writing hostapd config file", err)
		return err
	}

	HostapdCmd, err := hostapd.Run(AP.HostapdCFG, false)
	if err != nil {
		return err
	}
	if netShare != "false" {
		err = networkHandler.EnableInternetSharing(AP.IfaceName, AP.InternetIface, AP.IPRange, true)
		if err != nil {
			log.Println("error Enable internet sharing", err)
			return err
		}
	}

	select {
	case <-ctx.Done():
		log.Println("ap stopped...")
		if err := AP.CleanupAP(HostapdCmd, dhcp4PacketConn, wifidev); err != nil {
			log.Println("error cleaning up", err)
			return err
		}
		wg.Done()
		return nil
	}
}

func (AP *AccessPoint) CleanupAP(HostapdCmd *exec.Cmd, dhcpPacketConn net.PacketConn,
	wifidev *networkHandler.WifiDevice) (err error) {
	log.Println("clean up")
	if netShare != "false" {
		err = networkHandler.DisableInternetSharing(AP.IfaceName, AP.InternetIface, AP.IPRange)
		if err != nil {
			log.Println("error Disable internet sharing", err)
			return err
		}
		log.Println("Disable internet sharing")
	}
	if err = HostapdCmd.Process.Kill(); err != nil {
		log.Println("error killing hostapd", err)
		return err
	}
	log.Println("close hostapd process")

	if err = dhcpPacketConn.Close(); err != nil {
		log.Println("error closing dhcp server", err)
		return err
	}

	if err = wifidev.IWDeleteVirtualIface(AP.IfaceName); err != nil {
		log.Println("error deleting virtual interface", err)
		return err
	}
	log.Printf("Delete virtual interfaces on %v", AP.IfaceName)

	err = networkHandler.IWDeleteInterface(AP.IfaceName)
	if err != nil && err != networkHandler.ErrorInterfaceNotExist {
		log.Println("error deleting interface", err)
		return err
	}
	log.Printf("Delete interface %v", AP.IfaceName)

	if err = networkHandler.MainNetworkService.DisableIpForwardingIface(wifidev.Interface); err != nil {
		log.Println("error disabling ip forwarding", err)
		return err
	}
	log.Println("Disable IPForwarding")
	if err = hostapd.RemoveConfigFile(AP.HostapdCFG); err != nil {
		log.Println("error removing hostapd config file", err)
		return err
	}
	log.Println("hostapd config file Removed")
	return nil
}
