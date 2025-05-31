import { jsxs, jsx } from "react/jsx-runtime";
import { I as InputError } from "./input-error.js";
import { B as Button } from "./button.js";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { u as useHeader } from "./use-headers.js";
import { u as useTranslation } from "./use-translation.js";
import { useForm } from "@inertiajs/react";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-label";
function CreatePassword() {
  const { headers } = useHeader();
  const t = useTranslation().trans;
  const { data, setData, errors, post, processing } = useForm({
    password: "",
    password_confirmation: ""
  });
  const submit = (e) => {
    e.preventDefault();
    post("/password", { ...headers });
  };
  return /* @__PURE__ */ jsxs("form", { onSubmit: submit, className: "space-y-6", children: [
    /* @__PURE__ */ jsx(
      "div",
      {
        className: "mb-4 text-base text-gray-600 dark:text-gray-400",
        dangerouslySetInnerHTML: { __html: t(`verify.create-password-description`).replace(/\n/g, "<br>") }
      }
    ),
    /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
      /* @__PURE__ */ jsx(Label, { htmlFor: "password", children: t(`global.password`) }),
      /* @__PURE__ */ jsx(
        Input,
        {
          id: "password",
          type: "password",
          required: true,
          autoFocus: true,
          tabIndex: 1,
          autoComplete: "password",
          value: data.password,
          onChange: (e) => setData("password", e.target.value)
        }
      ),
      /* @__PURE__ */ jsx(InputError, { message: errors.password })
    ] }),
    /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
      /* @__PURE__ */ jsx(Label, { htmlFor: "password_confirmation", children: t(`global.password_confirmation`) }),
      /* @__PURE__ */ jsx(
        Input,
        {
          id: "password_confirmation",
          type: "password",
          required: true,
          autoFocus: true,
          tabIndex: 1,
          autoComplete: "password_confirmation",
          value: data.password_confirmation,
          onChange: (e) => setData("password_confirmation", e.target.value)
        }
      ),
      /* @__PURE__ */ jsx(InputError, { message: errors.password_confirmation })
    ] }),
    /* @__PURE__ */ jsx("div", { className: "mt-4 flex items-center justify-between", children: /* @__PURE__ */ jsx(Button, { disabled: processing, children: t(`verify.create-password`) }) })
  ] });
}
export {
  CreatePassword as default
};
