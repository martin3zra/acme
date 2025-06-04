import { jsx } from "react/jsx-runtime";
import { Link } from "@inertiajs/react";
const LinkCell = ({ href, value, columnWidth }) => {
  return /* @__PURE__ */ jsx("div", { className: "text-primary p-1.5 font-medium hover:underline hover:underline-offset-4", title: value, style: { width: columnWidth }, children: /* @__PURE__ */ jsx(Link, { href, children: value }) });
};
export {
  LinkCell as L
};
