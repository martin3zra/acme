import { createInertiaApp } from '@inertiajs/react';
import { resolvePageComponent } from 'laravel-vite-plugin/inertia-helpers';
import { createRoot } from 'react-dom/client';
import { toast } from 'sonner';
import '../css/app.css';

createInertiaApp({
  resolve: (name) => resolvePageComponent(`./Pages/${name}.tsx`, import.meta.glob('./Pages/**/*.tsx')),
  setup({ el, App, props }) {
    createRoot(el).render(<App {...props} />);
  },
});

let onSuccessEvent = (e) => {
  const props = e.detail.page.props || { flash: {} };
  if (props.flash) {
    const message = props.flash.success;
    if (message) {
      setTimeout(() => {
        toast.success(message);
      }, 200);
    }

    delete props.flash;
  }
};

document.addEventListener('inertia:success', onSuccessEvent);
document.removeEventListener('inertia:finish', onSuccessEvent);
