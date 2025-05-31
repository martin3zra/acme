import { jsx, jsxs } from "react/jsx-runtime";
import { A as AppLogoIcon } from "./app-logo-icon.js";
import { u as useHeader } from "./use-headers.js";
import { u as useTranslation } from "./use-translation.js";
import { useForm, Link } from "@inertiajs/react";
import { F as FormSection } from "./form-section.js";
import { I as InputError } from "./input-error.js";
import { B as Button } from "./button.js";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { S as Separator } from "./separator.js";
import Congrats from "./congrats.js";
import "react";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-label";
import "@radix-ui/react-separator";
import "./confetti-ball.js";
function CreateCompanyForm() {
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const { data, setData, errors, processing, post, reset } = useForm({
    name: "",
    rnc: "",
    city: "",
    address: ""
  });
  const handleChange = (event) => {
    setData(event.target.name, event.target.value);
  };
  const submit = () => {
    post("/companies", { ...headers, onFinish: () => reset() });
  };
  return /* @__PURE__ */ jsx("div", { children: /* @__PURE__ */ jsxs(FormSection, { onSubmit: submit, children: [
    /* @__PURE__ */ jsx(FormSection.Title, { children: t("onboarding.company.title") }),
    /* @__PURE__ */ jsx(FormSection.Description, { children: t("onboarding.company.description") }),
    /* @__PURE__ */ jsxs(FormSection.Form, { children: [
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-4", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "name", className: "text-end", children: t("onboarding.company.name") }),
        /* @__PURE__ */ jsx(Input, { type: "text", name: "name", className: "h-12 md:text-xl", onChange: handleChange, value: data.name, autoFocus: true }),
        /* @__PURE__ */ jsx(InputError, { message: errors.name })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "col-span-3 space-y-2 sm:col-span-3", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "rnc", children: t("onboarding.company.rnc") }),
        /* @__PURE__ */ jsx(Input, { type: "text", name: "rnc", maxLength: 11, className: "h-12 text-start md:text-xl", value: data.rnc, onChange: handleChange }),
        /* @__PURE__ */ jsx(InputError, { message: errors.rnc })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-4", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "city", children: t("onboarding.company.city") }),
        /* @__PURE__ */ jsx(Input, { type: "text", name: "city", className: "h-12 text-start md:text-xl", value: data.city, onChange: handleChange }),
        /* @__PURE__ */ jsx(InputError, { message: errors.city })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "col-span-6 space-y-2 sm:col-span-4", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "address", children: t("onboarding.company.address") }),
        /* @__PURE__ */ jsx(Input, { type: "text", name: "address", className: "h-12 text-start md:text-xl", value: data.address, onChange: handleChange }),
        /* @__PURE__ */ jsx(InputError, { message: errors.address })
      ] })
    ] }),
    /* @__PURE__ */ jsx(FormSection.Actions, { children: /* @__PURE__ */ jsx(Button, { type: "submit", disabled: processing, className: "h-12 md:text-xl", children: t("onboarding.company.action") }) })
  ] }) });
}
function AppSimpleLayout({ children }) {
  return /* @__PURE__ */ jsx("div", { className: "bg-background flex min-h-svh", children: /* @__PURE__ */ jsx("div", { className: "w-full", children }) });
}
function Index({ csrf_token, status }) {
  const t = useTranslation().trans;
  return /* @__PURE__ */ jsx(AppSimpleLayout, { children: /* @__PURE__ */ jsxs("div", { className: "grid grid-cols-12 p-10", children: [
    /* @__PURE__ */ jsxs("div", { className: "col-span-3 flex flex-col gap-4", children: [
      /* @__PURE__ */ jsx("div", { className: "size-12", children: /* @__PURE__ */ jsx(AppLogoIcon, { className: "size-4 fill-current text-white dark:text-black" }) }),
      /* @__PURE__ */ jsx("h1", { className: "text-4xl font-normal", children: t("onboarding.title") }),
      /* @__PURE__ */ jsx("p", { className: "text-base font-normal", children: t("onboarding.description") }),
      /* @__PURE__ */ jsx(Separator, { className: "max-w-xs" }),
      /* @__PURE__ */ jsx(
        Link,
        {
          href: "/logout",
          method: "post",
          headers: { "X-CSRF-Token": csrf_token },
          as: "button",
          className: "cursor-pointer text-start text-sm text-gray-600 underline hover:text-gray-900",
          children: t("global.navUser.logout")
        }
      )
    ] }),
    /* @__PURE__ */ jsx("div", { className: "col-span-9 py-16", children: status === "success" ? /* @__PURE__ */ jsx(Congrats, {}) : /* @__PURE__ */ jsx(CreateCompanyForm, {}) })
  ] }) });
}
export {
  Index as default
};
