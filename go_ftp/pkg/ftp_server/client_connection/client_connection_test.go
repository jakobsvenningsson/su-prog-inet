package client_connection_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_cmd"
	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_ip"
	"github.com/jakobsvenningsson/go_ftp/pkg/ftp_server/client_connection"
	"github.com/jakobsvenningsson/go_ftp/pkg/test_utils"
)

const root = "/tmp/test_dir"

func TestClientConnection(t *testing.T) {
	var tests = []struct {
		input       []ftp_cmd.Cmd
		expected    [][]byte
		expectedErr error
	}{
		// Auth Tests
		{
			[]ftp_cmd.Cmd{
				ftp_cmd.Cmd{Type: ftp_cmd.USER, Arg: "user"},
				ftp_cmd.Cmd{Type: ftp_cmd.PASS, Arg: "wrong_password"},
			},
			[][]byte{
				[]byte("331 Password required for user.\n"),
				[]byte("530 Login failed.\n"),
			}, nil,
		},
		{
			[]ftp_cmd.Cmd{
				ftp_cmd.Cmd{Type: ftp_cmd.USER, Arg: "wrong_user"},
				ftp_cmd.Cmd{Type: ftp_cmd.USER, Arg: "user"},
				ftp_cmd.Cmd{Type: ftp_cmd.PASS, Arg: "pass"},
			},
			[][]byte{
				[]byte("331 Password required for wrong_user.\n"),
				[]byte("331 Password required for user.\n"),
				[]byte("230 User logged in.\n"),
			}, nil,
		},
		{
			[]ftp_cmd.Cmd{
				ftp_cmd.Cmd{Type: ftp_cmd.USER, Arg: "wrong_user"},
				ftp_cmd.Cmd{Type: ftp_cmd.PASS, Arg: "pass"},
			},
			[][]byte{
				[]byte("331 Password required for wrong_user.\n"),
				[]byte("530 Login failed.\n"),
			}, nil,
		},
		{
			[]ftp_cmd.Cmd{
				ftp_cmd.Cmd{Type: ftp_cmd.USER, Arg: "user"},
				ftp_cmd.Cmd{Type: ftp_cmd.PASS, Arg: "pass"},
			},
			[][]byte{
				[]byte("331 Password required for user.\n"),
				[]byte("230 User logged in.\n"),
			}, nil,
		},
		// Path tests
		{
			[]ftp_cmd.Cmd{
				ftp_cmd.Cmd{Type: ftp_cmd.PWD, Arg: ""},
			},
			[][]byte{
				[]byte("257 \"/\" is current directory.\n"),
			}, nil,
		},
		{
			[]ftp_cmd.Cmd{
				ftp_cmd.Cmd{Type: ftp_cmd.PWD, Arg: ""},
				ftp_cmd.Cmd{Type: ftp_cmd.CWD, Arg: "1"},
				ftp_cmd.Cmd{Type: ftp_cmd.PWD, Arg: ""},
				ftp_cmd.Cmd{Type: ftp_cmd.CWD, Arg: ".."},
				ftp_cmd.Cmd{Type: ftp_cmd.PWD, Arg: ""},
				ftp_cmd.Cmd{Type: ftp_cmd.CWD, Arg: "/1"},
				ftp_cmd.Cmd{Type: ftp_cmd.PWD, Arg: ""},
				ftp_cmd.Cmd{Type: ftp_cmd.CWD, Arg: "."},
				ftp_cmd.Cmd{Type: ftp_cmd.PWD, Arg: ""},
				ftp_cmd.Cmd{Type: ftp_cmd.CWD, Arg: "/invalid_path"},
			},
			[][]byte{
				[]byte("257 \"/\" is current directory.\n"),
				[]byte("250 CWD command successful.\n"),
				[]byte("257 \"/1\" is current directory.\n"),
				[]byte("250 CWD command successful.\n"),
				[]byte("257 \"/\" is current directory.\n"),
				[]byte("250 CWD command successful.\n"),
				[]byte("257 \"/1\" is current directory.\n"),
				[]byte("250 CWD command successful.\n"),
				[]byte("257 \"/1\" is current directory.\n"),
				[]byte("550 Invalid path.\n"),
			}, nil,
		},
		// PASV, EPSV, PORT
		{
			[]ftp_cmd.Cmd{
				ftp_cmd.Cmd{Type: ftp_cmd.PASV, Arg: ""},
				ftp_cmd.Cmd{Type: ftp_cmd.EPSV, Arg: ""},
				ftp_cmd.Cmd{Type: ftp_cmd.PORT, Arg: "127,0,0,1,39,17"},
			},
			[][]byte{
				[]byte("227 Entering Passive Mode (127,0,0,1,"),
				[]byte("229 Entering Extended Passive Mode (|||"),
				[]byte("200 PORT command successful.\n"),
			}, nil,
		},
	}

	cc, buf, authCh := initCC()
	defer close(authCh)
	for _, test := range tests {
		for i, cmd := range test.input {
			expected := test.expected[i]
			err := cc.Reply(&cmd)
			if ok, want, have := test_utils.VerifyError(err, test.expectedErr); !ok {
				t.Errorf("Error actual = %v, and Expected = %v.", have, want)
			}
			if !bytes.Contains(buf.Bytes(), expected) {
				t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
					strings.TrimSuffix(string(expected), "\n"))
			}
			buf.Reset()
		}
	}
}

