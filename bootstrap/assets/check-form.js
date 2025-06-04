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
const CheckFormView = ({ amount, reference, onChange }) => {
  const t = useTranslation().trans;
  const handleChange = (event) => {
    if (event.currentTarget.name === "ck") {
      onChange(event.currentTarget.valueAsNumber);
      return;
    }
    onChange(event.currentTarget.value);
  };
  return /* @__PURE__ */ jsx("div", { children: /* @__PURE__ */ jsxs(FormSection, { onSubmit: () => {
  }, children: [
    /* @__PURE__ */ jsx(FormSection.Title, { children: t("global.paymentMethods.ck.form.title") }),
    /* @__PURE__ */ jsx(FormSection.Description, { children: t("global.paymentMethods.ck.form.description") }),
    /* @__PURE__ */ jsxs(FormSection.Form, { children: [
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-4", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "ck", className: "text-end", children: t("global.amount") }),
        /* @__PURE__ */ jsx(Input, { type: "number", min: 0, name: "ck", className: "h-12 text-end md:text-xl", onChange: handleChange, autoFocus: true, value: amount })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-4", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "ck", children: t("global.paymentMethods.ck.form.reference") }),
        /* @__PURE__ */ jsx(Input, { type: "text", name: "reference", className: "h-12 text-start md:text-xl", onChange: handleChange, value: reference })
      ] })
    ] })
  ] }) });
};
export {
  CheckFormView
};
