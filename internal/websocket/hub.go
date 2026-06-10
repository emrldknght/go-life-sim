package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"mysimulation/internal/world"

	"github.com/coder/websocket"
)

// Command — структура команды от клиента
type Command struct {
	Action string  `json:"action"`
	Speed  float64 `json:"speed,omitempty"`
}

// Hub — управляет WebSocket соединениями
type Hub struct {
	world *world.World
}

// NewHub создаёт новый WebSocket хаб
func NewHub(w *world.World) *Hub {
	return &Hub{
		world: w,
	}
}

// Handler возвращает HTTP хендлер для WebSocket
func (h *Hub) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opts := &websocket.AcceptOptions{
			OriginPatterns: []string{
				"localhost:3000",
				"localhost:8080",
				"http://localhost:3000",
				"http://localhost:8080",
			},
		}

		conn, err := websocket.Accept(w, r, opts)
		if err != nil {
			log.Printf("WebSocket accept error: %v", err)
			return
		}

		// Безопасное закрытие с обработкой ошибки
		defer func() {
			err := conn.Close(websocket.StatusNormalClosure, "")
			if err != nil {
				log.Printf("WebSocket close error: %v", err)
			}
		}()

		log.Println("✅ Клиент подключился")
		h.handleConnection(conn)
	}
}

// handleConnection — обработка одного соединения
func (h *Hub) handleConnection(conn *websocket.Conn) {
	ctx := context.Background()

	// Канал для команд от клиента
	go h.handleCommands(conn, ctx)

	// Отправка данных клиенту
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		agents := h.world.GetAgents()
		data, err := json.Marshal(agents)
		if err != nil {
			log.Printf("JSON marshal error: %v", err)
			continue
		}

		err = conn.Write(ctx, websocket.MessageText, data)
		if err != nil {
			log.Printf("Клиент отключился (запись): %v", err)
			break
		}
	}
}

// handleCommands — чтение команд от клиента
// handleCommands — чтение команд от клиента
// handleCommands — чтение команд от клиента
func (h *Hub) handleCommands(conn *websocket.Conn, ctx context.Context) {
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			log.Printf("Клиент отключился (чтение): %v", err)
			break
		}

		var cmd Command
		if err := json.Unmarshal(msg, &cmd); err != nil {
			log.Printf("Ошибка парсинга команды: %v", err)
			continue
		}

		switch cmd.Action {
		case "reset":
			h.world.Reset()
			log.Println("🔄 Получена команда перезапуска")
		case "set_speed":
			if cmd.Speed >= 0.25 && cmd.Speed <= 4.0 {
				h.world.SetSpeed(cmd.Speed)
				log.Printf("⚡ Скорость изменена на: x%.2f", cmd.Speed)
			} else {
				log.Printf("⚠️ Неверное значение скорости: %.2f (должно быть 0.25-4.0)", cmd.Speed)
			}
		case "pause":
			h.world.SetPaused(true)
			log.Println("⏸️ Получена команда паузы")
		case "resume":
			h.world.SetPaused(false)
			log.Println("▶️ Получена команда возобновления")
		default:
			log.Printf("Неизвестная команда: %s", cmd.Action)
		}
	}
}
