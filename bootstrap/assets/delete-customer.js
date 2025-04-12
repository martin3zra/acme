import { jsx, jsxs } from "react/jsx-runtime";
import { L as Label, I as InputError } from "./label.js";
import { B as Button } from "./button.js";
import { u as useHeader, D as Dialog, a as DialogContent, b as useDocument, c as DialogHeader, d as DialogTitle, e as DialogDescription, f as DialogFooter, g as DialogClose } from "./use-headers.js";
import { I as Input } from "./input.js";
import { useForm } from "@inertiajs/react";
import { useRef } from "react";
import "@radix-ui/react-label";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-dialog";
import "lucide-react";
const DeleteCustomer = ({ customer, open, onOpenChange }) => {
  const { removeElementParent } = useDocument();
  const passwordInput = useRef(null);
  const { data, setData, delete: destroy, processing, reset, errors, clearErrors } = useForm({ password: "" });
  const { headers } = useHeader();
  const deleteCustomer = (e) => {
    e.preventDefault();
    destroy(`/customers/${customer.id}`, {
      ...headers,
      preserveScroll: true,
      preserveState: false,
      onSuccess: () => handleOnOpenChange(),
      onError: () => {
        var _a;
        return (_a = passwordInput.current) == null ? void 0 : _a.focus();
      },
      onFinish: () => reset()
    });
  };
  const handleOnOpenChange = (open2 = false) => {
    clearErrors();
    reset();
    onOpenChange(open2);
  };
  return /* @__PURE__ */ jsx("div", { className: "space-y-6", children: /* @__PURE__ */ jsx(Dialog, { open, onOpenChange: handleOnOpenChange, children: /* @__PURE__ */ jsxs(DialogContent, { className: "sm:max-w-lg", onInteractOutside: (e) => removeElementParent(e, "iframe"), children: [
    /* @__PURE__ */ jsxs(DialogHeader, { children: [
      /* @__PURE__ */ jsxs(DialogTitle, { children: [
        "Are you sure you want to delete ",
        customer == null ? void 0 : customer.name,
        "?"
      ] }),
      /* @__PURE__ */ jsxs(DialogDescription, { children: [
        "Once ",
        customer == null ? void 0 : customer.name,
        " is deleted, all of its resources will continue to be available, but no new operation can be performed."
      ] })
    ] }),
    /* @__PURE__ */ jsxs("form", { className: "space-y-6", onSubmit: deleteCustomer, children: [
      /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "password", className: "sr-only", children: "Password" }),
        /* @__PURE__ */ jsx(
          Input,
          {
            id: "password",
            type: "password",
            name: "password",
            ref: passwordInput,
            value: data.password,
            onChange: (e) => setData("password", e.target.value),
            placeholder: "Password",
            autoComplete: "current-password"
          }
        ),
        /* @__PURE__ */ jsx(InputError, { message: errors.password })
      ] }),
      /* @__PURE__ */ jsxs(DialogFooter, { children: [
        /* @__PURE__ */ jsx(DialogClose, { asChild: true, children: /* @__PURE__ */ jsx(Button, { type: "button", variant: "secondary", children: "Close" }) }),
        /* @__PURE__ */ jsx(Button, { variant: "destructive", disabled: processing, asChild: true, children: /* @__PURE__ */ jsx("button", { type: "submit", children: "Delete customer" }) })
      ] })
    ] })
  ] }) }) });
};
export {
  DeleteCustomer
};
