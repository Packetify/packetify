package networkHandler

import (
	"errors"
	"testing"
)

func TestValidateIfaceName(t *testing.T) {
	funcErr := errors.New("given network Iface name is invalid")
	tests := []struct {
		ifaceName string
		want      error
	}{
		{"hotspot", nil},
		{"_jasd", funcErr},
		{"asd*", funcErr},
		{"asd_asd", funcErr},
		{"hotspot541", nil},
		{"124sadsd", funcErr},
		{"/ad", funcErr},
		{"", funcErr},
		{"@asd", funcErr},
		{"<kalsdj", funcErr},
		{"asd?ad", funcErr},
		{"sa21ad5", nil},
	}

	for _, tst := range tests {
		err := ValidateIfaceName(tst.ifaceName)
		if (tst.want == nil && err != nil) || (tst.want != nil && err == nil) {
			t.Errorf("ValidateIfaceName(%s)=%v", tst.ifaceName, err)
		}

	}
}
