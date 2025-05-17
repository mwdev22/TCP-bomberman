package board

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Tile rune

const (
	Wall      Tile = '#' // undestroyable
	Breakable Tile = '*' // destroyable
	Empty     Tile = ' '
	PlayerT   Tile = 'P'
	BombT     Tile = 'B'
	Explosion Tile = 'X'
)

type Board struct {
	Width, Height int
	Grid          [][]Tile
	Players       map[string]*Player
	Explosions    []ExplosionTile
	Bombs         []*Bomb
}

type ExplosionTile struct {
	X, Y      int
	CreatedAt int64
	Duration  int64
}

func New(width, height int) *Board {
	grid := make([][]Tile, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			// walls
			if x == 0 || y == 0 || x == width-1 || y == height-1 {
				grid[y][x] = Wall
			} else if x%2 == 0 && y%2 == 0 {

				grid[y][x] = Wall
			} else {
				// random: free space or breakable
				if rand.Float64() < 0.2 {
					grid[y][x] = Breakable
				} else {
					grid[y][x] = Empty
				}
			}
		}
	}

	return &Board{
		Width:   width,
		Height:  height,
		Grid:    grid,
		Players: make(map[string]*Player),
		Bombs:   make([]*Bomb, 0),
	}
}

func (m *Board) Print() {
	for _, row := range m.Grid {
		for _, tile := range row {
			fmt.Printf("%c", tile)
		}
		fmt.Println()
	}
}

func (b *Board) Tick() ([]string, bool) {
	now := time.Now().UnixNano()
	var remainingBombs []*Bomb
	changed := false
	var destroyedPlayers []string

	for _, bomb := range b.Bombs {
		if now-bomb.PlantedAt >= bomb.ExplodesIn {
			ids := b.explode(bomb)
			destroyedPlayers = append(destroyedPlayers, ids...)
			changed = true
		} else {
			remainingBombs = append(remainingBombs, bomb)
		}
	}
	b.Bombs = remainingBombs

	var remainingExplosions []ExplosionTile
	for _, exp := range b.Explosions {
		if now-exp.CreatedAt >= exp.Duration {
			if b.Grid[exp.Y][exp.X] == Explosion {
				b.Grid[exp.Y][exp.X] = Empty
				changed = true
			}
		} else {
			remainingExplosions = append(remainingExplosions, exp)
		}
	}
	b.Explosions = remainingExplosions

	return destroyedPlayers, changed
}

func (b *Board) explode(bomb *Bomb) []string {
	destroyedPlayers := []string{}

	positions := [][2]int{
		{bomb.X, bomb.Y},
		{bomb.X + 1, bomb.Y},
		{bomb.X - 1, bomb.Y},
		{bomb.X, bomb.Y + 1},
		{bomb.X, bomb.Y - 1},
	}

	now := time.Now().UnixNano()

	for _, pos := range positions {
		x, y := pos[0], pos[1]
		if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
			continue
		}

		tile := b.Grid[y][x]
		if tile == Wall {
			continue
		}

		for id, p := range b.Players {
			if p.X == x && p.Y == y {
				destroyedPlayers = append(destroyedPlayers, id)
				delete(b.Players, id)
			}
		}

		// Set explosion tile
		b.Grid[y][x] = Explosion

		// Track explosion for removal after 1s
		b.Explosions = append(b.Explosions, ExplosionTile{
			X:         x,
			Y:         y,
			CreatedAt: now,
			Duration:  int64(time.Second),
		})
	}

	return destroyedPlayers
}

func (b *Board) PlantBomb(playerID string) {
	player, ok := b.Players[playerID]
	if !ok {
		return
	}

	if b.Grid[player.Y][player.X] == BombT {
		return
	}
	b.Grid[player.Y][player.X] = BombT

	b.Bombs = append(b.Bombs, &Bomb{
		X:          player.X,
		Y:          player.Y,
		OwnerID:    playerID,
		PlantedAt:  time.Now().UnixNano(),
		ExplodesIn: int64(3 * time.Second), // bomb explodes in 3s
	})
}

func (b *Board) String() string {
	var sb strings.Builder
	for _, row := range b.Grid {
		for _, tile := range row {
			sb.WriteRune(rune(tile))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func (b *Board) AddPlayer(id string) *Player {
	emptyPositions := [][2]int{}
	for y := 1; y < b.Height-1; y++ {
		for x := 1; x < b.Width-1; x++ {
			if b.Grid[y][x] == Empty {
				emptyPositions = append(emptyPositions, [2]int{x, y})
			}
		}
	}

	if len(emptyPositions) == 0 {
		return nil // full board
	}

	// Jeśli nie ma innych graczy, wybierz losowo
	if len(b.Players) == 0 {
		pos := emptyPositions[rand.Intn(len(emptyPositions))]
		player := &Player{ID: id, X: pos[0], Y: pos[1]}
		b.Players[id] = player
		b.Grid[pos[1]][pos[0]] = PlayerT
		return player
	}

	manhattan := func(x1, y1, x2, y2 int) int {
		dx := x1 - x2
		if dx < 0 {
			dx = -dx
		}
		dy := y1 - y2
		if dy < 0 {
			dy = -dy
		}
		return dx + dy
	}

	var bestPos [2]int
	maxMinDist := -1

	for _, pos := range emptyPositions {
		minDist := 1<<31 - 1 // max int
		for _, p := range b.Players {
			dist := manhattan(pos[0], pos[1], p.X, p.Y)
			if dist < minDist {
				minDist = dist
			}
		}

		if minDist > maxMinDist {
			maxMinDist = minDist
			bestPos = pos
		}
	}

	player := &Player{ID: id, X: bestPos[0], Y: bestPos[1]}
	b.Players[id] = player
	b.Grid[bestPos[1]][bestPos[0]] = PlayerT
	return player
}

func (b *Board) RemovePlayer(id string) {
	player, ok := b.Players[id]
	if !ok {
		return
	}

	b.Grid[player.Y][player.X] = Empty
	delete(b.Players, id)
}

func (b *Board) MovePlayer(id string, dx, dy int) bool {
	player, ok := b.Players[id]
	if !ok {
		return false
	}

	newX := player.X + dx
	newY := player.Y + dy

	if newX < 0 || newX >= b.Width || newY < 0 || newY >= b.Height {
		return false
	}

	if b.Grid[newY][newX] != Empty {
		return false
	}

	if b.Grid[player.Y][player.X] == PlayerT {
		b.Grid[player.Y][player.X] = Empty
	}

	b.Grid[newY][newX] = PlayerT
	player.X = newX
	player.Y = newY
	return true
}