func TestListPASV(t *testing.T) {
	cc, buf, authCh := initCC()
	defer close(authCh)
	authenticate(cc, buf)
	cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.PASV, Arg: ""})
	addr, err := ftp_ip.Decode(string(buf.Bytes()))
	if err != nil {
		log.Fatal(err)
	}
	buf.Reset()

	var wg sync.WaitGroup
	wg.Add(1)

	dialDataConn(addr, &wg, func(conn net.Conn) {
		result, err := ioutil.ReadAll(conn)
		if err != nil {
			log.Fatal(err)
		}
		exp, _ := exec.Command("ls", "-l", root).Output()
		if !bytes.Equal(result, exp) {
			t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(result), "\n"),
				strings.TrimSuffix(string(exp), "\n"))
		}
	})

	buf.Reset()

	expected := []byte("150 Opening ASCII mode data connection for file list.\n226 Transfer complete.\n")
	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.LIST, Arg: ""})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	wg.Wait()
}

type connAction func(conn net.Conn)

func TestListACTIVE(t *testing.T) {
	cc, buf, authCh := initCC()
	defer close(authCh)
	authenticate(cc, buf)

	addr, port := "127.0.0.1", "10001"
	encodedAddr, err := ftp_ip.Encode(addr, port)
	if err != nil {
		log.Fatal(err)
	}
	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.PORT, Arg: encodedAddr})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	expected := []byte("200 PORT command successful.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	buf.Reset()

	var wg sync.WaitGroup
	wg.Add(1)

	listenDataConn(addr+":"+port, &wg, func(conn net.Conn) {
		result, err := ioutil.ReadAll(conn)
		if err != nil {
			log.Fatal(err)
		}
		expected, _ := exec.Command("ls", "-l", root).Output()
		if !bytes.Equal(result, expected) {
			t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(result), "\n"),
				strings.TrimSuffix(string(expected), "\n"))
		}
	})

	expected = []byte("150 Opening ASCII mode data connection for file list.\n226 Transfer complete.\n")
	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.LIST, Arg: ""})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	wg.Wait()
}

