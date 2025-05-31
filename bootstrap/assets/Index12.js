import { jsxs, jsx, Fragment } from "react/jsx-runtime";
import { I as InputError } from "./input-error.js";
import { B as Button } from "./button.js";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { u as useHeader } from "./use-headers.js";
import { u as useTranslation } from "./use-translation.js";
import { A as AppLogoIcon } from "./app-logo-icon.js";
import { Link, useForm, Head } from "@inertiajs/react";
import CreatePassword from "./create-password.js";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-label";
function Guest({ children }) {
  return /* @__PURE__ */ jsxs("div", { className: "flex min-h-screen flex-col items-center bg-gray-100 pt-6 sm:justify-center sm:pt-0 dark:bg-gray-900", children: [
    /* @__PURE__ */ jsx("div", { children: /* @__PURE__ */ jsx(Link, { href: "/", children: /* @__PURE__ */ jsx(AppLogoIcon, { className: "h-20 w-20 fill-current text-gray-500" }) }) }),
    /* @__PURE__ */ jsx("div", { className: "mt-0 w-full overflow-hidden bg-white px-6 py-4 shadow-md sm:max-w-md sm:rounded-lg dark:bg-gray-800", children })
  ] });
}
function Verify({ status, csrf_token }) {
  const { headers } = useHeader();
  const t = useTranslation().trans;
  const { data, setData, errors, post, processing } = useForm({
    email: ""
  });
  const showReSendEmailForm = status !== "verification-link-sent" && status !== "account-verified" && status !== "already-verified" && status !== "create-password";
  const submit = (e) => {
    e.preventDefault();
    post("/email/verification-notification", { ...headers });
  };
  return /* @__PURE__ */ jsxs(Guest, { children: [
    /* @__PURE__ */ jsx(Head, { title: "Email Verification" }),
    status !== "verification-link-sent" && status !== "create-password" && /* @__PURE__ */ jsx(
      "div",
      {
        className: "mb-4 text-base text-gray-600 dark:text-gray-400",
        dangerouslySetInnerHTML: { __html: t(`verify.${status}`).replace(/\n/g, "<br>") }
      }
    ),
    status === "already-verified" && /* @__PURE__ */ jsxs(Link, { href: "/home", children: [
      t("global.visit", { to: "tu" }),
      t("global.navMain.dashboard")
    ] }),
    status === "verification-link-sent" && /* @__PURE__ */ jsx("div", { className: "mb-4 text-base font-medium text-green-600 dark:text-green-400", children: t(`verify.${status}`) }),
    status === "account-verified" && /* @__PURE__ */ jsxs(Fragment, { children: [
      /* @__PURE__ */ jsx(Link, { href: "/login", children: t("verify.login") }),
      /* @__PURE__ */ jsx(
        Link,
        {
          href: "/logout",
          method: "post",
          headers: { "X-CSRF-Token": csrf_token },
          as: "button",
          className: "rounded-md text-sm text-gray-600 underline hover:text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:outline-none",
          children: "Log Out"
        }
      )
    ] }),
    status === "create-password" && /* @__PURE__ */ jsx(CreatePassword, {}),
    showReSendEmailForm && /* @__PURE__ */ jsxs("form", { onSubmit: submit, children: [
      /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "email", children: t(`global.email`) }),
        /* @__PURE__ */ jsx(
          Input,
          {
            id: "email",
            type: "email",
            required: true,
            autoFocus: true,
            tabIndex: 1,
            autoComplete: "email",
            value: data.email,
            onChange: (e) => setData("email", e.target.value)
          }
        ),
        /* @__PURE__ */ jsx(InputError, { message: errors.email })
      ] }),
      /* @__PURE__ */ jsx("div", { className: "mt-4 flex items-center justify-between", children: /* @__PURE__ */ jsx(Button, { disabled: processing, children: t(`verify.resend-verification-email`) }) })
    ] })
  ] });
}
export {
  Verify as default
};
