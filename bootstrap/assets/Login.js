import { jsx, jsxs } from "react/jsx-runtime";
import { I as InputError } from "./input-error.js";
import { B as Button } from "./button.js";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { u as useTranslation } from "./use-translation.js";
import { A as AppLogoIcon } from "./app-logo-icon.js";
import { Link, useForm, usePage } from "@inertiajs/react";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-label";
function AuthSimpleLayout({ children, title, description }) {
  return /* @__PURE__ */ jsx("div", { className: "bg-background flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10", children: /* @__PURE__ */ jsx("div", { className: "w-full max-w-sm", children: /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-6", children: [
    /* @__PURE__ */ jsxs("div", { className: "flex flex-col items-center gap-2", children: [
      /* @__PURE__ */ jsxs(Link, { href: "/", className: "flex flex-col items-center gap-2 font-medium", children: [
        /* @__PURE__ */ jsx("div", { className: "flex size-20 items-center justify-center rounded-md", children: /* @__PURE__ */ jsx(AppLogoIcon, { className: "fill-current text-black dark:text-white" }) }),
        /* @__PURE__ */ jsx("span", { className: "sr-only", children: title })
      ] }),
      /* @__PURE__ */ jsx("h1", { className: "text-xl font-bold", children: title }),
      /* @__PURE__ */ jsx("p", { className: "text-center text-sm", children: description })
    ] }),
    children
  ] }) }) });
}
function AuthLayout({ children, title, description, ...props }) {
  return /* @__PURE__ */ jsx(AuthSimpleLayout, { title, description, ...props, children });
}
function Login() {
  const { data, setData, post, processing, errors, reset } = useForm({
    email: "",
    password: "",
    remember: false
  });
  const props = usePage().props;
  const t = useTranslation().trans;
  const submit = (e) => {
    e.preventDefault();
    post("login", {
      headers: {
        "X-CSRF-Token": props.csrf_token
      },
      onFinish: () => reset("password")
    });
  };
  return /* @__PURE__ */ jsx(AuthLayout, { title: t("auth.login.title"), description: t("auth.login.description"), children: /* @__PURE__ */ jsx("form", { className: "flex flex-col gap-6", onSubmit: submit, children: /* @__PURE__ */ jsxs("div", { className: "grid gap-6", children: [
    /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
      /* @__PURE__ */ jsx(Label, { htmlFor: "email", children: t("global.email") }),
      /* @__PURE__ */ jsx(
        Input,
        {
          id: "email",
          type: "email",
          autoFocus: true,
          tabIndex: 1,
          autoComplete: "email",
          value: data.email,
          onChange: (e) => setData("email", e.target.value)
        }
      ),
      /* @__PURE__ */ jsx(InputError, { message: errors.email })
    ] }),
    /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
      /* @__PURE__ */ jsx("div", { className: "flex items-center", children: /* @__PURE__ */ jsx(Label, { htmlFor: "password", children: t("global.password") }) }),
      /* @__PURE__ */ jsx(
        Input,
        {
          id: "password",
          type: "password",
          tabIndex: 2,
          autoComplete: "current-password",
          value: data.password,
          onChange: (e) => setData("password", e.target.value)
        }
      ),
      /* @__PURE__ */ jsx(InputError, { message: errors.password })
    ] }),
    /* @__PURE__ */ jsx(Button, { type: "submit", className: "w-full", tabIndex: 3, disabled: processing, children: processing ? /* @__PURE__ */ jsx("div", { children: t("global.processing") }) : /* @__PURE__ */ jsx("span", { children: t("auth.login.action.login") }) })
  ] }) }) });
}
export {
  Login as default
};
