package main

import (
	//"github.com/miketruman/com"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"deltadiode/HashReceiver"
	"deltadiode/Message"
	"deltadiode/UDPSender"
	"deltadiode/UDPreceiver"
	"fmt"
	"github.com/gorilla/mux"
	"hash"
	"io"
	//	"io/ioutil"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func getSHA256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func RelayNode(numChannels int, fromInterface, toInterface string, groupFrom int, groupTo int) {
	for i := 0; i < numChannels; i++ {
		go func(i int) {
			receiver := HashReceiver.NewHASHreceiver("224.0.0.1:"+strconv.Itoa(groupFrom+i), fromInterface)
			err := receiver.Start()
			if err != nil {
				fmt.Println("Failed to start UDPreceiver:", err)
				return
			}

			sender := UDPSender.NewUDPSender("224.0.0.1:"+strconv.Itoa(groupTo+i), toInterface)
			err = sender.Start()
			if err != nil {
				fmt.Println("Failed to start UDPSender:", err)
				return
			}

			//			uniqueSet := make(map[string]bool)
			for msg := range receiver.MessageCh {
				/*
					if len(uniqueSet) > 10000 {
						uniqueSet = make(map[string]bool)
						fmt.Println("CLEAN")
					}

					if _, exists := uniqueSet[string(msg)]; exists {
						fmt.Println("PAAAAAAAAAANI")
					}
					uniqueSet[string(msg)] = true*/
				sender.SendHash(getSHA256(msg))
			}
		}(i)
	}
}

func NewMultiChannelQueue(numChannels int, fromInterface, toInterface string, fromPort, toPort int) *MultiChannelQueue {
	mcq := &MultiChannelQueue{
		locks:       make([]sync.Mutex, numChannels),
		udpsender:   make([]*UDPSender.UDPSender, numChannels),
		udpreceiver: make([]*HashReceiver.HASHreceiver, numChannels),
	}

	for i := range mcq.udpsender {
		mcq.udpsender[i] = UDPSender.NewUDPSender("224.0.0.1:"+strconv.Itoa(toPort+i), toInterface)
		err := mcq.udpsender[i].Start()
		if err != nil {
			fmt.Println("Failed to start UDPSender:", err)
		}
		mcq.udpreceiver[i] = HashReceiver.NewHASHreceiver("224.0.0.1:"+strconv.Itoa(fromPort+i), fromInterface)
		err = mcq.udpreceiver[i].Start()
		if err != nil {
			fmt.Println("Failed to start UDPreceiver:", err)
		}
	}

	return mcq
}

type MultiChannelQueue struct {
	//      channels []chan []byte
	locks       []sync.Mutex
	udpsender   []*UDPSender.UDPSender
	udpreceiver []*HashReceiver.HASHreceiver
}

type PitcherStore struct {
	multiChannelQueue *MultiChannelQueue
}

func (pitcherStore PitcherStore) uploadTestFile(w http.ResponseWriter, r *http.Request) {

	filename := randomString(10) + ".dat"
	//	username := "test"
	const size10MB = 100 * 1024 * 1024 // 100MB
	reader := &DataReader{size: size10MB}
	startTime := time.Now()
	err := pitcherStore.sharedUpload(filename, reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}

	// File transfer end
	endTime := time.Now()

	// Calculating the duration in seconds
	duration := endTime.Sub(startTime).Seconds()
	speedBps := float64(size10MB) * 8 / duration
	fmt.Fprintf(w, "Filename: %s \n", filename)
	fmt.Fprintf(w, "File Size: %d MB\n", size10MB/(1024*1024))
	fmt.Fprintf(w, "Time Duration: %.2f seconds\n", duration)
	fmt.Fprintf(w, "Network Speed: %.2f Mbps\n", speedBps/(1024*1024))
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		panic(err)
	}

	for i, byteVal := range b {
		b[i] = charset[byteVal%byte(len(charset))]
	}
	return string(b)
}

// Define a type that will implement io.Reader interface
type DataReader struct {
	size int
	read int
}

// Implement the Read method for DataReader
func (dr *DataReader) Read(p []byte) (int, error) {
	// Check if the read size has reached the total size
	if dr.read >= dr.size {
		return 0, io.EOF // End of file (or data stream in this case)
	}

	// Calculate the remaining data size
	remaining := dr.size - dr.read
	if remaining < len(p) {
		p = p[:remaining]
	}

	// Fill the byte slice with some data, e.g., 'A'
	for i := range p {
		p[i] = 'A'
	}

	// Update the read count
	dr.read += len(p)

	// Return the number of bytes read
	return len(p), nil
}

type FileWriterStore struct {
	FileIndex uint64
	Writer    *os.File
	Hash      hash.Hash
}

func (fileWriterStore *FileWriterStore) Process(msg *Message.Message) {
	if msg.FileIndex == 0 { // new File
		fileWriterStore.Writer.Close()
		fileWriterStore.FileIndex = msg.FileIndex
		fileWriterStore.Hash = sha256.New()
		tmpFile := "/data/" + msg.Filename

		var err error
		fileWriterStore.Writer, err = os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE, 0777)
		if err != nil {
			panic(err)
		}

	} else if fileWriterStore.FileIndex+1 == msg.FileIndex {
		// This is expected
	} else {
		fmt.Println("Panic! msg.FileIndex:", msg.FileIndex, " fileWriterStore.FileIndex:", fileWriterStore.FileIndex)
		return
	}
	if fileWriterStore.Writer != nil {
		fileWriterStore.Hash.Write(msg.Data)
		_, err := fileWriterStore.Writer.Write(msg.Data)
		if err != nil {
			panic(err)
		}

	}
	if msg.Last {
		fmt.Println("File completed")
		fileWriterStore.Writer.Close()
		readHash := fileWriterStore.Hash.Sum(nil)
		res := bytes.Compare(msg.Hash, readHash)
		if res == 0 {
			fmt.Println("Correct catched file:", msg.Filename)
		}

	}

	fileWriterStore.FileIndex = msg.FileIndex
}

