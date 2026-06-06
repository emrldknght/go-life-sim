import './style.css';
import { SimulationRenderer } from './renderer';
import { WorldConnection } from './world';
import { Agent } from './types';

// Инициализация
const renderer = new SimulationRenderer();
const worldConnection = new WorldConnection();

// Элементы UI
const plantCountEl = document.getElementById('plant-count');
const herbivoreCountEl = document.getElementById('herbivore-count');
const predatorCountEl = document.getElementById('predator-count');
const totalCountEl = document.getElementById('total-count');
const resetBtn = document.getElementById('reset-btn');
const pauseBtn = document.getElementById('pause-btn');
const statusEl = document.getElementById('status');

// Состояние
let isPaused = false;

function updateStats(agents: Agent[]): void {
	let plants = 0, herbivores = 0, predators = 0;

	for (const agent of agents) {
		switch (agent.type) {
			case 'plant': plants++; break;
			case 'herbivore': herbivores++; break;
			case 'predator': predators++; break;
		}
	}

	if (plantCountEl) plantCountEl.textContent = plants.toString();
	if (herbivoreCountEl) herbivoreCountEl.textContent = herbivores.toString();
	if (predatorCountEl) predatorCountEl.textContent = predators.toString();
	if (totalCountEl) totalCountEl.textContent = agents.length.toString();
}

// Подписка на обновления мира
worldConnection.onUpdate((agents) => {
	if (!isPaused) {
		renderer.updateAgents(agents);
		updateStats(agents);
	}
});

// Кнопки управления
resetBtn?.addEventListener('click', () => {
	worldConnection.sendCommand({ action: 'reset' });

	// Визуальная обратная связь
	if (resetBtn) {
		const originalText = resetBtn.textContent;
		resetBtn.textContent = '✅ Resetting...';
		setTimeout(() => {
			if (resetBtn) resetBtn.textContent = originalText;
		}, 1000);
	}

	// Если игра на паузе — снимаем паузу для отображения нового мира
	if (isPaused) {
		isPaused = false;
		if (pauseBtn) pauseBtn.textContent = '⏸️ Pause';
		if (statusEl) {
			statusEl.textContent = '🟢 Running';
			statusEl.style.opacity = '1';
		}
	}
});

pauseBtn?.addEventListener('click', () => {
	isPaused = !isPaused;
	if (pauseBtn) {
		pauseBtn.textContent = isPaused ? '▶️ Resume' : '⏸️ Pause';
	}
	if (statusEl) {
		statusEl.textContent = isPaused ? '⏸️ Paused' : '🟢 Running';
		statusEl.style.opacity = isPaused ? '0.7' : '1';
	}
});

// Адаптация под размер окна
window.addEventListener('resize', () => {
	renderer.resize(window.innerWidth, window.innerHeight);
});

// Подключаемся
worldConnection.connect();

console.log('🎮 Simulation frontend started');