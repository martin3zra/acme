import { jsx, jsxs, Fragment } from "react/jsx-runtime";
import { H as HeadingSmall, C as ConfirmsPassword } from "./heading-small.js";
import { B as Button } from "./button.js";
import { S as Sheet, a as SheetContent, b as SheetHeader, c as SheetTitle, d as SheetDescription } from "./sheet.js";
import { u as useVerb } from "./use-verbs.js";
import { u as useTranslation } from "./use-translation.js";
import { A as AppLayout } from "./app-layout.js";
import { router } from "@inertiajs/react";
import { Plus } from "lucide-react";
import { useState, useEffect } from "react";
import { List } from "./Index2.js";
import CreateForm from "./CreateForm.js";
import { d as defaultBreadcrumbs } from "./index13.js";
import "./input-error.js";
import "./dialog.js";
import "@radix-ui/react-dialog";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "./input.js";
import "./label.js";
import "@radix-ui/react-label";
import "./use-headers.js";
import "./alert-destructive.js";
import "class-variance-authority";
import "@radix-ui/react-slot";
import "./tooltip.js";
import "@radix-ui/react-tooltip";
import "@radix-ui/react-avatar";
import "./dropdown-menu.js";
import "@radix-ui/react-dropdown-menu";
import "@tabler/icons-react";
import "./app-logo-icon.js";
import "./separator.js";
import "@radix-ui/react-separator";
import "next-themes";
import "sonner";
import "./table.js";
import "@tanstack/react-table";
import "./columns-definitions.js";
import "./checkbox.js";
import "./use-number.js";
import "date-fns";
import "@radix-ui/react-checkbox";
import "./status-badge.js";
const breadcrumbs = [
  ...defaultBreadcrumbs,
  {
    title: "customers.title",
    href: "/customers"
  }
];
function Index({ auth, customers, customer }) {
  var _a, _b, _c;
  const t = useTranslation().trans;
  const [open, setOpen] = useState(customer !== void 0);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState({
    customer,
    action: customer !== void 0 ? "view" : "create"
  });
  const verbName = useVerb().action(selectedCustomer.action);
  const hasCustomers = customers.length > 0;
  const onCreateNewCustomer = () => {
    setSelectedCustomer({ customer: void 0, action: "create" });
    setOpen(!open);
  };
  const onSelectCustomer = (customer2, action) => {
    if (action === "record-payment") {
      router.visit(`/payments/create`, { data: { customer_id: customer2.uuid } });
      return;
    }
    setSelectedCustomer({ customer: customer2, action });
  };
  const onOpenChange = (open2) => {
    setOpen(open2);
    if (!open2) setSelectedCustomer({ customer: void 0, action: "create" });
  };
  useEffect(() => {
    if (selectedCustomer && selectedCustomer.customer !== void 0) {
      if (selectedCustomer.action !== "trash") {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedCustomer]);
  const modalHandler = (open2 = false) => {
    onOpenChange(open2);
    setDeleteDialogOpen(open2);
  };
  return /* @__PURE__ */ jsx(AppLayout, { user: auth.user, breadcrumbs, children: /* @__PURE__ */ jsxs("div", { className: "space-y-6", children: [
    hasCustomers && /* @__PURE__ */ jsx(
      HeadingSmall,
      {
        title: t("customers.title"),
        description: t("customers.description"),
        rightPanel: /* @__PURE__ */ jsxs(Button, { onClick: onCreateNewCustomer, children: [
          /* @__PURE__ */ jsx(Plus, {}),
          " ",
          t("customers.newCustomer.title")
        ] })
      }
    ),
    !hasCustomers && /* @__PURE__ */ jsx(Fragment, { children: /* @__PURE__ */ jsxs("div", { className: "absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]", children: [
      /* @__PURE__ */ jsx("h4", { className: "text-2xl", children: t("customers.emptyState.title") }),
      /* @__PURE__ */ jsx("p", { className: "text-sm text-gray-400", children: t("customers.emptyState.description") }),
      /* @__PURE__ */ jsxs(Button, { onClick: onCreateNewCustomer, children: [
        /* @__PURE__ */ jsx(Plus, {}),
        " ",
        t("customers.newCustomer.title")
      ] })
    ] }) }),
    hasCustomers && /* @__PURE__ */ jsx(List, { data: customers, onSelectCustomer }),
    /* @__PURE__ */ jsx(Sheet, { open, onOpenChange, children: /* @__PURE__ */ jsxs(SheetContent, { side: "right", className: "m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl", children: [
      /* @__PURE__ */ jsxs(SheetHeader, { children: [
        /* @__PURE__ */ jsxs(SheetTitle, { children: [
          t(`global.actions.${verbName}`),
          " ",
          t(`global.customer`).toLocaleLowerCase()
        ] }),
        /* @__PURE__ */ jsx(SheetDescription, { className: "text-[12px]", children: t(`customers.newCustomer.description`) })
      ] }),
      /* @__PURE__ */ jsx("div", { className: "grid gap-4 px-4", children: /* @__PURE__ */ jsx(CreateForm, { params: selectedCustomer, onFinish: () => modalHandler(false) }) })
    ] }) }),
    selectedCustomer.customer && /* @__PURE__ */ jsx(
      ConfirmsPassword,
      {
        title: t(`customers.confirmsPassword.title`, { customer: (_a = selectedCustomer == null ? void 0 : selectedCustomer.customer) == null ? void 0 : _a.name }),
        description: t(`customers.confirmsPassword.description`, { customer: (_b = selectedCustomer == null ? void 0 : selectedCustomer.customer) == null ? void 0 : _b.name }),
        action: t(`customers.confirmsPassword.confirm`),
        verb: "destroy",
        path: `/customers/${(_c = selectedCustomer == null ? void 0 : selectedCustomer.customer) == null ? void 0 : _c.id}`,
        open: deleteDialogOpen,
        onOpenChange: modalHandler
      }
    )
  ] }) });
}
export {
  Index as default
};
