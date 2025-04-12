import { jsxs, jsx } from "react/jsx-runtime";
import { A as AuthenticatedLayout } from "./authenticated-layout.js";
import "@tabler/icons-react";
import "react";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "lucide-react";
import "./button.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-dialog";
import "@radix-ui/react-tooltip";
import "@inertiajs/react";
import "@radix-ui/react-avatar";
import "./dropdown-menu.js";
import "@radix-ui/react-dropdown-menu";
import "@radix-ui/react-separator";
import "next-themes";
import "sonner";
function Home({ auth, flash }) {
  return /* @__PURE__ */ jsxs(AuthenticatedLayout, { user: auth.user, children: [
    /* @__PURE__ */ jsxs("div", { className: "grid auto-rows-min gap-4 md:grid-cols-3", children: [
      /* @__PURE__ */ jsx("div", { className: "bg-muted/50 aspect-video rounded-xl" }),
      /* @__PURE__ */ jsx("div", { className: "bg-muted/50 aspect-video rounded-xl" }),
      /* @__PURE__ */ jsx("div", { className: "bg-muted/50 aspect-video rounded-xl" })
    ] }),
    flash && /* @__PURE__ */ jsx("span", { children: flash.success }),
    /* @__PURE__ */ jsx("h1", { children: "Home Page" })
  ] });
}
export {
  Home as default
};
