package board

type Bomb struct {
	X, Y       int
	OwnerID    string
	PlantedAt  int64
	ExplodesIn int64
}
