import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
	plugins: [vue()],
	resolve: {
		alias: {
			'@': resolve(__dirname, 'src')
		}
	},
	test: {
		environment: 'happy-dom',
		include: ['src/**/*.test.ts', 'src/**/*.spec.ts'],
		coverage: {
			provider: 'v8',
			reporter: ['text', 'html'],
			// V2 preset system (this iteration's surface area).
			include: [
				'src/api/client.ts',
				'src/composables/presetDraft.ts',
				'src/stores/presetCatalog.ts',
				'src/stores/presets.ts',
				'src/views/PresetsView.vue',
				'src/views/PresetEditorView.vue',
				'src/components/common/Button.vue',
				'src/components/common/EmptyState.vue'
			],
			exclude: ['src/**/*.test.ts', 'src/**/*.spec.ts', 'src/**/*.d.ts'],
			thresholds: {
				statements: 80,
				lines: 80,
				functions: 80,
				branches: 70
			}
		}
	}
})
