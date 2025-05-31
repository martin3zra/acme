import { jsx, jsxs } from "react/jsx-runtime";
import { I as InputError } from "./input-error.js";
import { B as Button } from "./button.js";
import { D as Dialog, a as DialogContent, b as DialogHeader, c as DialogTitle, d as DialogDescription, e as DialogFooter, f as DialogClose } from "./dialog.js";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { u as useHeader } from "./use-headers.js";
import { u as useTranslation } from "./use-translation.js";
import { useForm } from "@inertiajs/react";
import { useRef } from "react";
import { A as AlertDestructive } from "./alert-destructive.js";
function useDocument() {
  return { removeElementParent };
}
function removeElementParent(e, tag) {
  var _a;
  const modal = document.getElementsByTagName(tag)[0];
  if (modal == null ? void 0 : modal.parentNode) {
    const grantFather = (_a = modal.parentNode) == null ? void 0 : _a.parentNode;
    grantFather == null ? void 0 : grantFather.removeChild(modal.parentNode);
    e.preventDefault();
  }
}
const ConfirmsPassword = ({ title, description, action, verb, path, open, onOpenChange }) => {
  const t = useTranslation().trans;
  const { removeElementParent: removeElementParent2 } = useDocument();
  const passwordInput = useRef(null);
  const {
    data,
    setData,
    delete: destroy,
    put,
    processing,
    reset,
    errors,
    clearErrors
  } = useForm({ current_password: "", status: "" });
  const { headers } = useHeader();
  const options = {
    ...headers,
    preserveScroll: true,
    onSuccess: () => handleOnOpenChange(),
    onError: () => {
      var _a;
      return (_a = passwordInput.current) == null ? void 0 : _a.focus();
    },
    onFinish: () => reset()
  };
  const onSubmit = (e) => {
    e.preventDefault();
    if (verb == "destroy") destroy(path, { ...options, preserveState: "errors" });
    if (verb == "update") put(path, { ...options, preserveState: "errors" });
  };
  const handleOnOpenChange = (open2 = false) => {
    clearErrors();
    reset();
    onOpenChange(open2);
  };
  return /* @__PURE__ */ jsx("div", { className: "space-y-6", children: /* @__PURE__ */ jsx(Dialog, { open, onOpenChange: handleOnOpenChange, children: /* @__PURE__ */ jsxs(DialogContent, { className: "sm:max-w-lg", onInteractOutside: (e) => removeElementParent2(e, "iframe"), children: [
    /* @__PURE__ */ jsxs(DialogHeader, { children: [
      /* @__PURE__ */ jsx(DialogTitle, { children: title }),
      /* @__PURE__ */ jsx(DialogDescription, { children: description })
    ] }),
    /* @__PURE__ */ jsxs("form", { className: "space-y-6", onSubmit, children: [
      errors.status && /* @__PURE__ */ jsx(AlertDestructive, { description: errors.status, onDestroy: () => delete errors.status }),
      /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "password", className: "sr-only", children: "Password" }),
        /* @__PURE__ */ jsx(
          Input,
          {
            id: "current_password",
            type: "password",
            name: "current_password",
            ref: passwordInput,
            value: data.current_password,
            onChange: (e) => setData("current_password", e.target.value),
            placeholder: t("global.password")
          }
        ),
        /* @__PURE__ */ jsx(InputError, { message: errors.current_password })
      ] }),
      /* @__PURE__ */ jsxs(DialogFooter, { children: [
        /* @__PURE__ */ jsx(DialogClose, { asChild: true, children: /* @__PURE__ */ jsx(Button, { type: "button", variant: "secondary", className: "uppercase", children: t("global.actions.cancel") }) }),
        /* @__PURE__ */ jsx(Button, { variant: "destructive", disabled: processing, asChild: true, className: "uppercase", children: /* @__PURE__ */ jsx("button", { type: "submit", children: action }) })
      ] })
    ] })
  ] }) }) });
};
function HeadingSmall({ title, description, rightPanel }) {
  return /* @__PURE__ */ jsxs("header", { className: "flex items-center justify-between", children: [
    /* @__PURE__ */ jsxs("div", { className: "space-y-1", children: [
      /* @__PURE__ */ jsx("h3", { className: "mb-0.5 text-base font-medium", children: title }),
      description && /* @__PURE__ */ jsx("p", { className: "text-muted-foreground text-sm", children: description })
    ] }),
    rightPanel && rightPanel
  ] });
}
export {
  ConfirmsPassword as C,
  HeadingSmall as H
};
