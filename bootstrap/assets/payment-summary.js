import { jsxs, jsx } from "react/jsx-runtime";
import { u as useNumber } from "./use-number.js";
import { u as useTranslation } from "./use-translation.js";
import "@inertiajs/react";
const categorizePayments = (data) => {
  const methods = ["cash", "ck", "card", "bt"];
  const usedMethods = methods.map((method) => {
    const entry = data[method];
    if (entry.amount > 0) {
      const categorized = {
        method,
        amount: entry.amount,
        reference: entry.reference || void 0
      };
      if (method === "card") {
        categorized.additionalInfo = {
          brand: entry.brand,
          last4: entry.last4
        };
      }
      return categorized;
    }
    return null;
  }).filter((item) => item !== null);
  const primaryMethod = usedMethods.reduce((prev, current) => {
    return !prev || current.amount > prev.amount ? current : prev;
  }, null);
  return {
    usedMethods,
    primaryMethod
  };
};
const PaymentSummary = ({ paymentData }) => {
  const t = useTranslation().trans;
  const { usedMethods, primaryMethod } = categorizePayments(paymentData);
  const { currency } = useNumber();
  return /* @__PURE__ */ jsxs("div", { className: "mx-auto w-full rounded-md bg-white", children: [
    /* @__PURE__ */ jsx("ul", { className: "space-y-1", children: usedMethods.map((method, index) => /* @__PURE__ */ jsxs("li", { className: "rounded-md border border-gray-100 bg-gray-50 p-3 text-gray-700", children: [
      /* @__PURE__ */ jsx("div", { className: "font-medium text-indigo-600 uppercase", children: t(`global.paymentMethods.${method.method}.title`).toUpperCase() }),
      /* @__PURE__ */ jsxs("div", { children: [
        /* @__PURE__ */ jsxs("span", { className: "font-semibold", children: [
          "$",
          method.amount.toFixed(2)
        ] }),
        method.reference && /* @__PURE__ */ jsxs("span", { className: "ml-2 text-sm text-gray-500", children: [
          "(Ref: ",
          method.reference,
          ")"
        ] })
      ] }),
      method.additionalInfo && /* @__PURE__ */ jsxs("div", { className: "text-sm text-gray-500", children: [
        method.additionalInfo.brand && /* @__PURE__ */ jsxs("span", { children: [
          "Brand: ",
          method.additionalInfo.brand,
          ", "
        ] }),
        method.additionalInfo.last4 && /* @__PURE__ */ jsxs("span", { children: [
          "Last4: ",
          method.additionalInfo.last4
        ] })
      ] })
    ] }, index)) }),
    primaryMethod && /* @__PURE__ */ jsx("div", { className: "mt-2 rounded-md border border-indigo-200 bg-indigo-50 p-4", children: /* @__PURE__ */ jsxs("p", { className: "font-semibold text-indigo-700", children: [
      t("global.paymentMethods.primary"),
      ": ",
      t(`global.paymentMethods.${primaryMethod.method}.title`).toUpperCase(),
      /* @__PURE__ */ jsx("span", { className: "block", children: currency(primaryMethod.amount) })
    ] }) })
  ] });
};
export {
  PaymentSummary as default
};
