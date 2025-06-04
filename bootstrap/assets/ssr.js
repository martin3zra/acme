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
    resolve: (name) => resolvePageComponent(`./Pages/${name}.tsx`, /* @__PURE__ */ Object.assign({ "./Pages/Auth/Login.tsx": () => import("./Login.js"), "./Pages/Customers/Index.tsx": () => import("./Index.js"), "./Pages/Customers/List/Index.tsx": () => import("./Index2.js"), "./Pages/Customers/List/columns-definitions.tsx": () => import("./columns-definitions.js"), "./Pages/Customers/Shared/CreateForm.tsx": () => import("./CreateForm.js"), "./Pages/Error/Index.tsx": () => import("./Index3.js"), "./Pages/Home/Index.tsx": () => import("./Index4.js"), "./Pages/Invoices/Create.tsx": () => import("./Create.js"), "./Pages/Invoices/Edit.tsx": () => import("./Edit.js"), "./Pages/Invoices/Index.tsx": () => import("./Index5.js"), "./Pages/Invoices/List/Index.tsx": () => import("./Index6.js"), "./Pages/Invoices/List/columns-definitions.tsx": () => import("./columns-definitions2.js"), "./Pages/Invoices/Shared/AddNewInvoice.tsx": () => import("./AddNewInvoice.js"), "./Pages/Invoices/Shared/bank-transfer.tsx": () => import("./bank-transfer.js"), "./Pages/Invoices/Shared/card-form.tsx": () => import("./card-form.js"), "./Pages/Invoices/Shared/check-form.tsx": () => import("./check-form.js"), "./Pages/Invoices/Shared/checkout-form.tsx": () => import("./checkout-form.js"), "./Pages/Invoices/Shared/customer-section.tsx": () => import("./customer-section.js").then((n) => n.c), "./Pages/Invoices/Shared/input-view.tsx": () => import("./input-view.js"), "./Pages/Invoices/Shared/lines-column-headers.tsx": () => import("./lines-column-headers.js"), "./Pages/Invoices/Shared/lines-form.tsx": () => import("./lines-form.js"), "./Pages/Invoices/Shared/lines-line.tsx": () => import("./lines-line.js"), "./Pages/Invoices/Shared/lines.tsx": () => import("./lines.js"), "./Pages/Invoices/Shared/payment-summary.tsx": () => import("./payment-summary.js"), "./Pages/Invoices/Show.tsx": () => import("./Show.js"), "./Pages/Items/Index.tsx": () => import("./Index7.js"), "./Pages/Items/List/Index.tsx": () => import("./Index8.js"), "./Pages/Items/List/columns-definitions.tsx": () => import("./columns-definitions3.js"), "./Pages/Items/Shared/CreateForm.tsx": () => import("./CreateForm2.js"), "./Pages/Onboarding/Index.tsx": () => import("./Index9.js"), "./Pages/Onboarding/Shared/confetti-ball.tsx": () => import("./confetti-ball.js"), "./Pages/Onboarding/Shared/congrats.tsx": () => import("./congrats.js"), "./Pages/Payments/Create.tsx": () => import("./Create2.js"), "./Pages/Payments/Edit.tsx": () => import("./Edit2.js"), "./Pages/Payments/Index.tsx": () => import("./Index10.js"), "./Pages/Payments/List/Index.tsx": () => import("./Index11.js"), "./Pages/Payments/List/columns-definitions.tsx": () => import("./columns-definitions4.js"), "./Pages/Payments/Shared/add-new-payment.tsx": () => import("./add-new-payment.js"), "./Pages/Payments/Shared/columns-definitions.tsx": () => import("./columns-definitions5.js"), "./Pages/Payments/Shared/lines-payment.tsx": () => import("./lines-payment.js"), "./Pages/Payments/Show.tsx": () => import("./Show2.js"), "./Pages/Verify/Index.tsx": () => import("./Index12.js"), "./Pages/Verify/Shared/create-password.tsx": () => import("./create-password.js") })),
    setup({ el, App, props }) {
      createRoot(el).render(/* @__PURE__ */ jsx(App, { ...props }));
    }
  });
});
