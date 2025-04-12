import { jsx } from "react/jsx-runtime";
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
function Index({ auth }) {
  return /* @__PURE__ */ jsx(AuthenticatedLayout, { user: auth.user, children: "Error Page" });
}
export {
  Index as default
};
