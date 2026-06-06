import * as THREE from 'three';
import { Agent } from './types';

export class SimulationRenderer {
	private scene: THREE.Scene;
	private camera: THREE.PerspectiveCamera;
	private renderer: THREE.WebGLRenderer;
	private agentsMap: Map<number, THREE.Mesh> = new Map();

	constructor() {
		this.scene = new THREE.Scene();
		this.scene.background = new THREE.Color(0x050b1a);
		this.scene.fog = new THREE.FogExp2(0x050b1a, 0.008);

		this.camera = new THREE.PerspectiveCamera(45, window.innerWidth / window.innerHeight, 0.1, 1000);
		this.camera.position.set(15, 18, 15);
		this.camera.lookAt(0, 0, 0);

		this.renderer = new THREE.WebGLRenderer({ antialias: true });
		this.renderer.setSize(window.innerWidth, window.innerHeight);
		this.renderer.shadowMap.enabled = true;
		document.body.appendChild(this.renderer.domElement);

		this.setupLighting();
		this.setupDecorations();
		this.startAnimationLoop();
	}

	private setupLighting(): void {
		const ambientLight = new THREE.AmbientLight(0x404060);
		this.scene.add(ambientLight);

		const mainLight = new THREE.DirectionalLight(0xffffff, 1);
		mainLight.position.set(10, 20, 5);
		mainLight.castShadow = true;
		mainLight.receiveShadow = true;
		mainLight.shadow.mapSize.width = 1024;
		mainLight.shadow.mapSize.height = 1024;
		this.scene.add(mainLight);

		const fillLight = new THREE.PointLight(0x4466cc, 0.3);
		fillLight.position.set(-5, 5, -5);
		this.scene.add(fillLight);

		const backLight = new THREE.PointLight(0xffaa66, 0.2);
		backLight.position.set(0, 5, -10);
		this.scene.add(backLight);
	}

	private setupDecorations(): void {
		// Земля
		const groundPlane = new THREE.Mesh(
				new THREE.PlaneGeometry(25, 25),
				new THREE.MeshStandardMaterial({ color: 0x0a0a1a, roughness: 0.8, metalness: 0.1, transparent: true, opacity: 0.6 })
		);
		groundPlane.rotation.x = -Math.PI / 2;
		groundPlane.position.y = -0.4;
		groundPlane.receiveShadow = true;
		this.scene.add(groundPlane);

		// Сетка
		const gridHelper = new THREE.GridHelper(30, 20, 0x3a5a8a, 0x1a2a4a);
		gridHelper.position.y = -0.3;
		this.scene.add(gridHelper);

		// Декоративные звёзды (частицы)
		const starGeometry = new THREE.BufferGeometry();
		const starCount = 500;
		const starPositions = new Float32Array(starCount * 3);
		for (let i = 0; i < starCount; i++) {
			starPositions[i * 3] = (Math.random() - 0.5) * 200;
			starPositions[i * 3 + 1] = (Math.random() - 0.5) * 50 + 10;
			starPositions[i * 3 + 2] = (Math.random() - 0.5) * 80 - 40;
		}
		starGeometry.setAttribute('position', new THREE.BufferAttribute(starPositions, 3));
		const starMaterial = new THREE.PointsMaterial({ color: 0xffffff, size: 0.08, transparent: true, opacity: 0.6 });
		const stars = new THREE.Points(starGeometry, starMaterial);
		this.scene.add(stars);
	}

	private createAgentMesh(color: string): THREE.Mesh {
		const geometry = new THREE.BoxGeometry(0.7, 0.7, 0.7);
		const material = new THREE.MeshStandardMaterial({
			color: color,
			roughness: 0.3,
			metalness: 0.1,
			emissive: color === '#44ff44' ? 0x113311 : 0x000000
		});
		const mesh = new THREE.Mesh(geometry, material);
		mesh.castShadow = true;
		mesh.receiveShadow = true;
		return mesh;
	}

	public updateAgents(agents: Agent[]): void {
		const receivedIds = new Set<number>();

		for (const agent of agents) {
			receivedIds.add(agent.id);

			if (this.agentsMap.has(agent.id)) {
				const mesh = this.agentsMap.get(agent.id)!;
				mesh.position.set(agent.x, agent.z, agent.y);
				// Обновляем цвет если изменился (для будущих мутаций)
				(mesh.material as THREE.MeshStandardMaterial).color.set(agent.color);
			} else {
				const mesh = this.createAgentMesh(agent.color);
				mesh.position.set(agent.x, agent.z, agent.y);
				this.scene.add(mesh);
				this.agentsMap.set(agent.id, mesh);
			}
		}

		// Удаляем умерших
		for (const [id, mesh] of this.agentsMap) {
			if (!receivedIds.has(id)) {
				this.scene.remove(mesh);
				this.agentsMap.delete(id);
			}
		}
	}

	private startAnimationLoop(): void {
		const animate = () => {
			requestAnimationFrame(animate);
			this.renderer.render(this.scene, this.camera);
		};
		animate();
	}

	public resize(width: number, height: number): void {
		this.camera.aspect = width / height;
		this.camera.updateProjectionMatrix();
		this.renderer.setSize(width, height);
	}
}