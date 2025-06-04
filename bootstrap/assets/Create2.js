import { jsxs, jsx } from "react/jsx-runtime";
import { A as AlertDestructive } from "./alert-destructive.js";
import { I as InputError } from "./input-error.js";
import { u as usePersistedState, C as Calendar, A as AlertDialog, a as AlertDialogContent, b as AlertDialogHeader, c as AlertDialogTitle, d as AlertDialogDescription, e as AlertDialogFooter, f as AlertDialogCancel, g as AlertDialogAction } from "./use-persisted-state.js";
import { B as Button } from "./button.js";
import { L as Label } from "./label.js";
import { C as CustomerSection, P as Popover, a as PopoverTrigger, b as PopoverContent } from "./customer-section.js";
import { u as useHeader } from "./use-headers.js";
import { u as useNumber } from "./use-number.js";
import { u as useDebounced } from "./use-debounced.js";
import { u as useTranslation } from "./use-translation.js";
import { A as AppLayout } from "./app-layout.js";
import { c as cn } from "./utils.js";
import { Textarea } from "@headlessui/react";
import { usePage, useForm, router } from "@inertiajs/react";
import { format } from "date-fns/format";
import { CalendarIcon } from "lucide-react";
import React__default, { useEffect } from "react";
import CheckoutForm from "./checkout-form.js";
import { d as defaultPaymentForm, c as createPaymentBreadcrumbs } from "./constants3.js";
import { List } from "./lines-payment.js";
import "class-variance-authority";
import "@radix-ui/react-alert-dialog";
import "react-day-picker";
import "@radix-ui/react-slot";
import "@radix-ui/react-label";
import "@radix-ui/react-popover";
import "./command.js";
import "cmdk";
import "./dialog.js";
import "@radix-ui/react-dialog";
import "./sheet.js";
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
import "clsx";
import "tailwind-merge";
import "./constants2.js";
import "./bank-transfer.js";
import "./form-section.js";
import "./input.js";
import "./card-form.js";
import "./select.js";
import "@radix-ui/react-select";
import "./check-form.js";
import "./input-view.js";
import "./index13.js";
import "./table.js";
import "@tanstack/react-table";
import "./columns-definitions5.js";
import "./checkbox.js";
import "date-fns";
import "@radix-ui/react-checkbox";
function Create({
  auth,
  customer,
  customers,
  receivables,
  invoice_uuid,
  forceInitial
}) {
  const t = useTranslation().trans;
  const { currency } = useNumber();
  const [openCancelConfirmation, setCancelConfirmation] = React__default.useState(false);
  const [openCheckout, setCheckout] = React__default.useState(false);
  const [initialized, setInitialized] = React__default.useState(false);
  const [open, setOpen] = React__default.useState(false);
  const [search, setSearch] = React__default.useState("");
  const dedbouncedSearch = useDebounced(search, 500);
  const [paymentForm, setPaymentForm, removePaymentForm] = usePersistedState(
    "payment",
    { ...defaultPaymentForm, header: { ...defaultPaymentForm.header, customer } },
    forceInitial
  );
  const [rowSelection, setRowSelection] = React__default.useState({});
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage().props;
  const { post, transform, processing, errors } = useForm({
    customer_id: 0,
    date: /* @__PURE__ */ new Date(),
    lines: []
  });
  useEffect(() => {
    const _rowSelection = {};
    paymentForm.lines.filter((line) => line.payment > 0).map((line) => {
      _rowSelection[`${line.id.toString()}`] = true;
    });
    if (Object.keys(_rowSelection).length > 0) {
      setRowSelection(_rowSelection);
    }
    if (receivables === void 0 || initialized) return;
    const lines = [];
    let selectedRowId = -1;
    receivables.map((receivable) => {
      selectedRowId = invoice_uuid === receivable.invoice.uuid ? receivable.invoice.id : -1;
      const line = {
        ...receivable.invoice,
        payment: invoice_uuid === receivable.invoice.uuid ? receivable.invoice.amount_due : 0,
        discount: 0,
        balance: 0
      };
      lines.push(line);
    });
    if (selectedRowId > 0) {
      setRowSelection((prev) => ({
        ...prev,
        [`${selectedRowId.toString()}`]: true
      }));
    }
    setPaymentForm((prev) => ({
      ...prev,
      lines: [...lines]
    }));
    setInitialized(true);
  }, [receivables, paymentForm, setPaymentForm, invoice_uuid, initialized]);
  useEffect(() => {
    const searchCustomer = () => {
      router.reload({ only: ["customers"], data: { search: dedbouncedSearch }, preserveUrl: true });
    };
    if (dedbouncedSearch) searchCustomer();
  }, [dedbouncedSearch]);
  const totalPaid = () => {
    return paymentForm.lines.reduce((acc, line) => {
      return acc + line.payment;
    }, 0);
  };
  const handleRecordPayment = (event) => {
    var _a;
    event.preventDefault();
    const payload = {
      customer_id: (_a = paymentForm.header.customer) == null ? void 0 : _a.uuid,
      date: paymentForm.header.date,
      amount: totalPaid(),
      notes: paymentForm.header.notes,
      lines: paymentForm.lines.filter((line) => line.payment > 0).map((line) => {
        return { uuid: line.uuid, amount_due: line.amount_due, payment: line.payment, discount: line.discount };
      }),
      payment: paymentForm.payment
    };
    transform((data) => ({
      ...data,
      ...payload
    }));
    post("/payments", {
      ...headers,
      preserveState: "errors",
      onSuccess: () => {
        removePaymentForm();
        router.get("/payments");
      }
    });
  };
  const handleCustomerSelection = (customer2) => {
    setPaymentForm(() => {
      return { ...paymentForm, header: { ...paymentForm.header, customer: customer2 }, lines: [] };
    });
    setOpen(false);
    if (customer2 !== void 0) {
      router.reload({ only: ["receivables"], data: { customer_id: customer2.uuid }, preserveUrl: true });
    }
  };
  const handleDateChange = (date) => {
    setPaymentForm(() => {
      return { ...paymentForm, header: { ...paymentForm.header, date } };
    });
  };
  const handleCellChange = (inputId, newValue) => {
    const index = paymentForm.lines.findIndex((l) => l.uuid === inputId);
    if (index === -1) return;
    setRowSelection((prev) => ({
      ...prev,
      [`${paymentForm.lines[index].id.toString()}`]: true
    }));
    paymentForm.lines[index].payment = Number(newValue);
    setPaymentForm((prev) => ({
      ...prev,
      lines: [...paymentForm.lines]
    }));
  };
  const onSelectionChange = (selection) => {
    paymentForm.lines = paymentForm.lines.map((line) => ({ ...line, payment: 0, balance: line.amount_due }));
    const selectedIds = Object.keys(selection);
    selectedIds.map((id) => {
      const index = paymentForm.lines.findIndex((l) => l.id === Number(id));
      if (index === -1) return;
      paymentForm.lines[index].payment = paymentForm.lines[index].amount_due;
    });
    setPaymentForm((prev) => ({
      ...prev,
      lines: [...paymentForm.lines]
    }));
  };
  const performPaymentCancelation = (event) => {
    event.preventDefault();
    removePaymentForm();
    router.get("/payments");
  };
  const handleCheckout = (event) => {
    event.preventDefault();
    if (totalPaid() === 0) return;
    setCheckout(true);
  };
  const handleCheckoutChange = (method, form) => {
    setPaymentForm(() => {
      return { ...paymentForm, payment: { ...paymentForm.payment, [method]: form } };
    });
  };
  return /* @__PURE__ */ jsxs(AppLayout, { user: auth.user, breadcrumbs: createPaymentBreadcrumbs, children: [
    /* @__PURE__ */ jsx(AppLayout.Actions, { children: /* @__PURE__ */ jsxs("div", { className: "flex justify-end gap-x-6", children: [
      /* @__PURE__ */ jsx(Button, { variant: "secondary", onClick: () => setCancelConfirmation(true), children: t("global.actions.cancel") }),
      /* @__PURE__ */ jsx(Button, { onClick: handleCheckout, disabled: totalPaid() === 0 || processing, children: t("global.actions.checkout") })
    ] }) }),
    /* @__PURE__ */ jsxs("div", { className: "grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4 bg-gray-50/10", children: [
      !openCheckout && propsErrors.status && /* @__PURE__ */ jsx("div", { className: "col-span-12", children: /* @__PURE__ */ jsx(AlertDestructive, { description: propsErrors.status, onDestroy: () => delete propsErrors.status }) }),
      /* @__PURE__ */ jsxs("div", { className: "z-50 col-span-12 grid min-h-42 grid-cols-2 gap-x-6", children: [
        /* @__PURE__ */ jsx(
          CustomerSection,
          {
            customer: paymentForm.header.customer,
            customers,
            errors,
            handleCustomerSelection,
            setSearch,
            setOpen,
            open,
            dedbouncedSearch
          }
        ),
        /* @__PURE__ */ jsxs("div", { className: "grid grid-cols-12", children: [
          /* @__PURE__ */ jsxs("div", { className: "col-span-6 flex flex-col gap-y-6", children: [
            /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-y-2", children: [
              /* @__PURE__ */ jsx(Label, { htmlFor: "date", children: t("global.date") }),
              /* @__PURE__ */ jsxs(Popover, { children: [
                /* @__PURE__ */ jsx(PopoverTrigger, { asChild: true, children: /* @__PURE__ */ jsxs(
                  Button,
                  {
                    variant: "outline",
                    className: cn("w-[280px] justify-start text-left font-normal", !paymentForm.header.date && "text-muted-foreground"),
                    children: [
                      /* @__PURE__ */ jsx(CalendarIcon, {}),
                      paymentForm.header.date ? format(paymentForm.header.date, "PPP") : /* @__PURE__ */ jsx("span", { children: t("global.datePlaceholder") })
                    ]
                  }
                ) }),
                /* @__PURE__ */ jsx(PopoverContent, { className: "w-auto p-0", children: /* @__PURE__ */ jsx(
                  Calendar,
                  {
                    mode: "single",
                    defaultMonth: paymentForm.header.date,
                    selected: paymentForm.header.date,
                    onSelect: handleDateChange,
                    initialFocus: true
                  }
                ) })
              ] }),
              /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.date })
            ] }),
            /* @__PURE__ */ jsx("div", { className: "flex flex-col", children: /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-y-2", children: [
              /* @__PURE__ */ jsx(Label, { className: "text-sm/6 font-medium", children: t("global.notes") }),
              /* @__PURE__ */ jsx(
                Textarea,
                {
                  name: "notes",
                  rows: 4,
                  className: "focus:no-data-focus:outline-none block resize-none rounded-lg border px-3 py-1.5 text-sm/6 data-focus:outline-2 data-focus:-outline-offset-2 data-focus:outline-white/25",
                  defaultValue: paymentForm.header.notes,
                  onChange: (e) => setPaymentForm(() => {
                    return { ...paymentForm, header: { ...paymentForm.header, notes: e.currentTarget.value } };
                  })
                }
              )
            ] }) })
          ] }),
          /* @__PURE__ */ jsx("div", { className: "col-span-6 grid place-items-end", children: /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-x-2", children: [
            /* @__PURE__ */ jsx(Label, { className: "text-muted-foreground block text-end text-lg", children: t("global.totalReceived") }),
            /* @__PURE__ */ jsx(Label, { className: "block text-end text-4xl", children: currency(totalPaid()) })
          ] }) })
        ] })
      ] }),
      /* @__PURE__ */ jsx("div", { className: "col-span-12", children: /* @__PURE__ */ jsx(
        List,
        {
          data: paymentForm.lines,
          rowSelection,
          setRowSelection,
          onSelectPaymentLine: () => {
          },
          onValueChange: handleCellChange,
          onSelectionChange
        }
      ) }),
      /* @__PURE__ */ jsx(AlertDialog, { open: openCancelConfirmation, onOpenChange: setCancelConfirmation, children: /* @__PURE__ */ jsxs(AlertDialogContent, { children: [
        /* @__PURE__ */ jsxs(AlertDialogHeader, { children: [
          /* @__PURE__ */ jsx(AlertDialogTitle, { children: t("payments.confirmsCancelation.title") }),
          /* @__PURE__ */ jsx(AlertDialogDescription, { children: t("payments.confirmsCancelation.description") })
        ] }),
        /* @__PURE__ */ jsxs(AlertDialogFooter, { children: [
          /* @__PURE__ */ jsx(AlertDialogCancel, { children: t("global.actions.cancel") }),
          /* @__PURE__ */ jsx(AlertDialogAction, { onClick: performPaymentCancelation, children: t("payments.confirmsCancelation.confirm") })
        ] })
      ] }) }),
      /* @__PURE__ */ jsx(
        CheckoutForm,
        {
          action: t("global.actions.recordPayment"),
          openCheckout,
          setCheckout,
          paymentForm: paymentForm.payment,
          totalAmount: totalPaid(),
          onCompleteCheckout: handleRecordPayment,
          processing,
          setCancelConfirmation,
          errors: propsErrors,
          onCheckoutChange: handleCheckoutChange,
          currency,
          t
        }
      )
    ] })
  ] });
}
export {
  Create as default
};
