import { jsxs, jsx } from "react/jsx-runtime";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { u as useTranslation } from "./use-translation.js";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-label";
import "@inertiajs/react";
function LinesForm({ referenceInputRef, qtyInputRef, currentItem, amount, currency, handleOnKeyDown, computedItemAmount }) {
  const t = useTranslation().trans;
  return /* @__PURE__ */ jsxs("tr", { children: [
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-60 border border-gray-300 pe-1", children: /* @__PURE__ */ jsx(
      Input,
      {
        name: "reference",
        ref: referenceInputRef,
        "data-reset": false,
        placeholder: t("invoices.line.form.placeholder"),
        onKeyDown: handleOnKeyDown,
        className: "rounded-none border-none focus-visible:border-none focus-visible:ring-[2px]",
        tabIndex: 0
      }
    ) }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-auto border border-gray-300 bg-gray-50 px-1", children: /* @__PURE__ */ jsx(Label, { children: currentItem == null ? void 0 : currentItem.description }) }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-36 border border-gray-300 bg-gray-50 px-1", children: /* @__PURE__ */ jsx(Label, { children: currentItem == null ? void 0 : currentItem.unit.name }) }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-36 border border-gray-300", children: /* @__PURE__ */ jsx(
      Input,
      {
        type: "number",
        min: 1,
        name: "qty",
        className: "rounded-none border-none text-end focus-visible:border-none focus-visible:ring-[2px]",
        tabIndex: 1,
        ref: qtyInputRef,
        onFocus: (e) => computedItemAmount(e.currentTarget.valueAsNumber),
        onChange: (e) => computedItemAmount(e.target.valueAsNumber),
        onKeyDown: handleOnKeyDown
      }
    ) }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-36 border border-gray-300 bg-gray-50 px-1 text-end", children: /* @__PURE__ */ jsx(Label, { className: "block", children: currency((currentItem == null ? void 0 : currentItem.price) || 0) }) }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-36 border border-gray-300 px-1 text-end", children: amount > 0 ? currency(amount) : "" }),
    /* @__PURE__ */ jsx("th", { scope: "col", className: "w-6 border border-gray-300 text-end" })
  ] });
}
export {
  LinesForm as default
};
