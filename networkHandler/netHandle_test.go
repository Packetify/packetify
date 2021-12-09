package networkHandler

import (
	"testing"
)
var ns = NetworkService{}
func TestIsCommandAvailable(t *testing.T) {
	var tests = []struct {
		command string
		want    bool
	}{
		{"", false},
		{"iw", true},
		{"iwlist", true},
		{"ip", true},
		{"iptables", true},
		{"ls", true},
		{"hostapd", true},
		{"asdadasdasd", false},
		{"mkdir", true},
		{"rm", true},
		{"rmdir", true},
	}

	for _, test := range tests {
		if got := ns.WhichCommand(test.command); got != test.want {
			t.Errorf("IsCommandAvailable(%s)=%v", test.command, got)
		}
	}
}
