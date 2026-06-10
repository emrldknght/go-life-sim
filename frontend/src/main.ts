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
const speedBtns = document.querySelectorAll('.speed-btn');

// Состояние
let isPaused = false;
let currentSpeed = 1;

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

// Функция изменения скорости
function setSpeed(speed: number): void {
	currentSpeed = speed;
	worldConnection.setSpeed(speed);

	// Обновляем активную кнопку
	speedBtns.forEach(btn => {
		const btnSpeed = parseFloat(btn.getAttribute('data-speed') || '1');
		if (btnSpeed === speed) {
			btn.classList.add('active');
		} else {
			btn.classList.remove('active');
		}
	});

	// Обновляем статус
	if (statusEl && !isPaused) {
		const speedText = speed === 1 ? '' : ` x${speed}`;
		statusEl.textContent = `🟢 Running${speedText}`;
	}
}

// Подписка на обновления мира
worldConnection.onUpdate((agents) => {
	if (!isPaused) {
		renderer.updateAgents(agents);
		updateStats(agents);
	}
});

// Кнопки управления скоростью
speedBtns.forEach(btn => {
	btn.addEventListener('click', () => {
		const speed = parseFloat(btn.getAttribute('data-speed') || '1');
		setSpeed(speed);
	});
});

// Кнопка перезапуска
resetBtn?.addEventListener('click', () => {
	worldConnection.reset();

	if (resetBtn) {
		const originalText = resetBtn.textContent;
		resetBtn.textContent = '✅ Resetting...';
		setTimeout(() => {
			if (resetBtn) resetBtn.textContent = originalText;
		}, 1000);
	}

	// Если игра на паузе — снимаем паузу
	if (isPaused) {
		isPaused = false;
		if (pauseBtn) pauseBtn.textContent = '⏸️ Pause';
		if (statusEl) {
			const speedText = currentSpeed === 1 ? '' : ` x${currentSpeed}`;
			statusEl.textContent = `🟢 Running${speedText}`;
			statusEl.style.opacity = '1';
		}
	}
});

// Кнопка паузы
pauseBtn?.addEventListener('click', () => {
	isPaused = !isPaused;
	if (isPaused) {
		worldConnection.pause();
	} else {
		worldConnection.resume();
	}

	if (pauseBtn) {
		pauseBtn.textContent = isPaused ? '▶️ Resume' : '⏸️ Pause';
	}
	if (statusEl) {
		if (isPaused) {
			statusEl.textContent = '⏸️ Paused';
		} else {
			const speedText = currentSpeed === 1 ? '' : ` x${currentSpeed}`;
			statusEl.textContent = `🟢 Running${speedText}`;
		}
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