import { Agent, Command } from './types';

type WorldUpdateCallback = (agents: Agent[]) => void;

export class WorldConnection {
	private socket: WebSocket | null = null;
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;
	private callbacks: WorldUpdateCallback[] = [];

	connect(): void {
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const wsUrl = `${protocol}//${window.location.hostname}:8080/ws`;

		this.socket = new WebSocket(wsUrl);

		this.socket.onopen = () => {
			console.log('✅ WebSocket connected');
			this.reconnectAttempts = 0;
		};

		this.socket.onmessage = (event) => {
			try {
				const agents = JSON.parse(event.data) as Agent[];
				this.callbacks.forEach(cb => cb(agents));
			} catch (err) {
				console.error('Failed to parse WebSocket message:', err);
			}
		};

		this.socket.onclose = () => {
			console.log('WebSocket closed');
			this.attemptReconnect();
		};

		this.socket.onerror = (error) => {
			console.error('WebSocket error:', error);
		};
	}

	private attemptReconnect(): void {
		if (this.reconnectAttempts < this.maxReconnectAttempts) {
			this.reconnectAttempts++;
			console.log(`Reconnecting in 2s... (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
			setTimeout(() => this.connect(), 2000);
		}
	}

	// Отправка команды (принимает любой допустимый тип)
	sendCommand(command: Command): void {
		if (this.socket && this.socket.readyState === WebSocket.OPEN) {
			this.socket.send(JSON.stringify(command));
		} else {
			console.warn('WebSocket not connected, command not sent');
		}
	}

	// Удобный метод для смены скорости
	setSpeed(speed: number): void {
		this.sendCommand({ action: 'set_speed', speed });
	}

	// Удобный метод для сброса
	reset(): void {
		this.sendCommand({ action: 'reset' });
	}

	// Удобный метод для паузы/возобновления
	pause(): void {
		this.sendCommand({ action: 'pause' });
	}

	resume(): void {
		this.sendCommand({ action: 'resume' });
	}

	onUpdate(callback: WorldUpdateCallback): () => void {
		this.callbacks.push(callback);
		return () => {
			this.callbacks = this.callbacks.filter(cb => cb !== callback);
		};
	}

	disconnect(): void {
		if (this.socket) {
			this.socket.close();
			this.socket = null;
		}
	}
}