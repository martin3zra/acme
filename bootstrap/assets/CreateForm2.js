import { jsxs, jsx } from "react/jsx-runtime";
import { H as HeadingSmall, C as ConfirmsPassword } from "./heading-small.js";
import { I as InputError } from "./input-error.js";
import { B as Button } from "./button.js";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { S as Select, a as SelectTrigger, b as SelectValue, c as SelectContent, d as SelectItem } from "./select.js";
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
import "@radix-ui/react-select";
function CreateForm({ onFinish, params }) {
  var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l, _m, _n, _o;
  const t = useTranslation().trans;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage().props;
  const { data, setData, post, put, errors, reset, processing } = useForm({
    id: (_a = params.item) == null ? void 0 : _a.id,
    name: ((_b = params.item) == null ? void 0 : _b.name) || "",
    description: ((_c = params.item) == null ? void 0 : _c.description) || "",
    price: ((_d = params.item) == null ? void 0 : _d.price) || 0,
    tax_id: ((_e = params.item) == null ? void 0 : _e.tax.id) || 0,
    unit_id: ((_f = params.item) == null ? void 0 : _f.unit.id) || 0
  });
  const viewMode = params.action === "view";
  const isDisabled = ((_g = params.item) == null ? void 0 : _g.status) === "disabled";
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
    if (params.action === "create") post("/items", options);
    if (params.action === "edit") put(`/items/${params.item.id}`, options);
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
              placeholder: "Item name",
              readOnly: viewMode
            }
          ),
          /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.name })
        ] }),
        /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
          /* @__PURE__ */ jsx(Label, { children: t("global.unit") }),
          /* @__PURE__ */ jsxs(
            Select,
            {
              onValueChange: (value) => setData("unit_id", Number(value)),
              disabled: viewMode,
              defaultValue: data.unit_id.toString(),
              value: data.unit_id.toString(),
              required: true,
              children: [
                /* @__PURE__ */ jsx(SelectTrigger, { className: "w-full", children: /* @__PURE__ */ jsx(SelectValue, { placeholder: "Select item unit" }) }),
                /* @__PURE__ */ jsx(SelectContent, { className: "", children: params.units.map((unit) => /* @__PURE__ */ jsx(SelectItem, { value: unit.id.toString(), children: unit.name }, unit.id)) })
              ]
            }
          ),
          /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.unit_id })
        ] })
      ] }),
      /* @__PURE__ */ jsx("div", { className: "grid grid-cols-2 gap-2", children: /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-between", children: [
        /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-2", children: [
          /* @__PURE__ */ jsx(Label, { children: t("global.tax") }),
          /* @__PURE__ */ jsxs(
            Select,
            {
              onValueChange: (value) => setData("tax_id", Number(value)),
              disabled: viewMode,
              defaultValue: data.tax_id.toString(),
              value: data.tax_id.toString(),
              required: true,
              children: [
                /* @__PURE__ */ jsx(SelectTrigger, { className: "w-44", children: /* @__PURE__ */ jsx(SelectValue, { placeholder: "Select item tax" }) }),
                /* @__PURE__ */ jsx(SelectContent, { className: "", children: params.taxes.map((tax) => /* @__PURE__ */ jsx(SelectItem, { value: tax.id.toString(), children: tax.name }, tax.id)) })
              ]
            }
          ),
          /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.tax_id })
        ] }),
        /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-2", children: [
          /* @__PURE__ */ jsx(Label, { htmlFor: "price", children: t("global.price") }),
          /* @__PURE__ */ jsx(
            Input,
            {
              id: "price",
              type: "number",
              className: "mt-0 block w-full max-w-40 text-right",
              value: data.price,
              onChange: (e) => setData("price", e.target.valueAsNumber),
              placeholder: "Jane Doe",
              readOnly: viewMode
            }
          ),
          /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.price })
        ] })
      ] }) }),
      /* @__PURE__ */ jsx("div", { className: "grid grid-cols-2 gap-2", children: /* @__PURE__ */ jsxs("div", { className: "grid gap-2", children: [
        /* @__PURE__ */ jsx(Label, { htmlFor: "description", children: t("global.description") }),
        /* @__PURE__ */ jsx(
          "textarea",
          {
            id: "description",
            className: "mt-1 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500",
            value: data.description,
            onChange: (e) => setData("description", e.target.value),
            placeholder: "Wrire some description here...",
            rows: 3,
            readOnly: viewMode
          }
        ),
        /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.description })
      ] }) }),
      !viewMode && /* @__PURE__ */ jsx("div", { className: "flex items-center gap-4", children: /* @__PURE__ */ jsxs(Button, { disabled: processing, children: [
        t(`global.actions.${verbName}`),
        " ",
        t("global.item")
      ] }) })
    ] }),
    viewMode && /* @__PURE__ */ jsxs("div", { className: "space-y-6 pt-12", children: [
      /* @__PURE__ */ jsx(
        HeadingSmall,
        {
          title: t(`items.statuses.${((_h = params.item) == null ? void 0 : _h.status) || "enabled"}.section.title`),
          description: t(`items.statuses.${((_i = params.item) == null ? void 0 : _i.status) || "enabled"}.section.description`)
        }
      ),
      /* @__PURE__ */ jsxs("div", { className: `space-y-4 rounded-lg border ${isDisabled ? "border-primary-100 bg-primary-50" : "border-red-100 bg-red-50"} p-4`, children: [
        /* @__PURE__ */ jsxs("div", { className: `relative space-y-0.5 ${isDisabled ? "text-primary" : "text-red-600"}`, children: [
          /* @__PURE__ */ jsx("p", { className: "font-medium", children: t("global.warning.title") }),
          /* @__PURE__ */ jsx("p", { className: "text-sm", children: t("global.warning.description") })
        ] }),
        /* @__PURE__ */ jsx(Button, { variant: isDisabled ? "default" : "destructive", onClick: () => setDialogOpen(true), children: t(`items.statuses.${((_j = params.item) == null ? void 0 : _j.status) || "enabled"}.section.title`) }),
        /* @__PURE__ */ jsx(
          ConfirmsPassword,
          {
            title: t(`items.statuses.${((_k = params.item) == null ? void 0 : _k.status) || "enabled"}.confirmsPassword.title`, { item: ((_l = params.item) == null ? void 0 : _l.name) || "" }),
            description: t(`items.statuses.${((_m = params.item) == null ? void 0 : _m.status) || "enabled"}.confirmsPassword.description`),
            action: t(`items.statuses.${((_n = params.item) == null ? void 0 : _n.status) || "enabled"}.confirmsPassword.confirm`),
            verb: "update",
            path: `/items/${(_o = params.item) == null ? void 0 : _o.id}/change-status`,
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
