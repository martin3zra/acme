import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import laravel from 'laravel-vite-plugin';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
  // In order to have all config on the .env, here we take
  // the APP_URL that contains the HTTP protocol and PORT number
  // and extract the IP numbers only.
  const env = loadEnv(mode, process.cwd(), '');
  var regx = /\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b/;
  const viteAppURL = env.APP_URL.split(':')[1];

  // If the .env use localhost we default to 0.0.0.0
  let host = '0.0.0.0';
  const matches = viteAppURL.match(regx);
  if (matches !== null && matches.length > 0) {
    host = matches[0];
  }

  return {
    plugins: [
      laravel({
        input: ['resources/js/app.jsx', 'resources/css/app.css'],
        publicDirectory: 'public',
        buildDirectory: 'build',
        refresh: true,
      }),
      react({ include: /\.(mdx|js|jsx|ts|tsx)$/ }),
      tailwindcss(),
    ],
    esbuild: {
      jsx: 'automatic',
    },
    build: {
      manifest: true, // Generate manifest.json file
      outDir: 'public/build',
      rollupOptions: {
        input: ['resources/js/app.jsx', 'resources/css/app.css'],
        output: {
          entryFileNames: 'assets/[name].js',
          chunkFileNames: 'assets/[name].js',
          assetFileNames: 'assets/[name].[ext]',
          manualChunks: undefined, // Disable automatic chunk splitting
        },
      },
    },
    server: {
      hmr: {
        host,
      },
      host: true,
    },
  };
});
