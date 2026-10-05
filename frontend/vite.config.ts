import react from '@vitejs/plugin-react'
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  build: {
    outDir: '../internal/web/dist',
    emptyOutDir: true,
    assetsDir: 'assets',
    // O paquete inicial (React + Mantine) ocupa ~550 kB (~170 kB con gzip);
    // as gráficas (Recharts) cárganse á parte e só cando se usan. Para unha
    // app servida na rede local é aceptable.
    chunkSizeWarningLimit: 650,
  },
  test: {
    environment: 'jsdom',
    setupFiles: 'src/test/setup.ts',
    globals: true,
    // En Windows o pool "forks" ás veces non consegue arrancar os workers.
    pool: 'threads',
    // Arrancar moitos workers á vez (jsdom + Mantine) supera o tempo de
    // arranque nesta máquina e Vitest acaba sen executar os tests.
    maxWorkers: 2,
  },
})
