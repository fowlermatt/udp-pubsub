package main

import (
	"strings"
)

const (
	ProtocolVersion    byte = 0x01
	MessageTypePublish byte = 0x01
	MessageTypeAlive   byte = 0x02

	PacketSize = 64
)

type Packet struct {
	Version     byte
	MessageType byte
	TopicLen    byte
	Payload     [61]byte
}

func ParsePacket(data []byte) *Packet {
	p := &Packet{
		Version:     data[0],
		MessageType: data[1],
		TopicLen:    data[2],
	}
	copy(p.Payload[:], data[3:])
	return p
}

func CreatePacket(topic, message string) []byte {
	packet := make([]byte, PacketSize)
	packet[0] = ProtocolVersion
	packet[1] = MessageTypePublish

	payload := topic + "|" + message
	payloadBytes := []byte(payload)

	if len(payloadBytes) > 61 {
		payloadBytes = payloadBytes[:61]
	}

	packet[2] = byte(len(topic))
	copy(packet[3:], payloadBytes)

	return packet
}

func CreatePacketFromPayload(payload string) []byte {
	parts := strings.SplitN(payload, "|", 2)
	if len(parts) != 2 {
		return CreatePacket(payload, "")
	}
	return CreatePacket(parts[0], parts[1])
}

func ExtractTopicAndMessage(p *Packet) (string, string) {
	payload := string(p.Payload[:])
	payload = strings.TrimRight(payload, "\x00")

	parts := strings.SplitN(payload, "|", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return payload, ""
}
