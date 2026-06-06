package agent

// Типы агентов
const (
	TypePlant     = "plant"
	TypeHerbivore = "herbivore"
	TypePredator  = "predator"
)

// Colors Цвета для отображения
var Colors = map[string]string{
	TypePlant:     "#44ff44",
	TypeHerbivore: "#44aaff",
	TypePredator:  "#ff4444",
}

// Agent — отдельное существо в мире
type Agent struct {
	ID     int     `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Z      float64 `json:"z"`
	Type   string  `json:"type"`
	Color  string  `json:"color"`
	Energy float64 `json:"energy"`
}

// New создаёт нового агента
func New(id int, agentType string, x, y, z, energy float64) *Agent {
	return &Agent{
		ID:     id,
		X:      x,
		Y:      y,
		Z:      z,
		Type:   agentType,
		Color:  Colors[agentType],
		Energy: energy,
	}
}
