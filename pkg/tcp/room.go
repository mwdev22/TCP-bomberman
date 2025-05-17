package tcp

import (
	"log"
	"sync"
	"time"

	"github.com/mwdev22/TCP-bomberman/pkg/board"
)

type Room struct {
	Name     string
	Board    *board.Board
	Clients  map[string]*Client
	RoomLock sync.Mutex

	shutdown chan struct{}
}

func NewRoom(name string, width, height int) *Room {
	room := &Room{
		Name:     name,
		Board:    board.New(width, height),
		Clients:  make(map[string]*Client),
		shutdown: make(chan struct{}),
	}
	go room.tickLoop()
	return room
}

func (r *Room) tickLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-r.shutdown:
			return
		case <-ticker.C:
			r.RoomLock.Lock()
			destroyedPlayers, changed := r.Board.Tick()

			for _, id := range destroyedPlayers {
				log.Printf("Destroying player %s\n", id)
				client, ok := r.Clients[id]
				if ok {
					client.Conn.Write([]byte("You have been destroyed!\n"))
					client.Disconnect()
					delete(r.Clients, id)
				}
			}

			r.RoomLock.Unlock()

			if changed {
				r.Broadcast(r.Board.String())
			}
		}
	}
}

func (r *Room) Broadcast(message string) {
	for _, client := range r.Clients {
		_, _ = client.Conn.Write([]byte(message))
	}
}

func (r *Room) Shutdown() {
	close(r.shutdown)
}
