package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

type StdioTransport struct {
	encoder    *json.Encoder
	decoder    *json.Decoder
	requests   chan *protocol.Request
	errors     chan error
	done       chan struct{}
	mu         sync.RWMutex
	isClosed   bool
}

func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		encoder:   json.NewEncoder(os.Stdout),
		decoder:   json.NewDecoder(os.Stdin),
		requests:  make(chan *protocol.Request),
		errors:    make(chan error),
		done:      make(chan struct{}),
	}
}

func (t *StdioTransport) Start(ctx context.Context) error {
	go t.readLoop(ctx)
	return nil
}

func (t *StdioTransport) Stop(_ context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if !t.isClosed {
		log.Printf("[STDIO] Stopping transport - closing channels")
		t.isClosed = true
		close(t.done)
		close(t.requests)
		close(t.errors)
	} else {
		log.Printf("[STDIO] Stop called but transport already closed")
	}
	return nil
}

func (t *StdioTransport) Send(response *protocol.Response) error {
	t.mu.RLock()
	if t.isClosed {
		t.mu.RUnlock()
		return fmt.Errorf("transport is closed")
	}
	t.mu.RUnlock()

	return t.encoder.Encode(response)
}

func (t *StdioTransport) Receive() <-chan *protocol.Request {
	return t.requests
}

func (t *StdioTransport) Errors() <-chan error {
	return t.errors
}

func (t *StdioTransport) readLoop(ctx context.Context) {
	defer t.Stop(ctx)
	log.Printf("[STDIO] Starting readLoop")

	for {
		select {
		case <-ctx.Done():
			log.Printf("[STDIO] readLoop: context cancelled")
			return
		case <-t.done:
			log.Printf("[STDIO] readLoop: done channel closed")
			return
		default:
			var request protocol.Request
			err := t.decoder.Decode(&request)

			if err == io.EOF {
				log.Printf("[STDIO] readLoop: EOF received on stdin")
				return
			}

			t.mu.RLock()
			isClosed := t.isClosed
			t.mu.RUnlock()

			if isClosed {
				return
			}

			if err != nil {
				log.Printf("[STDIO] readLoop: decode error: %v", err)
				select {
				case t.errors <- fmt.Errorf("decode error: %w", err):
				case <-ctx.Done():
					return
				case <-t.done:
					return
				default:
					log.Printf("[STDIO] Error decoding request: %v", err)
				}
				continue
			}

			if request.JSONRPC != "2.0" {
				select {
				case t.errors <- fmt.Errorf("invalid JSON-RPC version: %s", request.JSONRPC):
				case <-ctx.Done():
					return
				case <-t.done:
					return
				default:
					log.Printf("Invalid JSON-RPC version: %s", request.JSONRPC)
				}
				continue
			}

			select {
			case t.requests <- &request:
			case <-ctx.Done():
				return
			case <-t.done:
				return
			}
		}
	}
}
