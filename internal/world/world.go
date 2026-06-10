package world

import (
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"mysimulation/internal/agent"
)

// Config — параметры мира
type Config struct {
	Width             float64 // Ширина мира по X
	Height            float64 // Высота мира по Z
	BaseTickMs        int64   // базовый тик в миллисекундах
	Speed             float64 // множитель скорости (0.25, 0.5, 1, 2, 4)
	InitialPlants     int     // начальное количество растений
	InitialHerbivores int     // начальное количество травоядных
	InitialPredators  int     // начальное количество хищников
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	return Config{
		Width:             20.0,
		Height:            20.0,
		BaseTickMs:        50,
		Speed:             1.0,
		InitialPlants:     25,
		InitialHerbivores: 12,
		InitialPredators:  4,
	}
}

// World — основной объект мира
type World struct {
	mu     sync.RWMutex
	agents map[int]*agent.Agent
	nextID int
	cfg    Config
}

// New создаёт новый мир с конфигурацией по умолчанию
func New() *World {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig создаёт новый мир с пользовательской конфигурацией
func NewWithConfig(cfg Config) *World {
	w := &World{
		agents: make(map[int]*agent.Agent),
		nextID: 1,
		cfg:    cfg,
	}
	w.Reset()
	return w
}

// SetSpeed — устанавливает скорость симуляции
func (w *World) SetSpeed(speed float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Ограничиваем скорость разумными пределами
	if speed < 0.25 {
		speed = 0.25
	}
	if speed > 4.0 {
		speed = 4.0
	}

	w.cfg.Speed = speed
	log.Printf("⚡ Скорость симуляции изменена: x%.2f", speed)
}

// GetSpeed — возвращает текущую скорость
func (w *World) GetSpeed() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.cfg.Speed
}

// GetTickDuration — возвращает текущую длительность тика с учётом скорости
func (w *World) GetTickDuration() time.Duration {
	w.mu.RLock()
	defer w.mu.RUnlock()

	tickMs := float64(w.cfg.BaseTickMs) / w.cfg.Speed
	if tickMs < 10 {
		tickMs = 10 // Минимум 10 мс (100 FPS симуляции)
	}
	return time.Duration(tickMs) * time.Millisecond
}

// Reset — очищает мир и создаёт начальную популяцию
func (w *World) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.resetLocked()
	log.Println("🌍 Мир перезапущен!")
}

// resetLocked — внутренний метод без захвата мьютекса
func (w *World) resetLocked() {
	w.agents = make(map[int]*agent.Agent)
	w.nextID = 1

	// Растения
	for i := 0; i < w.cfg.InitialPlants; i++ {
		w.addAgentLocked(agent.TypePlant, 80.0)
	}

	// Травоядные
	for i := 0; i < w.cfg.InitialHerbivores; i++ {
		w.addAgentLocked(agent.TypeHerbivore, 60.0)
	}

	// Хищники
	for i := 0; i < w.cfg.InitialPredators; i++ {
		w.addAgentLocked(agent.TypePredator, 80.0)
	}
}

// addAgentLocked — добавляет агента (без захвата мьютекса)
func (w *World) addAgentLocked(agentType string, energy float64) *agent.Agent {
	x := (rand.Float64() - 0.5) * w.cfg.Width
	z := (rand.Float64() - 0.5) * w.cfg.Height

	a := agent.New(w.nextID, agentType, x, 0.5, z, energy)
	w.agents[w.nextID] = a
	w.nextID++
	return a
}

// AddAgent — публичный метод добавления агента
func (w *World) AddAgent(agentType string, energy float64) *agent.Agent {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.addAgentLocked(agentType, energy)
}

// removeAgentLocked — удаляет агента (без захвата мьютекса)
func (w *World) removeAgentLocked(id int) {
	delete(w.agents, id)
}

// GetAgents — возвращает всех агентов (копию слайса)
func (w *World) GetAgents() []*agent.Agent {
	w.mu.RLock()
	defer w.mu.RUnlock()

	agentsList := make([]*agent.Agent, 0, len(w.agents))
	for _, a := range w.agents {
		agentsList = append(agentsList, a)
	}
	return agentsList
}

// Update — один шаг симуляции
func (w *World) Update() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Получаем текущий множитель скорости
	speed := w.cfg.Speed

	// 1. Движение (скорость движения зависит от множителя)
	for _, a := range w.agents {
		w.moveAgentLocked(a, speed)
	}

	// 2. Поедание (радиус не меняется, но частота обновлений уже выше)
	for _, a := range w.agents {
		w.handleEatingLocked(a)
	}

	// 3. Трата энергии и смерть (трата зависит от скорости)
	toRemove := []int{}
	for _, a := range w.agents {
		w.spendEnergyLocked(a, speed)
		if a.Energy <= 0 {
			toRemove = append(toRemove, a.ID)
		}
	}
	for _, id := range toRemove {
		w.removeAgentLocked(id)
	}

	// 4. Размножение (шанс зависит от скорости)
	for _, a := range w.agents {
		w.tryReproduceLocked(a, speed)
	}

	// 5. Восполнение растений (скорость роста зависит от скорости мира)
	w.growPlantsLocked(speed)
}

// moveAgentLocked — движение агента с учётом скорости мира
func (w *World) moveAgentLocked(a *agent.Agent, worldSpeed float64) {
	if a.Type == agent.TypePlant {
		return
	}

	// Базовая скорость агента
	baseSpeed := 0.4
	if a.Type == agent.TypePredator {
		baseSpeed = 0.5
	}
	// Итоговая скорость = базовая * скорость мира
	speed := baseSpeed * worldSpeed

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
		// ИСПРАВЛЕНО: используем math.Sqrt для получения реального расстояния
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
		a.Z = -limit
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

// spendEnergyLocked — трата энергии с учётом скорости мира
func (w *World) spendEnergyLocked(a *agent.Agent, worldSpeed float64) {
	// Базовая стоимость
	baseCost := 0.5
	switch a.Type {
	case agent.TypeHerbivore:
		baseCost = 0.6
	case agent.TypePredator:
		baseCost = 0.8
	case agent.TypePlant:
		baseCost = 0.1
	}
	// Стоимость умножается на скорость мира
	cost := baseCost * worldSpeed
	a.Energy -= cost
}

// tryReproduceLocked — размножение с учётом скорости мира
func (w *World) tryReproduceLocked(a *agent.Agent, worldSpeed float64) {
	requiredEnergy := 55.0
	if a.Energy < requiredEnergy {
		return
	}

	// Базовый шанс
	baseChance := 0.02
	switch a.Type {
	case agent.TypeHerbivore:
		baseChance = 0.04
	case agent.TypePredator:
		baseChance = 0.025
	case agent.TypePlant:
		baseChance = 0.05
	}

	// Шанс умножается на скорость мира (но не более 0.15)
	chance := baseChance * worldSpeed
	if chance > 0.15 {
		chance = 0.15
	}

	if rand.Float64() < chance {
		a.Energy -= 25
		w.addAgentLocked(a.Type, 35.0)
	}
}

// growPlantsLocked — восстановление растений с учётом скорости
func (w *World) growPlantsLocked(worldSpeed float64) {
	plantCount := 0
	for _, a := range w.agents {
		if a.Type == agent.TypePlant {
			plantCount++
		}
	}

	// Целевое количество растений
	targetPlants := 15

	if plantCount < targetPlants {
		// Количество новых растений зависит от скорости
		newPlants := int(float64(targetPlants-plantCount) * worldSpeed)
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
