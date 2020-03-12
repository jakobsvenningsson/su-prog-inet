package ftp_client_test

import (
	"errors"
	"log"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_client"
	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_server"
	"github.com/jakobsvenningsson/go_ftp/pkg/test_utils"
)

func TestFTPClientAuth(t *testing.T) {
	srv := startFTPServer("/tmp", "127.0.0.1", "8999")
	client, conn := startFTPClient("/tmp")
	defer conn.Close()

	var tests = []struct {
		user        string
		pass        string
		expectedErr error
	}{
		{"demo", "password", nil},
		{"demo", "wrong_password", errors.New("Expected status code 230 when sending user to server but received status 530")},
		{"wrong_user", "wrong_password", errors.New("Expected status code 230 when sending user to server but received status 530")},
	}

	for _, test := range tests {
		err := client.Authenticate(test.user, test.pass)
		if ok, have, want := test_utils.VerifyError(err, test.expectedErr); !ok {
			t.Errorf("Error actual = %v, and Expected = %v.", have, want)
		}
	}
	srv.Stop()
}

func startFTPServer(root, ip, port string) *ftp_server.FtpServer {
	srv := ftp_server.New(root, ip, port)
	go func() {
		srv.Start()
	}()
	time.Sleep(time.Millisecond * 100)
	return srv
}

func startFTPClient(outDir string) (*ftp_client.FtpClient, net.Conn) {
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		log.Fatal(err)
	}
	in := strings.NewReader("")
	client, err := ftp_client.New(in, conn, "127.0.0.1", outDir)
	if err != nil {
		log.Fatal(err)
	}
	client.ReadWelcomeMessage()
	return client, conn
}
