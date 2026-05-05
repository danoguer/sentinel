package listener

import (
	"bufio"
	"log"
	"net"
	"os"
	"sync"
)

type TerminalData struct {
	sync.Mutex
	Memory []string
}

var GlobalData = &TerminalData{
	Memory: make([]string, 0, 50),
}

func (w *TerminalData) Addline(line string) {
	w.Lock()
	defer w.Unlock()
	w.Memory = append(w.Memory, line)

	if len(w.Memory) > 50 {
		w.Memory = w.Memory[1:]
	}
}

func handleConnection(connection net.Conn) {

	defer connection.Close()

	scanner := bufio.NewScanner(connection)

	for scanner.Scan() {
		line := scanner.Text()
		GlobalData.Addline(line)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}
}

func StartSocketListener() {
	socketPath := "/tmp/sentinel.sock"

	os.Remove(socketPath)

	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to start Sentinel listener: %v", err)
	}

	defer socket.Close()
	defer os.Remove(socketPath)
	for {
		connection, err := socket.Accept()
		if err != nil {
			continue
		}

		go handleConnection(connection)
	}
}
