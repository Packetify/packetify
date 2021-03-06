package networkHandler

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
)

func EnableDnsServer(ipRange net.IPNet, port uint16) error {

	commands := []string{
		fmt.Sprintf("-w -I INPUT -p tcp -m tcp --dport %d -j ACCEPT", port),
		fmt.Sprintf("-w -I INPUT -p udp -m udp --dport %d -j ACCEPT", port),
		fmt.Sprintf("-w -t nat -I PREROUTING -s %s -d %s -p tcp -m tcp --dport 53 -j REDIRECT --to-ports %d",
			ipRange.String(), ipRange.IP.String(), port),
		fmt.Sprintf("-w -t nat -I PREROUTING -s %s -d %s -p udp -m udp --dport 53 -j REDIRECT --to-ports %d",
			ipRange.String(), ipRange.IP.String(), port),
	}

	for _, command := range commands {
		cmd := exec.Command("iptables", strings.Split(command, " ")...)
		log.Println(cmd.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil

}

func EnableInternetSharing(iface string, netSahreIface string, ipRange net.IPNet, flush bool) error {
	if flush {
		if err := IPTablesFlash(); err != nil {
			return err
		}
	}
	commands := []string{
		fmt.Sprintf("-w -t nat -I POSTROUTING -s %s ! -o %s -j MASQUERADE", ipRange.String(), iface),
		fmt.Sprintf("-w -I FORWARD -i %s -s %s -j ACCEPT", iface, ipRange.String()),
		fmt.Sprintf("-w -I FORWARD -i %s -d %s -j ACCEPT", netSahreIface, ipRange.String()),
		fmt.Sprintf("-w -I INPUT -p udp -m udp --dport 67 -j ACCEPT"),
	}

	for _, command := range commands {
		cmd := exec.Command("iptables", strings.Split(command, " ")...)
		log.Println(cmd.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func IPTablesFlash() error {

	commandArgs := []string{
		"-t nat -F",
		"-F",
	}
	for _, command := range commandArgs {
		if err := exec.Command("iptables", strings.Split(command, " ")...).Run(); err != nil {
			return err
		}
	}
	return nil
}

func DisableInternetSharing(iface string, netSahreIface string, ipRange net.IPNet) error {
	commands := []string{
		fmt.Sprintf("-w -t nat -D POSTROUTING -s %s ! -o %s -j MASQUERADE", ipRange.String(), iface),
		fmt.Sprintf("-w -D FORWARD -i %s -s %s -j ACCEPT", iface, ipRange.String()),
		fmt.Sprintf("-w -D FORWARD -i %s -d %s -j ACCEPT", netSahreIface, ipRange.String()),
	}
	for _, command := range commands {
		cmd := exec.Command("iptables", strings.Split(command, " ")...)
		log.Println(cmd.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func DisableDnsServer(ipRange net.IPNet, port uint16) error {
	commands := []string{
		fmt.Sprintf("-w -D INPUT -p tcp -m tcp --dport %s -j ACCEPT", port),
		fmt.Sprintf("-w -D INPUT -p udp -m udp --dport %s -j ACCEPT", port),
		fmt.Sprintf("-w -t nat -D PREROUTING -s %s -d %s -p tcp -m tcp --dport 53 -j REDIRECT --to-ports %s",
			ipRange.String(), ipRange.IP.String(), port),
		fmt.Sprintf("-w -t nat -D PREROUTING -s %s -d %s -p udp -m udp --dport 53 -j REDIRECT --to-ports %s",
			ipRange.String(), ipRange.IP.String(), port),
		fmt.Sprintf("-w -D INPUT -p udp -m udp --dport 67 -j ACCEPT"),
	}
	for _, command := range commands {
		cmd := exec.Command("iptables", strings.Split(command, " ")...)
		log.Println(cmd.String())
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
