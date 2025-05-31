import { jsxs, jsx } from "react/jsx-runtime";
import { u as useTranslation } from "./use-translation.js";
import { Link } from "@inertiajs/react";
import ConfettiBall from "./confetti-ball.js";
function Congrats() {
  const t = useTranslation().trans;
  return /* @__PURE__ */ jsxs("div", { className: "flex flex-col items-center justify-center", children: [
    /* @__PURE__ */ jsx(ConfettiBall, {}),
    /* @__PURE__ */ jsxs("div", { className: "flex max-w-md flex-col items-center justify-center gap-2 p-10", children: [
      /* @__PURE__ */ jsx("h1", { className: "text-lg font-medium", children: t("onboarding.congrats.title") }),
      /* @__PURE__ */ jsx("p", { className: "text-base font-normal", children: t("onboarding.congrats.description") }),
      /* @__PURE__ */ jsx(
        Link,
        {
          href: "/home",
          className: "focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive bg-primary text-primary-foreground hover:bg-primary/90 mt-10 inline-flex h-9 shrink-0 items-center justify-center gap-2 rounded-md px-4 py-4 text-sm font-medium whitespace-nowrap shadow-xs transition-all outline-none focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 has-[>svg]:px-3 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
          children: t("onboarding.congrats.action")
        }
      )
    ] })
  ] });
}
export {
  Congrats as default
};
