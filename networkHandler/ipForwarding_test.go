package networkHandler

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"testing"
)

func TestNetworkService_IpForwardingStatus(t *testing.T) {

	tests := map[string]struct {
		want    bool
		errWant error
		runCmd  *exec.Cmd
	}{
		"true IpForwardingStatus": {
			want:   true,
			runCmd: exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1"),
		},
		"false IpForwardingStatus": {
			want:   false,
			runCmd: exec.Command("sysctl", "-w", "net.ipv4.ip_forward=0"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ns := &NetworkService{}
			if test.runCmd != nil {
				log.Println(test.runCmd.String())
				err := test.runCmd.Run()
				if test.errWant == nil {
					if err != nil {
						t.Errorf("Error: %v", err)
					}
				} else {
					if err != test.errWant {
						t.Errorf("Error: %v", err)
					}
				}
			}
			if got := ns.IpForwardingStatus(); got != test.want {
				t.Errorf("IpForwardingStatus() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNetworkService_IpForwardingStatusIface(t *testing.T) {
	ifaces, _ := net.Interfaces()
	tests := map[string]struct {
		want         bool
		errWant      error
		runCmd       string
		networkIface net.Interface
	}{
		"true IpForwardingStatus Iface": {
			networkIface: ifaces[0],
			want:         true,
			runCmd:       "echo 1 > /proc/sys/net/ipv4/conf/%s/forwarding",
		},
		"false IpForwardingStatus Iface": {
			networkIface: ifaces[0],
			want:         false,
			runCmd:       "echo 0 > /proc/sys/net/ipv4/conf/%s/forwarding",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ns := &NetworkService{}
			if len(test.runCmd) != 0 {
				cmd := fmt.Sprintf(test.runCmd, test.networkIface.Name)
				err := exec.Command("sh", "-c", cmd).Run()
				log.Println(cmd)
				if test.errWant == nil {
					if err != nil {
						t.Errorf("Error: %v", err)
					}
				} else if test.errWant != err {
					t.Errorf("Error: %v", err)
				}
			}
			if got := ns.IpForwardingStatusIface(test.networkIface); got != test.want {
				t.Errorf("IpForwardingStatus() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNetworkService_EnableIpForwarding(t *testing.T) {
	tests := map[string]struct {
		errWant error
		want    bool
	}{
		"Enable IPForwarding": {
			want: true,
		},
		"Enable IPForwarding when is enabled": {
			want: true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ns := &NetworkService{}
			err := ns.EnableIpForwarding()
			if test.errWant == nil {
				if err != nil {
					t.Errorf("Error: %v", err)
				}
				res := ns.IpForwardingStatus()
				if res != test.want {
					t.Errorf("IpForwardingStatus() = %v, want %v", res, test.want)
				}

			} else {
				if err != test.errWant {
					t.Errorf("Error: %v", err)
				}
			}

		})
	}
}

func TestNetworkService_EnableIpForwardingIface(t *testing.T) {
	ifaces, _ := net.Interfaces()
	tests := map[string]struct {
		errwant      error
		networkIface net.Interface
		runCmd       string
		want         bool
	}{
		"Enable IPForwarding iface": {
			networkIface: ifaces[0],
			runCmd:       "echo 0 > /proc/sys/net/ipv4/conf/%s/forwarding",
			want:         true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ns := &NetworkService{}
			if len(test.runCmd) != 0 {
				cmd := fmt.Sprintf(test.runCmd, test.networkIface.Name)
				err := exec.Command("sh", "-c", cmd).Run()
				if err != nil {
					t.Errorf("Error: %v", err)
				}
			}
			err := ns.EnableIpForwardingIface(test.networkIface)
			if test.errwant == nil {
				if err != nil {
					t.Errorf("Error: %v", err)
				}
				res := ns.IpForwardingStatusIface(test.networkIface)
				if res != test.want {
					t.Errorf("IpForwardingStatus() = %v, want %v", res, test.want)
				}

			} else {
				if err != test.errwant {
					t.Errorf("Error: %v", err)
				}
			}
		})
	}
}

func TestNetworkService_DisableIpForwarding(t *testing.T) {
	tests := map[string]struct {
		errWant error
		want    bool
	}{
		"Disable IPForwarding": {
			want: false,
		},
		"Disable IPForwarding when is Disable": {
			want: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ns := &NetworkService{}
			err := ns.DisableIpForwarding()
			if test.errWant == nil {
				if err != nil {
					t.Errorf("Error: %v", err)
				}
				res := ns.IpForwardingStatus()
				if res != test.want {
					t.Errorf("IpForwardingStatus() = %v, want %v", res, test.want)
				}

			} else {
				if err != test.errWant {
					t.Errorf("Error: %v", err)
				}
			}
		})
	}
}

func TestNetworkService_DisableIpForwardingIface(t *testing.T) {
	ifaces, _ := net.Interfaces()
	tests := map[string]struct {
		errwant      error
		networkIface net.Interface
		runCmd       string
		want         bool
	}{
		"Disable IPForwarding iface": {
			networkIface: ifaces[0],
			runCmd:       "echo 1 > /proc/sys/net/ipv4/conf/%s/forwarding",
			want:         true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ns := &NetworkService{}
			if len(test.runCmd) != 0 {
				cmd := fmt.Sprintf(test.runCmd, test.networkIface.Name)
				err := exec.Command("sh", "-c", cmd).Run()
				if err != nil {
					t.Errorf("Error: %v", err)
				}
			}
			err := ns.DisableIpForwardingIface(test.networkIface)
			if test.errwant == nil {
				if err != nil {
					t.Errorf("Error: %v", err)
				}
				res := ns.IpForwardingStatusIface(test.networkIface)
				if res != test.want {
					t.Errorf("IpForwardingStatus() = %v, want %v", res, test.want)
				}

			} else {
				if err != test.errwant {
					t.Errorf("Error: %v", err)
				}
			}
		})
	}
}