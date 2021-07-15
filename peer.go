package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Peer struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	ClientIP  string    `json:"clientIP"`
	UserAgent string    `json:"userAgent"`

	writeC chan interface{} `json:"-"` // required to prevent concurrent writes
}

func NewPeer(userAgent string, clientIP string) *Peer {
	return &Peer{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		ClientIP:  clientIP,
		UserAgent: userAgent,

		writeC: make(chan interface{}),
	}
}

func (p *Peer) Read(conn *websocket.Conn) (*InboundMessage, error) {
	var message InboundMessage
	if err := conn.ReadJSON(&message); err != nil {
		return nil, err
	}
	return &message, nil
}

func (p *Peer) Write(message *OutboundMessage) {
	p.writeC <- message
}

func (p *Peer) WriteLoop(conn *websocket.Conn) error {
	for v := range p.writeC {
		if err := conn.WriteJSON(v); err != nil {
			return err
		}
	}
	return nil
}

func (p *Peer) Close() {
	close(p.writeC)
}
