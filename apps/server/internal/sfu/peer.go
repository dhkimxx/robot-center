package sfu

import (
	"encoding/json"
	"strings"
)

func (p *peer) enqueue(message signalMessage) {
	select {
	case p.send <- message:
	default:
	}
}

func (p *peer) readLoop(h *Hub) {
	defer func() {
		h.unregisterPeer(p)
		_ = p.conn.Close()
	}()

	for {
		_, rawMessage, err := p.conn.ReadMessage()
		if err != nil {
			return
		}

		var message signalMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			continue
		}
		message.Type = strings.TrimSpace(message.Type)
		if message.Type == "" {
			continue
		}

		h.handleSignal(p, message)
	}
}

func (p *peer) writeLoop() {
	defer func() {
		_ = p.conn.Close()
	}()

	for message := range p.send {
		if err := p.conn.WriteJSON(message); err != nil {
			return
		}
	}
}
