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

	conn   *websocket.Conn
	writeC chan interface{}
}

func NewPeer(conn *websocket.Conn, userAgent string, clientIP string) *Peer {
	return &Peer{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		ClientIP:  clientIP,
		UserAgent: userAgent,

		conn:   conn,
		writeC: make(chan interface{}),
	}
}

func (p *Peer) Write(v interface{}) {
	p.writeC <- v
}

func (p *Peer) Read(v interface{}) error {
	return p.conn.ReadJSON(v)
}

func (p *Peer) Loop() error {
	for v := range p.writeC {
		if err := p.conn.WriteJSON(v); err != nil {
			return err
		}
	}
	return nil
}

func (p *Peer) Close() {
	close(p.writeC)
}
