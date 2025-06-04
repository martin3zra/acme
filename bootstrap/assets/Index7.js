import { jsx, jsxs, Fragment } from "react/jsx-runtime";
import { H as HeadingSmall, C as ConfirmsPassword } from "./heading-small.js";
import { B as Button } from "./button.js";
import { S as Sheet, a as SheetContent, b as SheetHeader, c as SheetTitle, d as SheetDescription } from "./sheet.js";
import { u as useVerb } from "./use-verbs.js";
import { u as useTranslation } from "./use-translation.js";
import { A as AppLayout } from "./app-layout.js";
import { Plus } from "lucide-react";
import { useState, useEffect } from "react";
import { d as defaultBreadcrumbs } from "./index13.js";
import { List } from "./Index8.js";
import CreateForm from "./CreateForm2.js";
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
import "@inertiajs/react";
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
import "./columns-definitions3.js";
import "./checkbox.js";
import "./use-number.js";
import "date-fns";
import "@radix-ui/react-checkbox";
import "./status-badge.js";
import "./select.js";
import "@radix-ui/react-select";
const breadcrumbs = [
  ...defaultBreadcrumbs,
  {
    title: "items.title",
    href: "/items"
  }
];
function Index({ auth, items, taxes, units }) {
  var _a, _b;
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedItem, setSelectedItem] = useState({
    item: void 0,
    taxes,
    units,
    action: "create"
  });
  const verbName = useVerb().action(selectedItem.action);
  const hasItems = items.length > 0;
  const onCreateNewItem = () => {
    setSelectedItem({ item: void 0, taxes, units, action: "create" });
    setOpen(!open);
  };
  const onSelectItem = (item, action) => {
    setSelectedItem({ item, taxes, units, action });
  };
  const onOpenChange = (open2) => {
    setOpen(open2);
    if (!open2) setSelectedItem({ item: void 0, taxes, units, action: "create" });
  };
  useEffect(() => {
    if (selectedItem && selectedItem.item !== void 0) {
      if (selectedItem.action !== "trash") {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedItem]);
  const modalHandler = (open2 = false) => {
    onOpenChange(open2);
    setDeleteDialogOpen(open2);
  };
  return /* @__PURE__ */ jsx(AppLayout, { user: auth.user, breadcrumbs, children: /* @__PURE__ */ jsxs("div", { className: "space-y-6", children: [
    hasItems && /* @__PURE__ */ jsx(
      HeadingSmall,
      {
        title: t("items.title"),
        description: t("items.description"),
        rightPanel: /* @__PURE__ */ jsxs(Button, { onClick: onCreateNewItem, children: [
          /* @__PURE__ */ jsx(Plus, {}),
          " ",
          t("items.newItem.title")
        ] })
      }
    ),
    !hasItems && /* @__PURE__ */ jsx(Fragment, { children: /* @__PURE__ */ jsxs("div", { className: "absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]", children: [
      /* @__PURE__ */ jsx("h4", { className: "text-2xl", children: t("items.emptyState.title") }),
      /* @__PURE__ */ jsx("p", { className: "text-sm text-gray-400", children: t("items.emptyState.description") }),
      /* @__PURE__ */ jsxs(Button, { onClick: onCreateNewItem, children: [
        /* @__PURE__ */ jsx(Plus, {}),
        " ",
        t("items.newItem.title")
      ] })
    ] }) }),
    hasItems && /* @__PURE__ */ jsx(List, { data: items, onSelectItem }),
    /* @__PURE__ */ jsx(Sheet, { open, onOpenChange, children: /* @__PURE__ */ jsxs(SheetContent, { side: "right", className: "m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl", children: [
      /* @__PURE__ */ jsxs(SheetHeader, { children: [
        /* @__PURE__ */ jsxs(SheetTitle, { children: [
          t(`global.actions.${verbName}`),
          " ",
          t("global.item")
        ] }),
        /* @__PURE__ */ jsx(SheetDescription, { className: "text-[12px]", children: t(`items.newItem.description`) })
      ] }),
      /* @__PURE__ */ jsx("div", { className: "grid gap-4 px-4", children: /* @__PURE__ */ jsx(CreateForm, { params: selectedItem, onFinish: () => modalHandler(false) }) })
    ] }) }),
    selectedItem.item && /* @__PURE__ */ jsx(
      ConfirmsPassword,
      {
        title: t(`items.confirmsPassword.title`, { item: selectedItem == null ? void 0 : selectedItem.item.name }),
        description: t(`items.confirmsPassword.description`, { item: (_a = selectedItem == null ? void 0 : selectedItem.item) == null ? void 0 : _a.name }),
        action: t(`items.confirmsPassword.confirm`),
        verb: "destroy",
        path: `/items/${(_b = selectedItem == null ? void 0 : selectedItem.item) == null ? void 0 : _b.id}`,
        open: deleteDialogOpen,
        onOpenChange: modalHandler
      }
    )
  ] }) });
}
export {
  Index as default
};
