package Message

type Message struct {
	SequenceNumber uint64
	FileIndex      uint64 // Unique file ID
	Data           []byte // data block
	BlockIndex     uint32 //
	FileSize       uint64
	Last	bool
	Filename       string
	Hash           []byte 
}
