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

export type WebSocketAction = 'reset' | 'pause' | 'resume' | 'set_speed';

export interface WebSocketCommand {
	action: WebSocketAction;
}
export interface SetSpeedCommand extends WebSocketCommand {
	action: 'set_speed';
	speed: number;
}

export type Command = WebSocketCommand | SetSpeedCommand;