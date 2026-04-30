import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwind from '@tailwindcss/vite'
import path from 'node:path'

const proxyTarget = process.env.VITE_API_PROXY || 'http://localhost:8080'

export default defineConfig({
	plugins: [vue(), tailwind()],
	resolve: {
		alias: {
			'@': path.resolve(__dirname, 'src')
		}
	},
	server: {
		port: 5173,
		proxy: {
			'/api': { target: proxyTarget, changeOrigin: true, ws: true },
			'/asynq': { target: proxyTarget, changeOrigin: true },
			'/healthz': { target: proxyTarget, changeOrigin: true }
		}
	},
	build: {
		outDir: 'dist',
		emptyOutDir: true,
		sourcemap: false
	}
})
