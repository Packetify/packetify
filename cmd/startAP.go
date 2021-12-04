package cmd

import (
	"context"
	"github.com/Packetify/ipcalc/ipv4calc"
	"github.com/Packetify/packetify/networkHandler"
	"github.com/Packetify/packetify/networkHandler/dhcp4d"
	"github.com/Packetify/packetify/networkHandler/hostapd"
	"github.com/Packetify/packetify/networkHandler/ipForward"
	"github.com/Packetify/packetify/networkHandler/networkManager"
	"github.com/Packetify/packetify/networkHandler/wifi"
	"github.com/krolaw/dhcp4"
	"github.com/krolaw/dhcp4/conn"
	"github.com/spf13/cobra"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

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
	startAP       = &cobra.Command{
		Use:   "start -w wlan0 -n eth0 --ssid \"APName\" -p \"12345678\"",
		Short: "start packetify access Point",

		Run: func(cmd *cobra.Command, args []string) {
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
				"/tmp/hostapdCruella.conf",
				netShare,
			}
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)

			go func() {
				wg.Add(1)
				if err := myAccessPoint.CreateAP(ctx, &wg, netShare); err != nil {
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

	startAP.Flags().StringVarP(&wlanIface, "wlaniface", "w", "", "--wlaniface \"wlan0\"")
	startAP.Flags().StringVarP(&virtIfaceName, "virtiface", "v", wifi.GetValidVirtIfaceName("hotSpot"), "--virtIface \"hotspot\"")
	startAP.Flags().IPNetVarP(&wlanIPNet, "ip", "", *defaultIPNet, "--ip ip")
	startAP.Flags().IPVarP(&dnsServer, "dns", "d", net.IP{1, 1, 1, 1}, "-dns \"8.8.8.8\"")
	startAP.Flags().StringVarP(&ssid, "ssid", "s", userdomain, "--ssid \"access point name\"")
	startAP.Flags().StringVarP(&password, "password", "p", "12345678", "--pasword \"securepass123\"")
	startAP.Flags().StringVarP(&netShare, "netshare", "n", "false", "--netshare \"eth0\"")
	startAP.Flags().BoolVarP(&daemon, "daemon", "", false, "--daemon")

	startAP.MarkFlagRequired("wlaniface")
}

func validateWlanIface(iface string) {
	//New also validates iface existance
	wlandev := wifi.New(iface)
	if wlandev.HasAPAndVirtIfaceMode() {
		log.Println("AP mode availabe")
		return
	} else {
		log.Panic("given wifi interface doesn't support AP mode")
	}
}

func (AP AccessPoint) CreateAP(ctx context.Context, wg *sync.WaitGroup, netShare string) error {

	wifidev := wifi.New(AP.WifiIface)

	ipForward.EnableIpForwardingIface(wifidev.Interface)

	wifi.DeleteInterface(AP.IfaceName)
	if err := wifidev.CreateVirtualIface(AP.IfaceName); err != nil {
		return err
	}

	if err := networkManager.UnmanageIface(AP.IfaceName); err != nil {
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
		Options: dhcp4.Options{
			dhcp4.OptionSubnetMask:       AP.IPRange.Mask,
			dhcp4.OptionRouter:           AP.IPRange.IP.To4(), // Presuming Server is also your router
			dhcp4.OptionDomainNameServer: AP.Dns.To4(),        // Presuming Server is also your DNS server
		},
	}

	pc, _ := conn.NewUDP4BoundListener(AP.IfaceName, ":67")

	go func() {
		if err := dhcp4.Serve(pc, handler); err != nil {
			log.Println("dhcp server stoped....")
			return
		}
	}()

	if err := wifidev.SetupIpToVirtIface(&AP.IPRange, AP.IfaceName); err != nil {
		return err
	}

	//hostapd
	testHstapd := hostapd.New(AP.IfaceName, AP.Ssid, AP.Password, 2)
	testHstapd[hostapd.Channel] = 6
	hostapd.WriteCfg(AP.HostapdCFG, testHstapd)

	cmd, err := hostapd.Run(AP.HostapdCFG, false)
	if err != nil {
		return err
	}
	if netShare != "false" {
		err = networkHandler.EnableInternetSharing(AP.IfaceName, AP.InternetIface, AP.IPRange, true)
		if err != nil {
			return err
		}
	}

	select {
	case <-ctx.Done():
		log.Println("ap stopped...")
		networkHandler.IPTablesFlash()
		cmd.Process.Kill()
		pc.Close()
		wifi.DeleteInterface(AP.IfaceName)
		wifidev.DeleteVirtualIface(AP.IfaceName)
		ipForward.DisableIpForwardingIface(wifidev.Interface)
		wg.Done()
		return nil
	}
}
