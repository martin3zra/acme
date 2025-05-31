import { jsx } from "react/jsx-runtime";
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
function Index({ auth }) {
  return /* @__PURE__ */ jsx(AppLayout, { user: auth.user, children: "Error Page" });
}
export {
  Index as default
};
