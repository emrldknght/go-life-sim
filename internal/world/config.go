package world

// Config — параметры мира
type Config struct {
	Width             float64
	Height            float64
	BaseTickMs        int64
	Speed             float64
	InitialPlants     int
	InitialHerbivores int
	InitialPredators  int
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
