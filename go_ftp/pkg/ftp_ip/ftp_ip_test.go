package ftp_ip_test

import (
	"errors"
	"testing"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_ip"
	"github.com/jakobsvenningsson/go_ftp/pkg/test_utils"
)

var encodeTests = []struct {
	inputIP     string
	inputPort   string
	expected    string
	expectedErr error
}{
	{"127.0.0.1", "513", "127,0,0,1,2,1", nil},
	{"127.0.0.1", "21", "127,0,0,1,0,21", nil},
	{"127.0.0.1", "", "", errors.New("Invalid addr format")},
	{"0.0.1", "1234", "", errors.New("Invalid addr format")},
	{"127.0.0.1", "b", "", errors.New("Invalid addr format")},
}

var decodeTests = []struct {
	input       string
	expected    string
	expectedErr error
}{
	{"127,0,0,1,2,1", "127.0.0.1:513", nil},
	{"127,0,0,1,2,0", "127.0.0.1:512", nil},
	{"127,0,0,1,2", "", errors.New("Invalid addr format")},
	{"1,2,,4,5,6", "", errors.New("Invalid addr format")},
}

func TestFtpIPEncode(t *testing.T) {
	for _, test := range encodeTests {
		encoded, err := ftp_ip.Encode(test.inputIP, test.inputPort)
		if ok, have, want := test_utils.VerifyError(err, test.expectedErr); !ok {
			t.Errorf("Error actual = %v, and Expected = %v.", have, want)
		}
		if encoded != test.expected {
			t.Errorf("Error actual = %v, and Expected = %v.", encoded, test.expected)
		}
	}
}

func TestFtpIPDecode(t *testing.T) {
	for _, test := range decodeTests {
		encoded, err := ftp_ip.Decode(test.input)
		if ok, have, want := test_utils.VerifyError(err, test.expectedErr); !ok {
			t.Errorf("Error actual = %v, and Expected = %v.", have, want)
		}
		if encoded != test.expected {
			t.Errorf("Error actual = %v, and Expected = %v.", encoded, test.expected)
		}
	}
}