func fileReceiver(numChannels int, fromInterface, toInterface string, fromPort, toPort int) {

	for i := 0; i < numChannels; i++ {
		go func(i int) {
			sender := UDPSender.NewUDPSender("224.0.0.1:"+strconv.Itoa(toPort+i), toInterface)
			err := sender.Start()
			if err != nil {
				fmt.Println("Failed to start UDPSender:", err)
				return
			}
			receiver := UDPreceiver.NewUDPreceiver("224.0.0.1:"+strconv.Itoa(fromPort+i), fromInterface)
			err = receiver.Start()
			if err != nil {
				fmt.Println("Failed to start UDPreceiver:", err)
				return
			}
			fileWriterStore := FileWriterStore{}

			for msg := range receiver.MessageCh {
				//		fmt.Println("Received message:", msg.FileIndex)

				fileWriterStore.Process(&msg)
				if msg.FileIndex == 0 { // new File

				}
				sender.SendHash(getSHA256(msg.Hash)) // Send ACK
			}
		}(i)
	}
}

func (pitcherStore PitcherStore) uploadFile(w http.ResponseWriter, r *http.Request) {
	multipartReader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "expecting a multipart message", http.StatusBadRequest)
		return
	}

	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		defer part.Close()
		fmt.Println("FormName:", part.FormName())
		if part.FormName() == "user" {
			/*
				b, err := ioutil.ReadAll(part)
				if err != nil {
					http.Error(w, "expecting user", http.StatusBadRequest)
					return
				}*/
		}

		if part.FormName() == "myFile" {
			/*	if len(user) == 0 {
				http.Error(w, "missing user", http.StatusBadRequest)
				return
			}*/
			fmt.Println("Sending:", part.FileName())
			pitcherStore.sharedUpload(part.FileName(), part)
			fmt.Println("Completed:", part.FileName())
			fmt.Fprintln(w, "Completed filename:", part.FileName())
		}
	}

}

func (pitcherStore PitcherStore) sharedUpload(filename string, file io.Reader) error {

	for i, _ := range pitcherStore.multiChannelQueue.locks {
		if pitcherStore.multiChannelQueue.locks[i].TryLock() {
			fmt.Println("Sending on channel:", i)
			hash := sha256.New()
			buf := make([]byte, 8000)

			fileIndex := uint64(0)
			for {
				now := time.Now()
				sequenceNumber := uint64(now.UnixNano())
				tr := io.TeeReader(file, hash)
				n, err := tr.Read(buf)
				lastMessage := false
				if n == 0 && err == io.EOF {
					lastMessage = true
				}
				fileHash := hash.Sum(nil)
				message := Message.Message{Data: buf[:n], Filename: filename, FileIndex: fileIndex, Hash: fileHash, SequenceNumber: sequenceNumber, Last: lastMessage}
				pitcherStore.multiChannelQueue.udpsender[i].Send(message)
				//			receiverHash := sha256.New()
				//			receiverHash.Write(fileHash)
				receiverHash := getSHA256(getSHA256(fileHash))

				reTransmitCount := 0
			NextMessage: // Label for break
				for {
					select {
					case payload, ok := <-pitcherStore.multiChannelQueue.udpreceiver[i].MessageCh:
						{
							if !ok {
								log.Fatalf("Panic! ok was not true")
							} else {
								if len(payload) == len(receiverHash) && bytes.Compare(payload, receiverHash) == 0 {
									break NextMessage // Message was delivered
								} else {
									fmt.Println("Panic received message did not match")
								}
							}

						}

					case <-time.After(50 * time.Millisecond):
						{
							fmt.Println("Message not received.....")
						}
					}
					fmt.Println("Re-transmit")
					reTransmitCount += 1
					if reTransmitCount > 10 {
						fmt.Println("Retransmissions exceeded limit")
						return errors.New("Err: Retransmissions exceeded limit")
					}
					pitcherStore.multiChannelQueue.udpsender[i].Send(message)
				}
				fileIndex += 1
				if n == 0 && err == io.EOF {
					break
				}

			}

			pitcherStore.multiChannelQueue.locks[i].Unlock()
			fmt.Println("unlocked")
			return nil
		}
	}
	fmt.Println("Err: Panic all channels in use")
	return errors.New("Err: All output channels in use")
}

func main() {

	threads := 32
	fileReceiver(threads, "enp0s31f6", "enp0s31f6", 3000, 4000)

	RelayNode(threads, "enxac1a3d8c49f9", "enxac1a3d8c49f9", 4000, 8000)

	pitcherStore := PitcherStore{multiChannelQueue: NewMultiChannelQueue(threads, "enx24f5a2f21f4b", "enx24f5a2f21f4b", 8000, 3000)}

	mux := mux.NewRouter()
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)
	mux.HandleFunc("/upload", pitcherStore.uploadFile)
	mux.HandleFunc("/test", pitcherStore.uploadTestFile)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
