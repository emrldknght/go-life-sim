export type AgentType = 'plant' | 'herbivore' | 'predator';

export interface Agent {
	id: number;
	x: number;
	y: number;
	z: number;
	type: AgentType;
	color: string;
	energy: number;
}

export interface WorldState {
	agents: Agent[];
	timestamp: number;
}

export interface WebSocketCommand {
	action: 'reset' | 'pause' | 'resume';
}