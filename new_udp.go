package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func startUDP(addr string, port int) {
	client, err := Listen("0.0.0.0:3000", nil)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	srv, err := net.Dial("udp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		panic(err)
	}
	defer srv.Close()
	go func() {
		for {
			b := make([]byte, 8192)
			i, err := client.Read(b)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(b[:i]))
			srv.Write(b[:i])
		}
	}()
	for {
		b := make([]byte, 8192)
		i, err := srv.Read(b)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(b[:i]))
		client.Write(b[:i])
	}
}

// Agent is a faux ICE agent
type Agent struct {
	udpConn net.PacketConn
	cache   chan []byte

	dstLock sync.RWMutex
	dst     *net.UDPAddr
}

// Listen creates a new listening Agent
func Listen(listenAddr string, initialDstAddr *net.UDPAddr) (*Agent, error) {
	pc, err := net.ListenPacket("udp4", listenAddr)
	if err != nil {
		panic(err)
	}

	a := &Agent{
		udpConn: pc,
		cache:   make(chan []byte, 100),
		dst:     initialDstAddr,
	}
	go func() {
		b := make([]byte, 8192)
		for {
			i, src, err := a.udpConn.ReadFrom(b)
			a.dstLock.Lock()
			a.dst = src.(*net.UDPAddr)
			a.dstLock.Unlock()
			if err != nil {
				panic(err)
			}
			a.cache <- append([]byte{}, b[:i]...)
		}
	}()

	return a, nil
}

func (a *Agent) Read(p []byte) (int, error) {
	out := <-a.cache
	if len(p) < len(out) {
		panic("Buffer too small")
	}

	copy(p, out)
	return len(out), nil
}

// Write writes len(p) bytes from p to the DTLS connection
func (a *Agent) Write(p []byte) (n int, err error) {
	a.dstLock.RLock()
	defer a.dstLock.RUnlock()

	return a.udpConn.WriteTo(p, a.dst)
}

// Close is a stub
func (a *Agent) Close() error {
	return nil
}

// LocalAddr is a stub
func (a *Agent) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr is a stub
func (a *Agent) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline is a stub
func (a *Agent) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline is a stub
func (a *Agent) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline is a stub
func (a *Agent) SetWriteDeadline(t time.Time) error {
	return nil
}
