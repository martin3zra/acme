import { jsx, jsxs, Fragment } from "react/jsx-runtime";
import { u as useVerb, H as HeadingSmall, C as CreateForm } from "./CreateForm.js";
import { B as Button } from "./button.js";
import { A as AuthenticatedLayout, S as Sheet, a as SheetContent, b as SheetHeader, c as SheetTitle, d as SheetDescription } from "./authenticated-layout.js";
import { Plus } from "lucide-react";
import { useState, useEffect } from "react";
import { List } from "./Index2.js";
import { DeleteCustomer } from "./delete-customer.js";
import "./label.js";
import "@radix-ui/react-label";
import "./use-headers.js";
import "@radix-ui/react-dialog";
import "@inertiajs/react";
import "./input.js";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "clsx";
import "tailwind-merge";
import "@tabler/icons-react";
import "@radix-ui/react-tooltip";
import "@radix-ui/react-avatar";
import "./dropdown-menu.js";
import "@radix-ui/react-dropdown-menu";
import "@radix-ui/react-separator";
import "next-themes";
import "sonner";
import "@tanstack/react-table";
import "./columns-definitions.js";
import "@radix-ui/react-checkbox";
const breadcrumbs = [
  {
    title: "Home",
    href: "/home"
  },
  {
    title: "Customers",
    href: "/customers"
  }
];
function Index({ auth, customers }) {
  const [open, setOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState({ customer: void 0, action: "create" });
  const verbName = useVerb().action(selectedCustomer.action);
  const hasCustomers = customers.length > 0;
  const onCreateNewCustomer = () => {
    setSelectedCustomer({ customer: void 0, action: "create" });
    setOpen(!open);
  };
  const onSelectCustomer = (customer, action) => {
    setSelectedCustomer({ customer, action });
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
  return /* @__PURE__ */ jsx(AuthenticatedLayout, { user: auth.user, breadcrumbs, children: /* @__PURE__ */ jsxs("div", { className: "space-y-6", children: [
    hasCustomers && /* @__PURE__ */ jsx(
      HeadingSmall,
      {
        title: "Customers",
        description: "All created customers are shown here.",
        rightPanel: /* @__PURE__ */ jsxs(Button, { onClick: onCreateNewCustomer, children: [
          /* @__PURE__ */ jsx(Plus, {}),
          " Add Customers"
        ] })
      }
    ),
    !hasCustomers && /* @__PURE__ */ jsx(Fragment, { children: /* @__PURE__ */ jsxs("div", { className: "absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]", children: [
      /* @__PURE__ */ jsx("h4", { className: "text-2xl", children: "Create your first customer" }),
      /* @__PURE__ */ jsx("p", { className: "text-sm text-gray-400", children: "Once you create your customer, it will appear here." }),
      /* @__PURE__ */ jsx(Button, { onClick: onCreateNewCustomer, children: "+ Create Customer" })
    ] }) }),
    hasCustomers && /* @__PURE__ */ jsx(List, { data: customers, onSelectCustomer }),
    /* @__PURE__ */ jsx(Sheet, { open, onOpenChange, children: /* @__PURE__ */ jsxs(SheetContent, { side: "right", className: "m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl", children: [
      /* @__PURE__ */ jsxs(SheetHeader, { children: [
        /* @__PURE__ */ jsxs(SheetTitle, { children: [
          verbName,
          " Customer"
        ] }),
        /* @__PURE__ */ jsx(SheetDescription, { className: "text-[12px]", children: "Create a new customer" })
      ] }),
      /* @__PURE__ */ jsx("div", { className: "grid gap-4 px-4", children: /* @__PURE__ */ jsx(CreateForm, { params: selectedCustomer, onFinish: () => modalHandler(false) }) })
    ] }) }),
    /* @__PURE__ */ jsx(DeleteCustomer, { customer: selectedCustomer.customer, open: deleteDialogOpen, onOpenChange: modalHandler })
  ] }) });
}
export {
  Index as default
};
