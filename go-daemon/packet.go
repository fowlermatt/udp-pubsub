package main

// Protocol constants
const (
	ProtocolVersion    byte = 0x01
	MessageTypePublish byte = 0x01
	MessageTypeAlive   byte = 0x02

	PacketSize = 64
)

// Packet represents our 64-byte UDP packet structure
type Packet struct {
	Version     byte     // Byte 0: Protocol version
	MessageType byte     // Byte 1: Message type
	TopicLen    byte     // Byte 2: Length of topic
	Payload     [61]byte // Bytes 3-63: Topic + "|" + Message
}
