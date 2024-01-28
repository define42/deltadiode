package Message

import (
//	"time"
)

type MessageType int

// Declare constants using the iota enumerator
const (
	Data MessageType = iota // 0
	ACK                     // 1
	LAST                     
)

type Message struct {
	SequenceNumber uint64
	FileIndex      uint64 // Unique file ID
	Data           []byte // data block
	BlockIndex     uint32 //
	FileSize       uint64
	Last	bool
//	Timestamp      time.Time
	Filename       string
	Hash           []byte 
}
