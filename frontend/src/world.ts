import { Agent, Command } from './types';

type WorldUpdateCallback = (agents: Agent[]) => void;

export class WorldConnection {
	private socket: WebSocket | null = null;
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;
	private callbacks: WorldUpdateCallback[] = [];

	/** Same-origin WebSocket: dev → Vite proxy, prod → Go на том же host:port */
	private getWebSocketUrl(): string {
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		return `${protocol}//${window.location.host}/ws`;
	}

	connect(): void {
		const wsUrl = this.getWebSocketUrl();
		console.log(`🔄 Подключение к WebSocket: ${wsUrl}`);

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

	sendCommand(command: Command): void {
		if (this.socket && this.socket.readyState === WebSocket.OPEN) {
			this.socket.send(JSON.stringify(command));
		} else {
			console.warn('WebSocket not connected, command not sent');
		}
	}

	setSpeed(speed: number): void {
		this.sendCommand({ action: 'set_speed', speed });
	}

	reset(): void {
		this.sendCommand({ action: 'reset' });
	}

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