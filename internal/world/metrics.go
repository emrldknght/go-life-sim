package world

import (
	"encoding/json"
	"time"
)

// Metrics — статистика производительности одного тика
type Metrics struct {
	UpdateDuration  time.Duration
	MarshalDuration time.Duration
	AgentCount      int
	SnapshotBytes   int
}

// UpdateWithMetrics — выполняет Update и возвращает метрики
func (w *World) UpdateWithMetrics() Metrics {
	start := time.Now()
	w.Update()
	updateDuration := time.Since(start)

	// Замеряем Marshal (только для статистики)
	marshalStart := time.Now()
	agents := w.GetAgentsDTO()
	data, _ := json.Marshal(agents)
	marshalDuration := time.Since(marshalStart)

	return Metrics{
		UpdateDuration:  updateDuration,
		MarshalDuration: marshalDuration,
		AgentCount:      len(agents),
		SnapshotBytes:   len(data),
	}
}
