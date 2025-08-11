package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Fprintf(os.Stderr, "Использование: %s host port [--timeout=10s]\n", os.Args[0])
		os.Exit(1)
	}

	host := flag.Arg(0)
	port := flag.Arg(1)
	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка подключения: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Printf("Connected to %s\n", address)

	done := make(chan struct{})

	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil && !isClosedNetworkError(err) {
			fmt.Fprintf(os.Stderr, "Ошибка чтения: %v\n", err)
		}
		close(done)
	}()

	go func() {
		_, err := io.Copy(conn, os.Stdin)
		if err != nil && !isClosedNetworkError(err) {
			fmt.Fprintf(os.Stderr, "Ошибка записи: %v\n", err)
		}
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	<-done
	fmt.Println("\nПодключение закрыто.")
}

func isClosedNetworkError(err error) bool {
	if err == io.EOF {
		return true
	}
	if netErr, ok := err.(net.Error); ok && !netErr.Timeout() {
		return true
	}
	return false
}
