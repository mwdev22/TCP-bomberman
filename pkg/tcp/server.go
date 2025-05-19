package tcp

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	Addr     string
	Port     string
	Listener net.Listener
	Clients  map[string]*Client
	Rooms    map[string]*Room
	mu       sync.Mutex

	shutdownChan chan struct{}
	running      bool
}

func NewServer(addr string, port string) *Server {
	return &Server{
		Addr:         addr,
		Port:         port,
		Clients:      make(map[string]*Client),
		Rooms:        make(map[string]*Room), // <-- WAŻNE
		shutdownChan: make(chan struct{}),
		running:      false,
	}
}

func (s *Server) buildInstructions() string {
	var builder strings.Builder
	builder.WriteString("=== Welcome to Bomberman ===\n")
	builder.WriteString("Commands:\n")
	builder.WriteString("  JOIN            - Join any available room\n")
	builder.WriteString("  JOIN <room>     - Join or create a specific room\n")
	builder.WriteString("  ROOMS           - List rooms and player counts\n")
	builder.WriteString("  Use arrow keys  - Move around (← ↑ → ↓)\n")
	builder.WriteString("  Press 'b'       - Plant a bomb\n")
	builder.WriteString("----------------------------------------\n")
	builder.WriteString("Current Rooms:\n")

	s.mu.Lock()
	defer s.mu.Unlock()
	for name, room := range s.Rooms {
		room.RoomLock.Lock()
		count := len(room.Clients)
		room.RoomLock.Unlock()
		builder.WriteString(fmt.Sprintf("  %s: %d players\n", name, count))
	}

	return builder.String()
}

func (s *Server) Listen() error {
	address := net.JoinHostPort(s.Addr, s.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.Listener = listener
	s.running = true
	log.Printf("Bomberman server listening on %s\n", address)

	s.Rooms["room-1"] = NewRoom("room-1", 13, 11)

	go func() {
		for {
			select {
			case <-s.shutdownChan:
				log.Println("server is shutting down - stopping accept loop.")
				return
			default:
				conn, err := s.Listener.Accept()
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						continue
					}
					log.Println("accept error:", err)
					continue
				}
				log.Println("new connection from", conn.RemoteAddr()) // <- teraz jest bezpiecznie

				go s.handleConnection(conn)
			}
		}
	}()
	return nil
}

func (s *Server) Stop() error {
	if !s.running {
		return nil
	}

	close(s.shutdownChan)
	s.running = false

	if s.Listener != nil {
		s.Listener.Close()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for addr, client := range s.Clients {
		log.Printf("closing connection to %s\n", addr)
		client.Disconnect()
	}
	s.Clients = make(map[string]*Client)

	log.Println("server shutdown complete.")
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	client := &Client{
		ID:   clientAddr,
		Conn: conn,
		Addr: clientAddr,
		Quit: make(chan struct{}),
	}
	s.addClient(clientAddr, client)
	defer s.removeClient(clientAddr)

	conn.Write([]byte(s.buildInstructions()))

	buf := make([]byte, 1024)
	var room *Room

	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		msg := strings.TrimSpace(string(buf[:n]))

		if msg == "ROOMS" {
			conn.Write([]byte(s.buildInstructions()))
			continue
		}

		if room == nil {
			if strings.HasPrefix(msg, "JOIN") {
				roomName := ""
				parts := strings.SplitN(msg, " ", 2)
				if len(parts) == 2 {
					roomName = strings.TrimSpace(parts[1])
				} else {
					// AUTO: find the room with the least players
					s.mu.Lock()
					minPlayers := int(^uint(0) >> 1)
					for name, r := range s.Rooms {
						r.RoomLock.Lock()
						count := len(r.Clients)
						r.RoomLock.Unlock()
						if count < minPlayers {
							minPlayers = count
							roomName = name
						}
					}
					if roomName == "" {
						roomName = fmt.Sprintf("room-%d", len(s.Rooms)+1)
					}
					s.mu.Unlock()
				}

				s.mu.Lock()
				r, ok := s.Rooms[roomName]
				if !ok {
					r = NewRoom(roomName, 13, 11)
					s.Rooms[roomName] = r
				}
				s.mu.Unlock()

				room = r
				room.RoomLock.Lock()
				room.Clients[client.ID] = client
				player := room.Board.AddPlayer(client.ID)
				room.RoomLock.Unlock()

				if player == nil {
					conn.Write([]byte("Room is full!\n"))
					return
				}

				conn.Write([]byte("Joined room: " + roomName + "\n"))
				room.Broadcast(room.Board.String())
				continue
			} else {
				conn.Write([]byte("Please JOIN <room> first.\n"))
				continue
			}
		}

		dx, dy := 0, 0
		switch msg {
		case "\x1b[A":
			dy = -1
		case "\x1b[B":
			dy = 1
		case "\x1b[C":
			dx = 1
		case "\x1b[D":
			dx = -1
		case "b":
			room.RoomLock.Lock()
			room.Board.PlantBomb(client.ID)
			room.RoomLock.Unlock()
			room.Broadcast(room.Board.String())

		default:
			conn.Write([]byte("Unknown command\n"))
			continue
		}

		room.RoomLock.Lock()
		moved := room.Board.MovePlayer(client.ID, dx, dy)
		room.RoomLock.Unlock()

		if moved || msg == "b" {
			room.Broadcast(room.Board.String())
		} else {
			conn.Write([]byte("Can't move\n"))
		}
	}

	if room != nil {
		room.RoomLock.Lock()
		delete(room.Clients, client.ID)
		room.Board.RemovePlayer(client.ID)
		room.RoomLock.Unlock()
		room.Broadcast(room.Board.String())
	}
}

func (s *Server) addClient(addr string, client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Clients[addr] = client
}

func (s *Server) removeClient(addr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Clients, addr)
}
