package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	UDPPort        = 9876
	UnixSocketPath = "/tmp/pubsub.sock"
	BroadcastAddr  = "255.255.255.255:9876"
)

type Daemon struct {
	clients   map[net.Conn]*Client
	clientsMu sync.RWMutex
	udpConn   *net.UDPConn
	listener  net.Listener
	wg        sync.WaitGroup
	shutdown  chan bool
}

func NewDaemon() *Daemon {
	return &Daemon{
		clients:  make(map[net.Conn]*Client),
		shutdown: make(chan bool),
	}
}

func (d *Daemon) Start() error {
	os.Remove(UnixSocketPath)

	listener, err := net.Listen("unix", UnixSocketPath)
	if err != nil {
		return fmt.Errorf("failed to create Unix socket: %v", err)
	}
	d.listener = listener
	log.Printf("Unix socket listening on %s", UnixSocketPath)

	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", UDPPort))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	d.udpConn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %v", err)
	}
	log.Printf("UDP listening on port %d", UDPPort)

	d.wg.Add(2)
	go d.acceptClients()
	go d.handleUDPMessages()

	return nil
}

func (d *Daemon) acceptClients() {
	defer d.wg.Done()

	for {
		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.shutdown:
				return
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		client := NewClient(conn)
		d.clientsMu.Lock()
		d.clients[conn] = client
		d.clientsMu.Unlock()

		go d.handleClient(client)
		log.Printf("New client connected: %v", conn.RemoteAddr())
	}
}

func (d *Daemon) handleClient(client *Client) {
	defer func() {
		client.conn.Close()
		d.clientsMu.Lock()
		delete(d.clients, client.conn)
		d.clientsMu.Unlock()
		log.Printf("Client disconnected: %v", client.conn.RemoteAddr())
	}()

	buffer := make([]byte, 1024)
	for {
		n, err := client.conn.Read(buffer)
		if err != nil {
			return
		}

		message := string(buffer[:n])
		d.processClientMessage(client, message)
	}
}

func (d *Daemon) processClientMessage(client *Client, message string) {
	if len(message) < 4 {
		return
	}

	cmd := message[:3]
	payload := message[4:]

	switch cmd {
	case "PUB":
		d.handlePublish(payload)
	case "SUB":
		d.handleSubscribe(client, []byte(payload))
	default:
		log.Printf("Unknown command: %s", cmd)
	}
}

func (d *Daemon) handlePublish(payload string) {
	packet := CreatePacketFromPayload(payload)
	d.broadcastPacket(packet)
}

func (d *Daemon) handleSubscribe(client *Client, filterData []byte) {
	err := client.UpdateFilter(filterData)
	if err != nil {
		log.Printf("Failed to update filter: %v", err)
		return
	}
	log.Printf("Client filter updated")
}

func (d *Daemon) broadcastPacket(packet []byte) {
	addr, _ := net.ResolveUDPAddr("udp", BroadcastAddr)
	_, err := d.udpConn.WriteToUDP(packet, addr)
	if err != nil {
		log.Printf("Broadcast error: %v", err)
	}
}

func (d *Daemon) handleUDPMessages() {
	defer d.wg.Done()

	buffer := make([]byte, PacketSize)
	for {
		n, _, err := d.udpConn.ReadFromUDP(buffer)
		if err != nil {
			select {
			case <-d.shutdown:
				return
			default:
				log.Printf("UDP read error: %v", err)
				continue
			}
		}

		if n != PacketSize {
			continue
		}

		packet := ParsePacket(buffer)
		if packet.Version != ProtocolVersion {
			continue
		}

		d.routePacket(packet)
	}
}

func (d *Daemon) routePacket(packet *Packet) {
	topic, message := ExtractTopicAndMessage(packet)

	d.clientsMu.RLock()
	defer d.clientsMu.RUnlock()

	for conn, client := range d.clients {
		if client.TestTopic(topic) {
			msg := fmt.Sprintf("MSG %s|%s", topic, message)
			conn.Write([]byte(msg))
		}
	}
}

func (d *Daemon) Shutdown() {
	close(d.shutdown)
	d.listener.Close()
	d.udpConn.Close()
	d.wg.Wait()
	os.Remove(UnixSocketPath)
	log.Println("Daemon shutdown complete")
}

func main() {
	log.Println("UDP PubSub Daemon Starting...")

	daemon := NewDaemon()
	if err := daemon.Start(); err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down daemon...")
	daemon.Shutdown()
}
