package client

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string
	ID    string
	Data  string
}

// SSEParser parses Server-Sent Events from a response stream
type SSEParser struct {
	scanner *bufio.Scanner
}

// NewSSEParser creates a new SSE parser
func NewSSEParser(r io.Reader) *SSEParser {
	return &SSEParser{
		scanner: bufio.NewScanner(r),
	}
}

// ParseNext parses the next SSE event from the stream
func (p *SSEParser) ParseNext() (*SSEEvent, error) {
	event := &SSEEvent{}

	for p.scanner.Scan() {
		line := p.scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			// Return the event if we have data
			if event.Data != "" {
				return event, nil
			}
			// Otherwise continue to next event
			continue
		}

		// Parse SSE field
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			field := strings.TrimSpace(parts[0])
			value := ""
			if len(parts) > 1 {
				value = strings.TrimSpace(parts[1])
			}

			switch field {
			case "event":
				event.Event = value
			case "id":
				event.ID = value
			case "data":
				if event.Data != "" {
					event.Data += "\n" + value
				} else {
					event.Data = value
				}
			}
		}
	}

	// Check for scanner errors
	if err := p.scanner.Err(); err != nil {
		return nil, fmt.Errorf("SSE scanner error: %w", err)
	}

	// Return final event if we have data
	if event.Data != "" {
		return event, nil
	}

	// No more events
	return nil, io.EOF
}

// ParseSingle parses a single SSE response and returns the data
func ParseSSEResponse(r io.Reader) (string, error) {
	parser := NewSSEParser(r)
	event, err := parser.ParseNext()
	if err != nil {
		return "", fmt.Errorf("failed to parse SSE event: %w", err)
	}

	if event.Event != "message" {
		return "", fmt.Errorf("unexpected SSE event type: %s", event.Event)
	}

	return event.Data, nil
}
