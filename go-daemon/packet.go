package main

import (
	"strings"
)

// ParsePacket converts bytes to Packet struct
func ParsePacket(data []byte) *Packet {
	p := &Packet{
		Version:     data[0],
		MessageType: data[1],
		TopicLen:    data[2],
	}
	copy(p.Payload[:], data[3:])
	return p
}

// CreatePacket creates a packet from topic and message
func CreatePacket(topic, message string) []byte {
	packet := make([]byte, PacketSize)
	packet[0] = ProtocolVersion
	packet[1] = MessageTypePublish

	payload := topic + "|" + message
	payloadBytes := []byte(payload)

	// Truncate if needed
	if len(payloadBytes) > 61 {
		payloadBytes = payloadBytes[:61]
	}

	packet[2] = byte(len(topic))
	copy(packet[3:], payloadBytes)

	return packet
}

// CreatePacketFromPayload creates packet from "topic|message" string
func CreatePacketFromPayload(payload string) []byte {
	parts := strings.SplitN(payload, "|", 2)
	if len(parts) != 2 {
		return CreatePacket(payload, "")
	}
	return CreatePacket(parts[0], parts[1])
}

// ExtractTopicAndMessage extracts topic and message from packet
func ExtractTopicAndMessage(p *Packet) (string, string) {
	payload := string(p.Payload[:])
	payload = strings.TrimRight(payload, "\x00") // Remove null bytes

	parts := strings.SplitN(payload, "|", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return payload, ""
}
