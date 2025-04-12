import { jsx } from "react/jsx-runtime";
import { createInertiaApp } from "@inertiajs/react";
import { createServer } from "@inertiajs/react/server";
import { createRoot } from "react-dom/client";
import { renderToString } from "react-dom/server";
async function resolvePageComponent(path, pages) {
  for (const p of Array.isArray(path) ? path : [path]) {
    const page = pages[p];
    if (typeof page === "undefined") {
      continue;
    }
    return typeof page === "function" ? page() : page;
  }
  throw new Error(`Page not found: ${path}`);
}
createServer((page) => {
  createInertiaApp({
    page,
    render: renderToString,
    resolve: (name) => resolvePageComponent(`./Pages/${name}.tsx`, /* @__PURE__ */ Object.assign({ "./Pages/Auth/Login.tsx": () => import("./Login.js"), "./Pages/Customers/Index.tsx": () => import("./Index.js"), "./Pages/Customers/List/Index.tsx": () => import("./Index2.js"), "./Pages/Customers/List/columns-definitions.tsx": () => import("./columns-definitions.js"), "./Pages/Customers/Shared/CreateForm.tsx": () => import("./CreateForm.js").then((n) => n.a), "./Pages/Customers/Shared/delete-customer.tsx": () => import("./delete-customer.js"), "./Pages/Error/Index.tsx": () => import("./Index3.js"), "./Pages/Home/Index.tsx": () => import("./Index4.js") })),
    setup({ el, App, props }) {
      createRoot(el).render(/* @__PURE__ */ jsx(App, { ...props }));
    }
  });
});
