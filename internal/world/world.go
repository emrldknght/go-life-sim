package world

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"mysimulation/internal/agent"
)

// World — основной объект мира
type World struct {
	mu     sync.RWMutex
	agents map[int]*agent.Agent
	nextID int
	cfg    Config
	paused bool
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
		paused: false,
	}
	w.Reset()
	return w
}

// SetPaused — устанавливает состояние паузы
func (w *World) SetPaused(paused bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.paused = paused
	if paused {
		log.Println("⏸️ Симуляция на паузе")
	} else {
		log.Println("▶️ Симуляция возобновлена")
	}
}

// IsPaused — возвращает состояние паузы
func (w *World) IsPaused() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.paused
}

// SetSpeed — устанавливает скорость симуляции
func (w *World) SetSpeed(speed float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

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
		tickMs = 10
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

	for i := 0; i < w.cfg.InitialPlants; i++ {
		w.addAgentLocked(agent.TypePlant, 80.0)
	}
	for i := 0; i < w.cfg.InitialHerbivores; i++ {
		w.addAgentLocked(agent.TypeHerbivore, 60.0)
	}
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
