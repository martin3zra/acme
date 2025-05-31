import { jsx, jsxs } from "react/jsx-runtime";
import { A as AppLayout } from "./app-layout.js";
import "react";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "lucide-react";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "./button.js";
import "./sheet.js";
import "@radix-ui/react-dialog";
import "./tooltip.js";
import "@radix-ui/react-tooltip";
import "./use-translation.js";
import "@inertiajs/react";
import "@radix-ui/react-avatar";
import "./dropdown-menu.js";
import "@radix-ui/react-dropdown-menu";
import "@tabler/icons-react";
import "./app-logo-icon.js";
import "./separator.js";
import "@radix-ui/react-separator";
import "next-themes";
import "sonner";
function Home({ auth, flash }) {
  return /* @__PURE__ */ jsx(AppLayout, { user: auth.user, children: /* @__PURE__ */ jsxs("div", { children: [
    /* @__PURE__ */ jsxs("div", { className: "grid auto-rows-min gap-4 md:grid-cols-3", children: [
      /* @__PURE__ */ jsx("div", { className: "bg-muted/50 aspect-video rounded-xl" }),
      /* @__PURE__ */ jsx("div", { className: "bg-muted/50 aspect-video rounded-xl" }),
      /* @__PURE__ */ jsx("div", { className: "bg-muted/50 aspect-video rounded-xl" })
    ] }),
    flash && /* @__PURE__ */ jsx("span", { children: flash.success }),
    /* @__PURE__ */ jsx("h1", { children: "Home Page" })
  ] }) });
}
export {
  Home as default
};
