import { jsx, jsxs } from "react/jsx-runtime";
import { F as FormSection } from "./form-section.js";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { S as Select, a as SelectTrigger, b as SelectValue, c as SelectContent, d as SelectItem } from "./select.js";
import { f as defaultCardBrands } from "./constants2.js";
import { u as useTranslation } from "./use-translation.js";
import "react";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-label";
import "@radix-ui/react-select";
import "lucide-react";
import "@inertiajs/react";
const CardFormView = ({ last4, brand, amount, reference, onChange }) => {
  const t = useTranslation().trans;
  const handleChange = (event) => {
    if (event.currentTarget.name === "last4") {
      event.currentTarget.value = event.currentTarget.value.replace(/\D/g, "");
      if (event.currentTarget.value.length > event.currentTarget.maxLength) {
        event.currentTarget.value = event.currentTarget.value.slice(0, event.currentTarget.maxLength);
      }
      onChange(event.currentTarget.valueAsNumber, event.currentTarget.name);
      return;
    }
    if (event.currentTarget.name === "amount") {
      onChange(event.currentTarget.valueAsNumber, "amount");
    }
    onChange(event.currentTarget.value, event.currentTarget.name);
  };
  return /* @__PURE__ */ jsx("div", { children: /* @__PURE__ */ jsxs(FormSection, { onSubmit: () => {
  }, children: [
    /* @__PURE__ */ jsx(FormSection.Title, { children: t("global.paymentMethods.card.form.title") }),
    /* @__PURE__ */ jsx(FormSection.Description, { children: t("global.paymentMethods.card.form.description") }),
    /* @__PURE__ */ jsxs(FormSection.Form, { children: [
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-3", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "last4", className: "text-end", children: t("global.paymentMethods.card.form.last4") }),
        /* @__PURE__ */ jsx(
          Input,
          {
            type: "number",
            inputMode: "numeric",
            name: "last4",
            pattern: "[0-9]*",
            maxLength: 4,
            className: "h-12 text-end md:text-xl",
            onChange: handleChange,
            autoFocus: true,
            value: last4
          }
        )
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-3", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "brand", className: "text-end", children: t("global.paymentMethods.card.form.brand") }),
        /* @__PURE__ */ jsxs(Select, { name: "brand", onValueChange: (value) => onChange(value, "brand"), value: brand, required: true, children: [
          /* @__PURE__ */ jsx(SelectTrigger, { className: "w-full", size: "lg", children: /* @__PURE__ */ jsx(SelectValue, { placeholder: "Select brand" }) }),
          /* @__PURE__ */ jsx(SelectContent, { className: "w-full", children: defaultCardBrands.map((brand2, index) => /* @__PURE__ */ jsx(SelectItem, { value: brand2.value.toString(), children: brand2.name }, index.toString())) })
        ] })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-3", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "reference", children: t("global.paymentMethods.card.form.authorization") }),
        /* @__PURE__ */ jsx(Input, { type: "text", name: "reference", className: "h-12 text-start md:text-xl", onChange: handleChange, value: reference })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-3", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "amount", className: "text-end", children: t("global.amount") }),
        /* @__PURE__ */ jsx(
          Input,
          {
            type: "number",
            inputMode: "numeric",
            name: "amount",
            pattern: "[0-9]*",
            className: "h-12 text-end md:text-xl",
            onChange: handleChange,
            value: amount
          }
        )
      ] })
    ] })
  ] }) });
};
export {
  CardFormView
};
