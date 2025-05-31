import { jsxs, jsx } from "react/jsx-runtime";
import { u as useNumber } from "./use-number.js";
import { formatDate } from "date-fns";
import * as CheckboxPrimitive from "@radix-ui/react-checkbox";
import { CheckIcon } from "lucide-react";
import { c as cn } from "./utils.js";
const CurrencyCell = ({ value, columnWidth, suffix }) => {
  const { currency } = useNumber();
  let formattedValue = value !== void 0 ? currency(value) : "$0.00";
  formattedValue = suffix !== void 0 ? formattedValue.substring(1) : formattedValue;
  return /* @__PURE__ */ jsxs("div", { className: "p-1.5 text-right", title: formattedValue, style: { width: columnWidth }, children: [
    formattedValue,
    suffix && /* @__PURE__ */ jsx("span", { children: suffix })
  ] });
};
const DateCell = ({ value, columnWidth }) => {
  const formattedValue = value !== void 0 || !!value ? formatDate(new Date(value), "dd-MM-yyyy") : "";
  return /* @__PURE__ */ jsx("div", { className: "truncate p-1.5 text-left tabular-nums", title: formattedValue, style: { width: columnWidth }, children: formattedValue });
};
const HeaderCell = ({ title, columnWidth, alignment }) => {
  return /* @__PURE__ */ jsx("div", { className: `p-1.5 font-semibold text-${alignment} uppercase`, style: { width: columnWidth }, children: title });
};
const TextCell = ({ value, columnWidth }) => {
  return /* @__PURE__ */ jsx("div", { className: "p-1.5", title: value, style: { width: columnWidth }, children: value });
};
function Checkbox({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    CheckboxPrimitive.Root,
    {
      "data-slot": "checkbox",
      className: cn(
        "peer border-input dark:bg-input/30 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground dark:data-[state=checked]:bg-primary data-[state=checked]:border-primary focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive size-4 shrink-0 rounded-[4px] border shadow-xs transition-shadow outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50",
        className
      ),
      ...props,
      children: /* @__PURE__ */ jsx(
        CheckboxPrimitive.Indicator,
        {
          "data-slot": "checkbox-indicator",
          className: "flex items-center justify-center text-current transition-none",
          children: /* @__PURE__ */ jsx(CheckIcon, { className: "size-3.5" })
        }
      )
    }
  );
}
export {
  Checkbox as C,
  DateCell as D,
  HeaderCell as H,
  TextCell as T,
  CurrencyCell as a
};
