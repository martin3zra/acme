import { jsxs, jsx } from "react/jsx-runtime";
import { B as Button } from "./button.js";
import { XCircleIcon } from "lucide-react";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "./utils.js";
import "clsx";
import "tailwind-merge";
function Line({ line, index, currency, handleRemoveLine }) {
  return /* @__PURE__ */ jsxs("tr", { children: [
    /* @__PURE__ */ jsx("td", { className: "border border-gray-300 px-1 text-start", children: line.name }),
    /* @__PURE__ */ jsx("td", { className: "border border-gray-300 px-1 text-start", children: line.description }),
    /* @__PURE__ */ jsx("td", { className: "border border-gray-300 px-1 text-start", children: line.unit.name }),
    /* @__PURE__ */ jsx("td", { className: "border border-gray-300 px-1 text-end", children: line.qty }),
    /* @__PURE__ */ jsx("td", { className: "border border-gray-300 px-1 text-end", children: currency(line.price || 0) }),
    /* @__PURE__ */ jsx("td", { className: "border border-gray-300 px-1 text-end", children: currency(line.amount || 0) }),
    /* @__PURE__ */ jsx("td", { className: "border border-gray-300 px-1 text-end", children: /* @__PURE__ */ jsx(Button, { variant: "link", size: "icon", className: "h-8 w-8 rounded-full p-0", "data-index": index, onClick: handleRemoveLine, children: /* @__PURE__ */ jsx(XCircleIcon, {}) }) })
  ] });
}
export {
  Line as default
};
