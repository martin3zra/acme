import { jsxs, jsx } from "react/jsx-runtime";
import { u as useTranslation } from "./use-translation.js";
import "@inertiajs/react";
function LinesColumnHeaders() {
  const t = useTranslation().trans;
  return /* @__PURE__ */ jsxs("tr", { children: [
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-60 border border-gray-300 px-1 text-start", children: t("global.reference") }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-auto border border-gray-300 px-1 text-start", children: t("global.description") }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-36 border border-gray-300 px-1 text-start", children: t("global.unit") }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-36 border border-gray-300 px-1 text-end", children: t("global.qty") }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-36 border border-gray-300 px-1 text-end", children: t("global.price") }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-36 border border-gray-300 px-1 text-end", children: t("global.amount") }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-6 gap-2 border border-gray-300 px-5 text-end whitespace-nowrap" })
  ] });
}
export {
  LinesColumnHeaders as default
};
