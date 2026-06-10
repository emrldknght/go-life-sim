package world

import (
	"math"
	"math/rand"

	"mysimulation/internal/agent"
)

// Update — один шаг симуляции
func (w *World) Update() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.paused {
		return
	}

	// 1. Движение
	for _, a := range w.agents {
		w.moveAgentLocked(a)
	}

	// 2. Поедание
	for _, a := range w.agents {
		w.handleEatingLocked(a)
	}

	// 3. Трата энергии и смерть
	toRemove := []int{}
	for _, a := range w.agents {
		w.spendEnergyLocked(a)
		if a.Energy <= 0 {
			toRemove = append(toRemove, a.ID)
		}
	}
	for _, id := range toRemove {
		w.removeAgentLocked(id)
	}

	// 4. Размножение
	for _, a := range w.agents {
		w.tryReproduceLocked(a)
	}

	// 5. Восполнение растений
	w.growPlantsLocked()
}

// moveAgentLocked — движение агента
func (w *World) moveAgentLocked(a *agent.Agent) {
	if a.Type == agent.TypePlant {
		return
	}

	speed := 0.4
	if a.Type == agent.TypePredator {
		speed = 0.5
	}

	// Поиск цели
	var target *agent.Agent
	minDist := 100.0

	for _, other := range w.agents {
		if other == a {
			continue
		}
		isTarget := false
		switch a.Type {
		case agent.TypeHerbivore:
			isTarget = other.Type == agent.TypePlant
		case agent.TypePredator:
			isTarget = other.Type == agent.TypeHerbivore
		}
		if isTarget {
			dist := distance(a, other)
			if dist < minDist {
				minDist = dist
				target = other
			}
		}
	}

	if target != nil {
		dx := target.X - a.X
		dz := target.Z - a.Z
		length := math.Sqrt(dx*dx + dz*dz)
		if length > 0 {
			dx /= length
			dz /= length
			a.X += dx * speed
			a.Z += dz * speed
		}
	} else {
		a.X += (rand.Float64() - 0.5) * speed
		a.Z += (rand.Float64() - 0.5) * speed
	}

	limit := w.cfg.Width/2 - 1
	if a.X > limit {
		a.X = limit
	}
	if a.X < -limit {
		a.X = -limit
	}
	if a.Z > limit {
		a.Z = limit
	}
	if a.Z < -limit {
		a.Z = limit
	}
}

// handleEatingLocked — поедание
func (w *World) handleEatingLocked(a *agent.Agent) {
	eatRadius := 1.5

	switch a.Type {
	case agent.TypeHerbivore:
		var nearest *agent.Agent
		minDist := eatRadius
		for _, other := range w.agents {
			if other.Type == agent.TypePlant && other != a {
				dist := distance(a, other)
				if dist < minDist {
					minDist = dist
					nearest = other
				}
			}
		}
		if nearest != nil {
			w.removeAgentLocked(nearest.ID)
			a.Energy += 30
			if a.Energy > 100 {
				a.Energy = 100
			}
		}

	case agent.TypePredator:
		var nearest *agent.Agent
		minDist := eatRadius
		for _, other := range w.agents {
			if other.Type == agent.TypeHerbivore && other != a {
				dist := distance(a, other)
				if dist < minDist {
					minDist = dist
					nearest = other
				}
			}
		}
		if nearest != nil {
			w.removeAgentLocked(nearest.ID)
			a.Energy += 40
			if a.Energy > 100 {
				a.Energy = 100
			}
		}
	}
}

// spendEnergyLocked — трата энергии
func (w *World) spendEnergyLocked(a *agent.Agent) {
	baseCost := 0.5
	switch a.Type {
	case agent.TypeHerbivore:
		baseCost = 0.6
	case agent.TypePredator:
		baseCost = 0.8
	case agent.TypePlant:
		baseCost = 0.1
	}
	a.Energy -= baseCost
}

// tryReproduceLocked — размножение
func (w *World) tryReproduceLocked(a *agent.Agent) {
	requiredEnergy := 55.0
	if a.Energy < requiredEnergy {
		return
	}

	chance := 0.02
	switch a.Type {
	case agent.TypeHerbivore:
		chance = 0.04
	case agent.TypePredator:
		chance = 0.025
	case agent.TypePlant:
		chance = 0.05
	}

	if rand.Float64() < chance {
		a.Energy -= 25
		w.addAgentLocked(a.Type, 35.0)
	}
}

// growPlantsLocked — восстановление растений
func (w *World) growPlantsLocked() {
	plantCount := 0
	for _, a := range w.agents {
		if a.Type == agent.TypePlant {
			plantCount++
		}
	}

	targetPlants := 15

	if plantCount < targetPlants {
		newPlants := targetPlants - plantCount
		if newPlants < 1 {
			newPlants = 1
		}
		if newPlants > 10 {
			newPlants = 10
		}
		for i := 0; i < newPlants; i++ {
			w.addAgentLocked(agent.TypePlant, 50.0)
		}
	}
}

// distance — реальное расстояние между агентами
func distance(a, b *agent.Agent) float64 {
	dx := a.X - b.X
	dz := a.Z - b.Z
	return math.Sqrt(dx*dx + dz*dz)
}
