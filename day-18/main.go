package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func downloadURL(ctx context.Context, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}
	n, err := io.Copy(io.Discard, resp.Body)
	return n, err
}

func downloadAll(ctx context.Context, urls []string) {
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			start := time.Now()
			n, err := downloadURL(ctx, u)
			if err != nil {
				if ctx.Err() != nil {
					fmt.Printf("  CANCELLED  %s\n", u)
				} else {
					fmt.Printf("  ERROR      %s: %v\n", u, err)
				}
				return
			}
			fmt.Printf("  OK  %6d bytes  %v  %s\n", n, time.Since(start).Round(time.Millisecond), u)
		}(url)
	}
	wg.Wait()
}

func startEchoServer(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	fmt.Printf("Echo server listening on %s\n", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		go func(c net.Conn) {
			defer c.Close()
			scanner := bufio.NewScanner(c)
			for scanner.Scan() {
				fmt.Fprintln(c, strings.ToUpper(scanner.Text()))
			}
		}(conn)
	}
}

func main() {
	fmt.Println("=== Cancellable HTTP Downloader (5s timeout) ===")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	urls := []string{
		"https://httpbin.org/get",
		"https://httpbin.org/delay/2",
		"https://httpbin.org/bytes/1024",
	}
	downloadAll(ctx, urls)

	fmt.Println("\n=== TCP Echo Server ===")
	srvCtx, srvCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer srvCancel()

	go startEchoServer(srvCtx, "127.0.0.1:9999")
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", "127.0.0.1:9999")
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial error: %v\n", err)
		return
	}
	defer conn.Close()

	for _, msg := range []string{"hello", "world", "go is great"} {
		fmt.Fprintln(conn, msg)
		buf := make([]byte, 256)
		n, _ := conn.Read(buf)
		fmt.Printf("sent: %q  got: %q\n", msg, strings.TrimSpace(string(buf[:n])))
	}
}
