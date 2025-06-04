import { jsxs, jsx } from "react/jsx-runtime";
import { u as useTranslation } from "./use-translation.js";
import { Link } from "@inertiajs/react";
import { Plus } from "lucide-react";
const AddNewPayment = () => {
  const t = useTranslation().trans;
  return /* @__PURE__ */ jsxs(
    Link,
    {
      href: "/payments/create",
      as: "button",
      className: "focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive bg-primary text-primary-foreground hover:bg-primary/90 inline-flex h-9 shrink-0 cursor-pointer items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium whitespace-nowrap shadow-xs transition-all outline-none focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 has-[>svg]:px-3 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
      children: [
        /* @__PURE__ */ jsx(Plus, {}),
        " ",
        t("payments.newPayment.title")
      ]
    }
  );
};
export {
  AddNewPayment
};
