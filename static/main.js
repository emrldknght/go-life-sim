import * as THREE from 'three';

// --- Сцена ---
const scene = new THREE.Scene();
scene.background = new THREE.Color(0x111122);

// --- Камера (вид сверху для лучшего обзора) ---
const camera = new THREE.PerspectiveCamera(45, window.innerWidth / window.innerHeight, 0.1, 1000);
camera.position.set(15, 15, 15);
camera.lookAt(0, 0, 0);

// --- Рендерер ---
const renderer = new THREE.WebGLRenderer({ antialias: true });
renderer.setSize(window.innerWidth, window.innerHeight);
document.body.appendChild(renderer.domElement);

// --- Освещение ---
const ambientLight = new THREE.AmbientLight(0x404060);
scene.add(ambientLight);
const mainLight = new THREE.DirectionalLight(0xffffff, 1);
mainLight.position.set(5, 10, 7);
scene.add(mainLight);

// --- Вспомогательная сетка ---
const gridHelper = new THREE.GridHelper(25, 20, 0x88aaff, 0x335588);
scene.add(gridHelper);

// --- Хранилище объектов ---
const agents = new Map(); // id -> mesh

// --- Функция создания меша по типу/цвету ---
function createAgentMesh(color) {
    const geometry = new THREE.BoxGeometry(0.6, 0.6, 0.6);
    const material = new THREE.MeshStandardMaterial({ color: color });
    const mesh = new THREE.Mesh(geometry, material);
    mesh.castShadow = true;
    return mesh;
}

// --- WebSocket ---
const socket = new WebSocket('ws://localhost:8080/ws');

socket.onopen = () => {
    console.log('✅ WebSocket подключён');
};

socket.onmessage = (event) => {
    const agentsData = JSON.parse(event.data);

    // Обновляем или создаём меши
    const receivedIds = new Set();

    for (const agent of agentsData) {
        receivedIds.add(agent.id);

        if (agents.has(agent.id)) {
            // Обновляем позицию существующего
            const mesh = agents.get(agent.id);
            mesh.position.set(agent.x, agent.y, agent.z);
        } else {
            // Создаём нового агента
            const mesh = createAgentMesh(agent.color);
            mesh.position.set(agent.x, agent.y, agent.z);
            scene.add(mesh);
            agents.set(agent.id, mesh);
            console.log(`➕ Добавлен агент ${agent.id}`);
        }
    }

    // Удаляем агентов, которые больше не существуют
    for (const [id, mesh] of agents) {
        if (!receivedIds.has(id)) {
            scene.remove(mesh);
            agents.delete(id);
            console.log(`➖ Удалён агент ${id}`);
        }
    }

    // Обновляем счётчик
    document.getElementById('count').textContent = agents.size;
};

socket.onerror = (error) => {
    console.error('❌ Ошибка:', error);
};

socket.onclose = () => {
    console.log('🔌 WebSocket закрыт');
};

// --- Анимация (просто перерисовка, позиции уже обновлены по WebSocket) ---
function animate() {
    requestAnimationFrame(animate);
    renderer.render(scene, camera);
}
animate();

// Адаптация под размер окна
window.addEventListener('resize', () => {
    camera.aspect = window.innerWidth / window.innerHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
});

console.log('🎮 Запущено, ожидаем агентов...');