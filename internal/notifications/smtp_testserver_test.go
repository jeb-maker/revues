package notifications_test

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"testing"
)

func startTestSMTPServer(t *testing.T) (host string, port int) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen(): %v", err)
	}
	t.Cleanup(func() {
		_ = ln.Close()
	})

	go func() {
		for {
			conn, acceptErr := ln.Accept()
			if acceptErr != nil {
				return
			}
			go handleTestSMTP(conn)
		}
	}()

	host, portStr, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort(): %v", err)
	}
	port, err = strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("Atoi(): %v", err)
	}

	return host, port
}

func handleTestSMTP(conn net.Conn) {
	defer conn.Close()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	writeSMTPLine(rw, "220 localhost ESMTP")

	for {
		line, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.ToUpper(strings.Fields(strings.TrimSpace(line))[0])
		switch cmd {
		case "EHLO", "HELO":
			writeSMTPLine(rw, "250-localhost")
			writeSMTPLine(rw, "250 OK")
		case "MAIL", "RCPT":
			writeSMTPLine(rw, "250 OK")
		case "DATA":
			writeSMTPLine(rw, "354 End data with <CR><LF>.<CR><LF>")
			for {
				part, err := rw.ReadString('\n')
				if err != nil {
					return
				}
				if strings.TrimSpace(part) == "." {
					break
				}
			}
			writeSMTPLine(rw, "250 OK")
			notifySMTPCapture()
		case "QUIT":
			writeSMTPLine(rw, "221 Bye")
			return
		default:
			writeSMTPLine(rw, "250 OK")
		}
	}
}

func writeSMTPLine(rw *bufio.ReadWriter, line string) {
	_, _ = rw.WriteString(line + "\r\n")
	_ = rw.Flush()
}
