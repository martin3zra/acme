import { jsxs, jsx } from "react/jsx-runtime";
import { L as Label, I as InputError } from "./label.js";
import { B as Button } from "./button.js";
import { u as useHeader, D as Dialog, a as DialogContent, b as useDocument, c as DialogHeader, d as DialogTitle, e as DialogDescription, f as DialogFooter, g as DialogClose } from "./use-headers.js";
import { I as Input } from "./input.js";
import { useForm, usePage } from "@inertiajs/react";
import { useRef, useState } from "react";
function HeadingSmall({ title, description, rightPanel }) {
  return /* @__PURE__ */ jsxs("header", { className: "flex items-center justify-between", children: [
    /* @__PURE__ */ jsxs("div", { className: "space-y-1", children: [
      /* @__PURE__ */ jsx("h3", { className: "mb-0.5 text-base font-medium", children: title }),
      description && /* @__PURE__ */ jsx("p", { className: "text-muted-foreground text-sm", children: description })
    ] }),
    rightPanel && rightPanel
  ] });
}
function useVerb() {
  return { action };
}
function action(verb) {
  return {
    view: "View",
    edit: "Update",
    trash: "Trash",
    create: "Create"
  }[verb];
}
const SecureActionDialog = ({ title, description, action: action2, verb, path, open, onOpenChange }) => {
  const { removeElementParent } = useDocument();
  const passwordInput = useRef(null);
  const { data, setData, delete: destroy, put, processing, reset, errors, clearErrors } = useForm({ password: "" });
  const { headers } = useHeader();
  const options = {
    ...headers,
    preserveScroll: true,
    preserveState: false,
    onSuccess: () => handleOnOpenChange(),
    onError: () => {
      var _a;
      return (_a = passwordInput.current) == null ? void 0 : _a.focus();
    },
    onFinish: () => reset()
  };
  const onSubmit = (e) => {
    e.preventDefault();
    if (verb == "destroy") destroy(path, options);
    if (verb == "update") put(path, options);
  };
  const handleOnOpenChange = (open2 = false) => {
    clearErrors();
    reset();
    onOpenChange(open2);
  };
  return /* @__PURE__ */ jsx("div", { className: "space-y-6", children: /* @__PURE__ */ jsx(Dialog, { open, onOpenChange: handleOnOpenChange, children: /* @__PURE__ */ jsxs(DialogContent, { className: "sm:max-w-lg", onInteractOutside: (e) => removeElementParent(e, "iframe"), children: [
    /* @__PURE__ */ jsxs(DialogHeader, { children: [
      /* @__PURE__ */ jsx(DialogTitle, { children: title }),
      /* @__PURE__ */ jsx(DialogDescription, { children: description })
    ] }),
    /* @__PURE__ */ jsxs("form", { className: "space-y-6", onSubmit, children: [
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
        /* @__PURE__ */ jsx(Button, { variant: "destructive", disabled: processing, asChild: true, children: /* @__PURE__ */ jsx("button", { type: "submit", children: action2 }) })
      ] })
    ] })
  ] }) }) });
};
function CreateForm({ onFinish, params }) {
  var _a, _b, _c, _d, _e, _f, _g, _h, _i;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage().props;
  const { data, setData, post, put, errors, reset, processing } = useForm({
    name: ((_a = params.customer) == null ? void 0 : _a.name) || "",
    contact: ((_b = params.customer) == null ? void 0 : _b.contact_name) || "",
    email: ((_c = params.customer) == null ? void 0 : _c.email) || "",
    phone: ((_d = params.customer) == null ? void 0 : _d.phone) || ""
  });
  const viewMode = params.action === "view";
  const isDisabled = ((_e = params.customer) == null ? void 0 : _e.status) === "disabled";
  const options = {
    ...headers,
    preserveScroll: true,
    onSuccess: () => {
      reset();
      onFinish();
    }
  };
  const verbName = useVerb().action(params.action);
  const submit = (e) => {
    e.preventDefault();
    if (params.action === "create") post("/customers", options);
    if (params.action === "edit") put(`/customers/${params.customer.id}`, options);
  };
  const dialogProps = {
    enabled: {
      title: `Are you sure you want to disable ${(_f = params.customer) == null ? void 0 : _f.name}?`,
      description: "Once this customer is disabled, all of its resources and data will also be lock. Please enter your password to confirm you would like to disabled this customer.",
      action: "Disabled",
      variant: "destructive"
    },
    disabled: {
      title: `Are you sure you want to enable ${(_g = params.customer) == null ? void 0 : _g.name}?`,
      description: "Once this customer is enable, all of its resources and data will also be unlock. Please enter your password to confirm you would like to enabled this customer.",
      action: "Enabled",
      variant: "primary"
    }
  }[((_h = params.customer) == null ? void 0 : _h.status) || "enabled"];
  return /* @__PURE__ */ jsxs("div", { children: [
    propsErrors.status && /* @__PURE__ */ jsx("div", { className: "mb-4 text-center text-sm font-medium text-red-600", children: propsErrors.status }),
    /* @__PURE__ */ jsxs("form", { onSubmit: submit, className: "space-y-6", children: [
      /* @__PURE__ */ jsxs("div", { className: "grid grid-cols-2 gap-2", children: [
        /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
          /* @__PURE__ */ jsx(Label, { htmlFor: "name", children: "Name" }),
          /* @__PURE__ */ jsx(
            Input,
            {
              id: "name",
              className: "mt-1 block w-full",
              value: data.name,
              onChange: (e) => setData("name", e.target.value),
              required: true,
              autoComplete: "name",
              placeholder: "Full name",
              readOnly: viewMode
            }
          ),
          /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.name })
        ] }),
        /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
          /* @__PURE__ */ jsx(Label, { htmlFor: "contact", children: "Contact" }),
          /* @__PURE__ */ jsx(
            Input,
            {
              id: "name",
              className: "mt-1 block w-full",
              value: data.contact,
              onChange: (e) => setData("contact", e.target.value),
              placeholder: "Jane Doe",
              readOnly: viewMode
            }
          ),
          /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.contact })
        ] })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "grid grid-cols-2 gap-2", children: [
        /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
          /* @__PURE__ */ jsx(Label, { htmlFor: "email", children: "Email" }),
          /* @__PURE__ */ jsx(
            Input,
            {
              id: "email",
              type: "email",
              className: "mt-1 block w-full",
              value: data.email,
              onChange: (e) => setData("email", e.target.value),
              required: true,
              placeholder: "jane.doe@example.com",
              readOnly: viewMode
            }
          ),
          /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.email })
        ] }),
        /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
          /* @__PURE__ */ jsx(Label, { htmlFor: "phone", children: "Phone" }),
          /* @__PURE__ */ jsx(
            Input,
            {
              id: "phone",
              className: "mt-1 block w-full",
              value: data.phone,
              onChange: (e) => setData("phone", e.target.value),
              placeholder: "809-983-3897",
              readOnly: viewMode
            }
          ),
          /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.phone })
        ] })
      ] }),
      !viewMode && /* @__PURE__ */ jsx("div", { className: "flex items-center gap-4", children: /* @__PURE__ */ jsxs(Button, { disabled: processing, children: [
        verbName,
        " Customer"
      ] }) })
    ] }),
    viewMode && /* @__PURE__ */ jsxs("div", { className: "space-y-6 pt-12", children: [
      /* @__PURE__ */ jsx(
        HeadingSmall,
        {
          title: `${dialogProps == null ? void 0 : dialogProps.action} customer`,
          description: `${isDisabled ? "Unlock" : "Lock"} customer and all of its resources`
        }
      ),
      /* @__PURE__ */ jsxs("div", { className: `space-y-4 rounded-lg border ${isDisabled ? "border-primary-100 bg-primary-50" : "border-red-100 bg-red-50"} p-4`, children: [
        /* @__PURE__ */ jsxs("div", { className: `relative space-y-0.5 ${isDisabled ? "text-primary" : "text-red-600"}`, children: [
          /* @__PURE__ */ jsx("p", { className: "font-medium", children: "Warning" }),
          /* @__PURE__ */ jsx("p", { className: "text-sm", children: "Please proceed with caution, this is not permanent." })
        ] }),
        /* @__PURE__ */ jsx(Button, { variant: isDisabled ? "default" : "destructive", onClick: () => setDialogOpen(true), children: `${dialogProps == null ? void 0 : dialogProps.action} customer` }),
        /* @__PURE__ */ jsx(
          SecureActionDialog,
          {
            title: (dialogProps == null ? void 0 : dialogProps.title) || "",
            description: (dialogProps == null ? void 0 : dialogProps.description) || "",
            action: `${dialogProps == null ? void 0 : dialogProps.action} it`,
            verb: "update",
            path: `/customers/${(_i = params.customer) == null ? void 0 : _i.id}/change-status`,
            open: dialogOpen,
            onOpenChange: setDialogOpen
          }
        )
      ] })
    ] })
  ] });
}
const CreateForm$1 = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  default: CreateForm
}, Symbol.toStringTag, { value: "Module" }));
export {
  CreateForm as C,
  HeadingSmall as H,
  CreateForm$1 as a,
  useVerb as u
};
