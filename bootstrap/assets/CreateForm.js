import { jsxs, jsx } from "react/jsx-runtime";
import { H as HeadingSmall, C as ConfirmsPassword } from "./heading-small.js";
import { I as InputError } from "./input-error.js";
import { B as Button } from "./button.js";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { u as useHeader } from "./use-headers.js";
import { u as useVerb } from "./use-verbs.js";
import { u as useTranslation } from "./use-translation.js";
import { usePage, useForm } from "@inertiajs/react";
import { useState } from "react";
import "./dialog.js";
import "@radix-ui/react-dialog";
import "lucide-react";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "./alert-destructive.js";
import "class-variance-authority";
import "@radix-ui/react-slot";
import "@radix-ui/react-label";
function CreateForm({ onFinish, params }) {
  var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l, _m, _n;
  const t = useTranslation().trans;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage().props;
  const { data, setData, post, put, errors, reset, processing } = useForm({
    id: (_a = params.customer) == null ? void 0 : _a.id,
    name: ((_b = params.customer) == null ? void 0 : _b.name) || "",
    contact: ((_c = params.customer) == null ? void 0 : _c.contact_name) || "",
    email: ((_d = params.customer) == null ? void 0 : _d.email) || "",
    phone: ((_e = params.customer) == null ? void 0 : _e.phone) || ""
  });
  const viewMode = params.action === "view";
  const isDisabled = ((_f = params.customer) == null ? void 0 : _f.status) === "disabled";
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
  return /* @__PURE__ */ jsxs("div", { children: [
    propsErrors.status && /* @__PURE__ */ jsx("div", { className: "mb-4 text-center text-sm font-medium text-red-600", children: propsErrors.status }),
    /* @__PURE__ */ jsxs("form", { onSubmit: submit, className: "space-y-6", children: [
      /* @__PURE__ */ jsxs("div", { className: "grid grid-cols-2 gap-2", children: [
        /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
          /* @__PURE__ */ jsx(Label, { htmlFor: "name", children: t("global.name") }),
          /* @__PURE__ */ jsx(
            Input,
            {
              id: "name",
              className: "mt-1 block w-full",
              value: data.name,
              onChange: (e) => setData("name", e.target.value),
              required: true,
              autoComplete: "name",
              placeholder: t("global.name"),
              readOnly: viewMode
            }
          ),
          /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.name })
        ] }),
        /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
          /* @__PURE__ */ jsx(Label, { htmlFor: "contact", children: t("global.contact") }),
          /* @__PURE__ */ jsx(
            Input,
            {
              id: "contact",
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
          /* @__PURE__ */ jsx(Label, { htmlFor: "email", children: t("global.email") }),
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
          /* @__PURE__ */ jsx(Label, { htmlFor: "phone", children: t("global.phone") }),
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
      !viewMode && /* @__PURE__ */ jsx("div", { className: "customers-center flex gap-4", children: /* @__PURE__ */ jsxs(Button, { disabled: processing, children: [
        t(`global.actions.${verbName}`),
        " ",
        t("global.customer")
      ] }) })
    ] }),
    viewMode && /* @__PURE__ */ jsxs("div", { className: "space-y-6 pt-12", children: [
      /* @__PURE__ */ jsx(
        HeadingSmall,
        {
          title: t(`customers.statuses.${((_g = params.customer) == null ? void 0 : _g.status) || "enabled"}.section.title`),
          description: t(`customers.statuses.${((_h = params.customer) == null ? void 0 : _h.status) || "enabled"}.section.description`)
        }
      ),
      /* @__PURE__ */ jsxs("div", { className: `space-y-4 rounded-lg border ${isDisabled ? "border-primary-100 bg-primary-50" : "border-red-100 bg-red-50"} p-4`, children: [
        /* @__PURE__ */ jsxs("div", { className: `relative space-y-0.5 ${isDisabled ? "text-primary" : "text-red-600"}`, children: [
          /* @__PURE__ */ jsx("p", { className: "font-medium", children: t("global.warning.title") }),
          /* @__PURE__ */ jsx("p", { className: "text-sm", children: t("global.warning.description") })
        ] }),
        /* @__PURE__ */ jsx(Button, { variant: isDisabled ? "default" : "destructive", onClick: () => setDialogOpen(true), children: t(`customers.statuses.${((_i = params.customer) == null ? void 0 : _i.status) || "enabled"}.section.title`) }),
        /* @__PURE__ */ jsx(
          ConfirmsPassword,
          {
            title: t(`customers.statuses.${((_j = params.customer) == null ? void 0 : _j.status) || "enabled"}.confirmsPassword.title`, {
              customer: ((_k = params.customer) == null ? void 0 : _k.name) || ""
            }),
            description: t(`customers.statuses.${((_l = params.customer) == null ? void 0 : _l.status) || "enabled"}.confirmsPassword.description`),
            action: t(`customers.statuses.${((_m = params.customer) == null ? void 0 : _m.status) || "enabled"}.confirmsPassword.confirm`),
            verb: "update",
            path: `/customers/${(_n = params.customer) == null ? void 0 : _n.id}/change-status`,
            open: dialogOpen,
            onOpenChange: setDialogOpen
          }
        )
      ] })
    ] })
  ] });
}
export {
  CreateForm as default
};