func TestDele(t *testing.T) {
	cc, buf, authCh := initCC()
	defer close(authCh)
	authenticate(cc, buf)
	buf.Reset()

	fileName := root + "/1/t1"
	if err := ioutil.WriteFile(fileName, []byte("Hello, World!"), 0644); err != nil {
		log.Fatal(err)
	}

	err := cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.DELE, Arg: "/1/t1"})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	expected := []byte("200 DELE command successful.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	buf.Reset()

	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		t.Errorf("File %s not deleted", fileName)
	}

	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.DELE, Arg: "/1/t1/non_exist"})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	expected = []byte("550 File not found.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	buf.Reset()

	if err := ioutil.WriteFile(fileName, []byte("Hello, World!"), 0644); err != nil {
		log.Fatal(err)
	}
	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.CWD, Arg: "./1"})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	buf.Reset()
	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.DELE, Arg: "t1"})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	expected = []byte("200 DELE command successful.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		t.Errorf("File %s not deleted", fileName)
	}
	buf.Reset()

	if err := ioutil.WriteFile(fileName, []byte("Hello, World!"), 0644); err != nil {
		log.Fatal(err)
	}
	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.DELE, Arg: "/2/t1"})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	expected = []byte("550 File not found.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		t.Errorf("File %s not deleted", fileName)
	}
	buf.Reset()

	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.DELE, Arg: "/1/t1"})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	expected = []byte("200 DELE command successful.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	buf.Reset()
}

func TestStorPASV(t *testing.T) {
	cc, buf, authCh := initCC()
	defer close(authCh)
	authenticate(cc, buf)

	cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.PASV, Arg: ""})
	addr, err := ftp_ip.Decode(string(buf.Bytes()))
	if err != nil {
		log.Fatal(err)
	}
	buf.Reset()

	var wg sync.WaitGroup
	wg.Add(1)

	filename := "filename.txt"
	filedata := []byte("Hello, world!")

	dialDataConn(addr, &wg, func(conn net.Conn) {
		_, err = conn.Write(filedata)
		if err != nil {
			log.Fatal(err)
		}
	})

	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.STOR, Arg: filename})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	expected := []byte("150 Opening ASCII mode data connection for file.\n226 Transfer complete.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	wg.Wait()

	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Errorf("File %s not created.", filename)
	}

	file, err := os.Open(root + "/" + filename)
	if err != nil {
		log.Fatal(err)
	}
	actualFileData, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	if !bytes.Equal(actualFileData, filedata) {
		t.Errorf("Error actual = %s, and Expected = %s.", string(actualFileData), string(filedata))
	}

	buf.Reset()

	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.DELE, Arg: filename})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	expected = []byte("200 DELE command successful.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	if _, err := os.Stat(root + "/" + filename); !os.IsNotExist(err) {
		t.Errorf("File %s not deleted", "test_dir/"+filename)
	}
	buf.Reset()
}

func TestStorACTIVE(t *testing.T) {
	cc, buf, authCh := initCC()
	defer close(authCh)
	authenticate(cc, buf)

	addr, port := "127.0.0.1", "10001"
	encodedAddr, err := ftp_ip.Encode(addr, port)
	if err != nil {
		log.Fatal(err)
	}
	cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.PORT, Arg: encodedAddr})
	expected := []byte("200 PORT command successful.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	buf.Reset()

	filedata := []byte("Hello, World!")
	filename := "filename.txt"

	var wg sync.WaitGroup
	wg.Add(1)

	listenDataConn(addr+":"+port, &wg, func(conn net.Conn) {
		conn.Write(filedata)
	})

	expected = []byte("150 Opening ASCII mode data connection for file.\n226 Transfer complete.\n")
	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.STOR, Arg: filename})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	buf.Reset()
	wg.Wait()

	if _, err := os.Stat(root + "/" + filename); os.IsNotExist(err) {
		t.Errorf("File %s not created.", filename)
	}

	file, err := os.Open(root + "/" + filename)
	if err != nil {
		log.Fatal(err)
	}
	actualFileData, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}
	if !bytes.Equal(actualFileData, filedata) {
		t.Errorf("Error actual = %s, and Expected = %s.", string(actualFileData), string(filedata))
	}

	buf.Reset()

	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.DELE, Arg: filename})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	expected = []byte("200 DELE command successful.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	if _, err := os.Stat(root + "/" + filename); !os.IsNotExist(err) {
		t.Errorf("File %s not deleted", "test_dir/"+filename)
	}
	buf.Reset()
}

