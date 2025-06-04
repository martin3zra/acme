import { jsx, jsxs, Fragment } from "react/jsx-runtime";
import { H as HeadingSmall, C as ConfirmsPassword } from "./heading-small.js";
import { B as Button } from "./button.js";
import { S as Separator } from "./separator.js";
import { S as Sheet, a as SheetContent, b as SheetHeader, c as SheetTitle, d as SheetDescription } from "./sheet.js";
import { u as useCallbackState } from "./use-callback-state.js";
import { u as useTranslation } from "./use-translation.js";
import { A as AppLayout } from "./app-layout.js";
import { usePage, Link, router } from "@inertiajs/react";
import { Ban, NotebookPen, Printer } from "lucide-react";
import { b as breadcrumbs } from "./constants3.js";
import { List } from "./Index11.js";
import { AddNewPayment } from "./add-new-payment.js";
import Show from "./Show2.js";
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
import "react";
import "./alert-destructive.js";
import "class-variance-authority";
import "@radix-ui/react-slot";
import "@radix-ui/react-separator";
import "./tooltip.js";
import "@radix-ui/react-tooltip";
import "@radix-ui/react-avatar";
import "./dropdown-menu.js";
import "@radix-ui/react-dropdown-menu";
import "@tabler/icons-react";
import "./app-logo-icon.js";
import "next-themes";
import "sonner";
import "./constants2.js";
import "./index13.js";
import "./table.js";
import "@tanstack/react-table";
import "./columns-definitions4.js";
import "./checkbox.js";
import "./use-number.js";
import "date-fns";
import "@radix-ui/react-checkbox";
import "./link-cell.js";
import "./status-badge.js";
import "date-fns/format";
import "./payment-summary.js";
function Index({
  auth,
  payments,
  payment,
  showPayment
}) {
  const t = useTranslation().trans;
  const [open, setOpen] = useCallbackState(showPayment);
  const [selectedPayment, setSelectedPayment] = useCallbackState(void 0);
  const [deleteDialogOpen, setDeleteDialogOpen] = useCallbackState(false);
  const page = usePage();
  const hasPayments = payments.length > 0;
  const onSelectPayment = (payment2, action) => {
    setSelectedPayment(payment2);
    if (action === "void") {
      setDeleteDialogOpen(true);
      return;
    }
    if (action === "edit") {
      router.visit(`/payments/${payment2.uuid}/edit`);
      return;
    }
    if (action !== "view") return;
    setOpen(
      (open2) => !open2,
      (newVal) => {
        if (newVal) findSelectedPayment(payment2.uuid);
      }
    );
  };
  const findSelectedPayment = (uuid) => {
    router.visit(page.url, {
      except: ["payments"],
      data: { id: uuid },
      preserveScroll: true,
      preserveState: true
    });
  };
  const onOpenChange = (open2) => {
    setOpen(open2);
    if (!open2) {
      router.replace({
        url: window.location.pathname,
        preserveScroll: true,
        preserveState: true
      });
    }
  };
  const modalHandler = (open2 = false) => {
    onOpenChange(open2);
    setDeleteDialogOpen(open2);
  };
  return /* @__PURE__ */ jsx(AppLayout, { user: auth.user, breadcrumbs, children: /* @__PURE__ */ jsxs("div", { className: "space-y-6", children: [
    hasPayments && /* @__PURE__ */ jsx(HeadingSmall, { title: t("payments.title"), description: t("payments.description"), rightPanel: /* @__PURE__ */ jsx(AddNewPayment, {}) }),
    !hasPayments && /* @__PURE__ */ jsx(Fragment, { children: /* @__PURE__ */ jsxs("div", { className: "absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]", children: [
      /* @__PURE__ */ jsx("h4", { className: "text-2xl", children: t("payments.emptyState.title") }),
      /* @__PURE__ */ jsx("p", { className: "text-sm text-gray-400", children: t("payments.emptyState.description") }),
      /* @__PURE__ */ jsx(AddNewPayment, {})
    ] }) }),
    hasPayments && /* @__PURE__ */ jsx(List, { data: payments, onSelectPayment }),
    payment && /* @__PURE__ */ jsx(Sheet, { open, onOpenChange, children: /* @__PURE__ */ jsxs(SheetContent, { side: "right", className: "m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-[1380px]", children: [
      /* @__PURE__ */ jsx(SheetHeader, { children: /* @__PURE__ */ jsxs("div", { className: "mr-6 flex items-start justify-between", children: [
        /* @__PURE__ */ jsxs("div", { className: "flex flex-col", children: [
          /* @__PURE__ */ jsx(SheetTitle, { children: page.props.auth.company.name }),
          /* @__PURE__ */ jsx(SheetDescription, { className: "text-[12px]", children: t("payments.viewPayment.description") })
        ] }),
        /* @__PURE__ */ jsxs("div", { className: "mx-4 flex gap-x-3", children: [
          payment.header.status !== "void" && /* @__PURE__ */ jsxs(Fragment, { children: [
            /* @__PURE__ */ jsxs(Button, { variant: "destructive", onClick: () => onSelectPayment(payment.header, "void"), children: [
              /* @__PURE__ */ jsx(Ban, {}),
              " ",
              t("global.actions.void")
            ] }),
            /* @__PURE__ */ jsx(Separator, { orientation: "vertical" }),
            /* @__PURE__ */ jsx(Button, { asChild: true, disabled: payment.header.status === "void", children: /* @__PURE__ */ jsxs(Link, { href: `/payments/${payment.header.uuid}/edit`, as: "button", children: [
              /* @__PURE__ */ jsx(NotebookPen, {}),
              " ",
              t("global.actions.edit")
            ] }) })
          ] }),
          /* @__PURE__ */ jsxs(Button, { children: [
            /* @__PURE__ */ jsx(Printer, {}),
            " ",
            t("global.actions.print")
          ] })
        ] })
      ] }) }),
      /* @__PURE__ */ jsxs("div", { className: "relative grid gap-4 px-4", children: [
        payment.header.status === "void" && /* @__PURE__ */ jsx("div", { className: "absolute inset-0 flex w-full items-center justify-center overflow-y-hidden bg-transparent", children: /* @__PURE__ */ jsx("h1", { className: "-rotate-45 border-8 border-red-500/25 p-8 text-8xl font-extrabold text-red-500/25", children: "VOID" }) }),
        /* @__PURE__ */ jsx(Show, { payment, auth })
      ] })
    ] }) }),
    selectedPayment && /* @__PURE__ */ jsx(
      ConfirmsPassword,
      {
        title: t("payments.confirmsPassword.title", { payment: selectedPayment.number }),
        description: t("payments.confirmsPassword.description", { total: selectedPayment.amount }),
        action: t("payments.confirmsPassword.confirm"),
        verb: "update",
        path: `/payments/${selectedPayment.uuid}/void`,
        open: deleteDialogOpen,
        onOpenChange: modalHandler
      }
    )
  ] }) });
}
export {
  Index as default
};
