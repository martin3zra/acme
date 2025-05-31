import { jsx, jsxs } from "react/jsx-runtime";
import { F as FormSection } from "./form-section.js";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { u as useTranslation } from "./use-translation.js";
import "react";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-label";
import "@inertiajs/react";
const BankTransferFormView = ({ amount, reference, onChange }) => {
  const t = useTranslation().trans;
  const handleChange = (event) => {
    if (event.currentTarget.name === "amount") {
      onChange(event.currentTarget.valueAsNumber);
      return;
    }
    onChange(event.currentTarget.value);
  };
  return /* @__PURE__ */ jsx("div", { children: /* @__PURE__ */ jsxs(FormSection, { onSubmit: () => {
  }, children: [
    /* @__PURE__ */ jsx(FormSection.Title, { children: t("global.paymentMethods.bt.form.title") }),
    /* @__PURE__ */ jsx(FormSection.Description, { children: t("global.paymentMethods.bt.form.description") }),
    /* @__PURE__ */ jsxs(FormSection.Form, { children: [
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-4", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "amount", className: "text-end", children: t("global.amount") }),
        /* @__PURE__ */ jsx(Input, { type: "number", min: 0, name: "amount", className: "h-12 text-end md:text-xl", onChange: handleChange, autoFocus: true, value: amount })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-4", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "reference", children: t("global.paymentMethods.bt.form.reference") }),
        /* @__PURE__ */ jsx(Input, { type: "text", name: "reference", className: "h-12 text-start md:text-xl", onChange: handleChange, value: reference })
      ] })
    ] })
  ] }) });
};
export {
  BankTransferFormView
};
