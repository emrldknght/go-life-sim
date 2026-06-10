package world

// AgentDTO — копия агента для безопасной сериализации
type AgentDTO struct {
	ID     int     `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Z      float64 `json:"z"`
	Type   string  `json:"type"`
	Color  string  `json:"color"`
	Energy float64 `json:"energy"`
}

// GetAgentsDTO — возвращает безопасную копию всех агентов для JSON
func (w *World) GetAgentsDTO() []AgentDTO {
	w.mu.RLock()
	defer w.mu.RUnlock()

	agentsList := make([]AgentDTO, 0, len(w.agents))
	for _, a := range w.agents {
		agentsList = append(agentsList, AgentDTO{
			ID:     a.ID,
			X:      a.X,
			Y:      a.Y,
			Z:      a.Z,
			Type:   a.Type,
			Color:  a.Color,
			Energy: a.Energy,
		})
	}
	return agentsList
}
