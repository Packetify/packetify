package cmd

import (
	"context"
	"fmt"
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
	startAP.Flags().StringVarP(&virtIfaceName, "virtiface", "v", networkHandler.GetValidVirtIfaceName("hotSpot"), "--virtIface \"hotspot\"")
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
	wlandev := networkHandler.NewWIFI(iface)
	if wlandev.HasAPAndVirtIfaceMode() {
		log.Println("AP mode availabe")
		return
	} else {
		log.Panic("given wifi interface doesn't support AP mode")
	}
}

func (AP *AccessPoint) CreateAP(ctx context.Context, wg *sync.WaitGroup, netShare string) error {

	wifidev := networkHandler.NewWIFI(AP.WifiIface)

	if err := networkHandler.MainNetworkService.EnableIpForwardingIface(wifidev.Interface); err != nil {
		log.Println("Error enabling IP forwarding", err)
		return err
	}
	log.Println("Enable IPForwarding for iface", AP.IfaceName)

	err := networkHandler.IWDeleteInterface(AP.IfaceName)
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
		Options: dhcp4.Options{
			dhcp4.OptionSubnetMask:       AP.IPRange.Mask,
			dhcp4.OptionRouter:           AP.IPRange.IP.To4(), // Presuming Server is also your router
			dhcp4.OptionDomainNameServer: AP.Dns.To4(),        // Presuming Server is also your DNS server
		},
	}

	dhcp4PacketConn, _ := conn.NewUDP4BoundListener(AP.IfaceName, ":67")
	dhcpErr := false
	go func() {
		if err := dhcp4.Serve(dhcp4PacketConn, handler); err != nil {
			log.Println("dhcp server stoped....")
			dhcpErr = true
			return
		}
	}()
	if dhcpErr {
		log.Println("dhcp server stoped....")
		return fmt.Errorf("dhcp server error")
	}

	if err := wifidev.SetupIpToVirtIface(&AP.IPRange, AP.IfaceName); err != nil {
		log.Println("Error setting up IP to virtual interface", err)
		return err
	}

	//hostapd
	testHstapd := hostapd.New(AP.IfaceName, AP.Ssid, AP.Password, 2)
	testHstapd[hostapd.Channel] = 6
	if err := hostapd.WriteCfg(AP.HostapdCFG, testHstapd); err != nil {
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
		if err := CleanUP(netShare, AP, HostapdCmd, dhcp4PacketConn, wifidev); err != nil {
			log.Println("error cleaning up", err)
			return err
		}
		wg.Done()
		return nil
	}
}

func CleanUP(netShare string, AP *AccessPoint, HostapdCmd *exec.Cmd, dhcpPacketConn net.PacketConn, wifidev *networkHandler.WifiDevice) (err error) {
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
	return nil
}
