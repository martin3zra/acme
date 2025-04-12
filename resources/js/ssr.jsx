import { createInertiaApp } from '@inertiajs/react';
import { createServer } from '@inertiajs/react/server';
import { resolvePageComponent } from 'laravel-vite-plugin/inertia-helpers';
import { createRoot } from 'react-dom/client';
import { renderToString } from 'react-dom/server';

createServer((page) => {
  createInertiaApp({
    page,
    render: renderToString,
    resolve: (name) => resolvePageComponent(`./Pages/${name}.tsx`, import.meta.glob('./Pages/**/*.tsx')),
    setup({ el, App, props }) {
      createRoot(el).render(<App {...props} />);
    },
  });
});
