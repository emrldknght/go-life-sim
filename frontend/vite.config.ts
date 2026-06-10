import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), '');
	const backend = env.VITE_BACKEND_URL ?? 'http://localhost:8080';

	return {
		server: {
			port: 3000,
			proxy: {
				'/ws': {
					target: backend,
					ws: true,
					changeOrigin: true,
				},
			},
		},
		build: {
			outDir: '../static',
			emptyOutDir: true,
		},
	};
});