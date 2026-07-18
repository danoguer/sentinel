package transport

import (
	"encoding/json"
	"fmt"
	"net"
)

func SendToAgent(socketPath string, envelope ContextEnvelope) (*AgentResponse, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Sentinel Agent: %w", err)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(envelope); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	var resp AgentResponse

	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	return &resp, nil
}