func TestRetrPASV(t *testing.T) {
	cc, buf, authCh := initCC()
	defer close(authCh)
	authenticate(cc, buf)

	cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.PASV, Arg: ""})
	addr, err := ftp_ip.Decode(string(buf.Bytes()))
	if err != nil {
		log.Fatal(err)
	}
	buf.Reset()

	var wg sync.WaitGroup
	wg.Add(1)

	dialDataConn(addr, &wg, func(conn net.Conn) {
		result, err := ioutil.ReadAll(conn)
		if err != nil {
			log.Fatal(err)
		}
		exp := []byte("Hello, World!")
		if !bytes.Equal(result, exp) {
			t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(result), "\n"),
				strings.TrimSuffix(string(exp), "\n"))
		}
		conn.Close()
	})

	expected := []byte("150 Opening ASCII mode data connection.\n226 Transfer complete.\n")
	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.RETR, Arg: "test_file"})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	wg.Wait()
}

func TestRetrACTIVE(t *testing.T) {
	cc, buf, authCh := initCC()
	defer close(authCh)
	authenticate(cc, buf)

	addr, port := "127.0.0.1", "10001"
	encodedAddr, err := ftp_ip.Encode(addr, port)
	if err != nil {
		log.Fatal(err)
	}
	cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.PORT, Arg: encodedAddr})
	expected := []byte("200 PORT command successful.\n")
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	buf.Reset()

	var wg sync.WaitGroup
	wg.Add(1)

	listenDataConn(addr+":"+port, &wg, func(conn net.Conn) {
		result, err := ioutil.ReadAll(conn)
		if err != nil {
			log.Fatal(err)
		}
		exp := []byte("Hello, World!")
		if !bytes.Equal(result, exp) {
			t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(result), "\n"),
				strings.TrimSuffix(string(expected), "\n"))
		}
	})

	expected = []byte("150 Opening ASCII mode data connection.\n226 Transfer complete.\n")
	err = cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.RETR, Arg: "test_file"})
	if ok, want, have := test_utils.VerifyError(err, nil); !ok {
		t.Errorf("Error actual = %v, and Expected = %v.", have, want)
	}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("Error actual = %s, and Expected = %s.", strings.TrimSuffix(string(buf.Bytes()), "\n"),
			strings.TrimSuffix(string(expected), "\n"))
	}
	wg.Wait()
}

func initCC() (*client_connection.ClientConnection, *bytes.Buffer, chan client_connection.AuthPkg) {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		os.Mkdir(root, os.ModePerm)
		os.Mkdir(root+"/1", os.ModePerm)
		os.Mkdir(root+"/2", os.ModePerm)
		err := ioutil.WriteFile(root+"/test_file", []byte("Hello, World!"), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	authChan := make(chan client_connection.AuthPkg)
	go func() {
		for auth := range authChan {
			auth.ReplyCh <- auth.User == "user" && auth.Password == "pass"
		}
	}()
	buf := make([]byte, 0, 1024)
	bytesBuf := bytes.NewBuffer(buf)
	cc := client_connection.New(bytesBuf, authChan, root, "127.0.0.1")
	return cc, bytesBuf, authChan
}

func authenticate(cc *client_connection.ClientConnection, buf *bytes.Buffer) {
	if err := cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.USER, Arg: "user"}); err != nil {
		log.Fatal(err)
	}
	if err := cc.Reply(&ftp_cmd.Cmd{Type: ftp_cmd.PASS, Arg: "pass"}); err != nil {
		log.Fatal(err)
	}
	buf.Reset()
}

func listenDataConn(addr string, wg *sync.WaitGroup, action func(net.Conn)) {
	go func() {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		action(conn)
		ln.Close()
		conn.Close()
		wg.Done()
	}()
}

func dialDataConn(addr string, wg *sync.WaitGroup, action func(net.Conn)) {
	go func() {
		conn, err := net.Dial("tcp", ":"+strings.Split(addr, ":")[1])
		if err != nil {
			log.Fatal(err)
		}
		action(conn)
		wg.Done()
	}()
}
