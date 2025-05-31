import { jsx } from "react/jsx-runtime";
import { I as Input } from "./input.js";
import "./utils.js";
import "clsx";
import "tailwind-merge";
const InputView = ({ value, method, autoFocus, onChange, onFocus }) => {
  const handleChange = (event) => {
    onChange(method, event.currentTarget.valueAsNumber);
  };
  return /* @__PURE__ */ jsx("div", { className: "p-0", children: /* @__PURE__ */ jsx(
    Input,
    {
      type: "number",
      min: 0,
      className: `${method !== "cash" && "cursor-pointer"} border-none text-end`,
      value,
      autoFocus,
      onFocus: () => onFocus(method),
      onChange: handleChange,
      readOnly: method !== "cash"
    },
    method
  ) });
};
export {
  InputView
};
