package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

const (
	MsgChoke         messageID = 0
	MsgUnchoke       messageID = 1
	MsgInterested    messageID = 2
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
)

type Message struct {
	ID      messageID
	Payload []byte
}

func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}

func (m *Message) string() string {
	if m == nil {
		return m.name()
	}
	return fmt.Sprintf("%s (%d bytes)", m.name(), len(m.Payload))
}

func (m *Message) Serialize() []byte {
	if m == nil {
		return make([]byte, 4) // Keep Alive message
	}
	length := uint32(len(m.Payload) + 1) // +1 for the message ID
	buf := make([]byte, length+4)
	binary.BigEndian.PutUint32(buf[0:4], uint32(length))
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

func Read(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	if length == 0 {
		return nil, nil // Keep Alive message
	}
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}
	m := Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}
	return &m, nil
}

func ParseHave(m *Message) (int, error) {
	if m.ID != MsgHave {
		return 0, fmt.Errorf("Expected Have message, got %s", m.string())
	}
	if len(m.Payload) != 4 {
		return 0, fmt.Errorf("Malformed Have message: expected 4 bytes payload, got %d", len(m.Payload))
	}
	index := int(binary.BigEndian.Uint32(m.Payload))
	return index, nil
}

func ParsePiece(index int, buf []byte, m *Message) (int, error) {
	if m.ID != MsgPiece {
		return 0, fmt.Errorf("Expected Piece message, got %s", m.string())
	}
	if len(m.Payload) < 8 {
		return 0, fmt.Errorf("Malformed Piece message: expected at least 8 bytes payload, got %d", len(m.Payload))
	}
	parsedIndex := int(binary.BigEndian.Uint32(m.Payload[0:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("Piece index mismatch: expected %d, got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(m.Payload[4:8])) // Fix: Read from correct offset
	if begin >= len(buf) {
		return 0, fmt.Errorf("Begin index %d is out of bounds for buffer length %d", begin, len(buf))
	}
	data := m.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("Data length %d with begin index %d exceeds buffer length %d", len(data), begin, len(buf))
	}
	copy(buf[begin:], data)
	return len(data), nil
}

func (m *Message) name() string {
	if m == nil {
		return "Keep Alive"
	}
	switch m.ID {
	case MsgChoke:
		return "Choke"
	case MsgUnchoke:
		return "Unchoke"
	case MsgInterested:
		return "Interested"
	case MsgNotInterested:
		return "Not Interested"
	case MsgHave:
		return "Have"
	case MsgBitfield:
		return "Bitfield"
	case MsgRequest:
		return "Request"
	case MsgPiece:
		return "Piece"
	case MsgCancel:
		return "Cancel"
	default:
		return fmt.Sprintf("Unknown Message ID %d", m.ID)
	}
}
