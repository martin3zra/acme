import { jsx, jsxs } from "react/jsx-runtime";
import { L as Label, I as InputError } from "./label.js";
import { B as Button } from "./button.js";
import { I as Input } from "./input.js";
import { Link, useForm, usePage } from "@inertiajs/react";
import "@radix-ui/react-label";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "clsx";
import "tailwind-merge";
function AppLogoIcon(props) {
  return /* @__PURE__ */ jsx("svg", { ...props, xmlns: "http://www.w3.org/2000/svg", fill: "none", viewBox: "0 0 24 24", strokeWidth: "1.5", stroke: "currentColor", className: "size-6", children: /* @__PURE__ */ jsx(
    "path",
    {
      strokeLinecap: "round",
      strokeLinejoin: "round",
      d: "M10.5 6h9.75M10.5 6a1.5 1.5 0 1 1-3 0m3 0a1.5 1.5 0 1 0-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-9.75 0h9.75"
    }
  ) });
}
function AuthSimpleLayout({ children, title, description }) {
  return /* @__PURE__ */ jsx("div", { className: "bg-background flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10", children: /* @__PURE__ */ jsx("div", { className: "w-full max-w-sm", children: /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-6", children: [
    /* @__PURE__ */ jsxs("div", { className: "flex flex-col items-center gap-2", children: [
      /* @__PURE__ */ jsxs(Link, { href: "/", className: "flex flex-col items-center gap-2 font-medium", children: [
        /* @__PURE__ */ jsx("div", { className: "flex h-10 w-10 items-center justify-center rounded-md", children: /* @__PURE__ */ jsx(AppLogoIcon, { className: "size-10 fill-current text-black dark:text-white" }) }),
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
    email: "martin3zra@gmail.com",
    password: "AmAr$.@z",
    remember: false
  });
  const props = usePage().props;
  const submit = (e) => {
    e.preventDefault();
    post("login", {
      headers: {
        "X-CSRF-Token": props.csrf_token
      },
      onFinish: () => reset("password")
    });
  };
  return /* @__PURE__ */ jsx(AuthLayout, { title: "Log in to your account", description: "Enter your email and password below to log in", children: /* @__PURE__ */ jsx("form", { className: "flex flex-col gap-6", onSubmit: submit, children: /* @__PURE__ */ jsxs("div", { className: "grid gap-6", children: [
    /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
      /* @__PURE__ */ jsx(Label, { htmlFor: "email", children: "Email Address" }),
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
      /* @__PURE__ */ jsx("div", { className: "flex items-center", children: /* @__PURE__ */ jsx(Label, { htmlFor: "password", children: "Password" }) }),
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
    /* @__PURE__ */ jsx(Button, { type: "submit", className: "w-full", tabIndex: 3, disabled: processing, children: processing ? /* @__PURE__ */ jsx("div", { children: "Processing..." }) : /* @__PURE__ */ jsx("span", { children: "Log In" }) })
  ] }) }) });
}
export {
  Login as default
};
