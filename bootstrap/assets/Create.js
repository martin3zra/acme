import { jsxs, jsx, Fragment } from "react/jsx-runtime";
import { A as AlertDestructive } from "./alert-destructive.js";
import { I as InputError } from "./input-error.js";
import { u as usePersistedState, C as Calendar, A as AlertDialog, a as AlertDialogContent, b as AlertDialogHeader, c as AlertDialogTitle, d as AlertDialogDescription, e as AlertDialogFooter, f as AlertDialogCancel, g as AlertDialogAction } from "./use-persisted-state.js";
import { B as Button } from "./button.js";
import { I as Input } from "./input.js";
import { L as Label } from "./label.js";
import { C as CustomerSection, P as Popover, a as PopoverTrigger, b as PopoverContent } from "./customer-section.js";
import { S as Select, a as SelectTrigger, b as SelectValue, c as SelectContent, d as SelectItem } from "./select.js";
import { S as Separator } from "./separator.js";
import { u as useHeader } from "./use-headers.js";
import { u as useNumber } from "./use-number.js";
import { u as useDebounced } from "./use-debounced.js";
import { u as useTranslation } from "./use-translation.js";
import { A as AppLayout } from "./app-layout.js";
import { c as cn, i as isNotEmpty, a as addDays } from "./utils.js";
import { Textarea } from "@headlessui/react";
import { usePage, useForm, router } from "@inertiajs/react";
import { format } from "date-fns";
import { CalendarIcon } from "lucide-react";
import React__default, { useEffect, useCallback } from "react";
import { d as defaultDiscount, c as createBreadcrumbs, p as paymentTerms, a as defaultInvoiceForm } from "./constants.js";
import CheckoutForm from "./checkout-form.js";
import { Lines } from "./lines.js";
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
import "@radix-ui/react-select";
import "@radix-ui/react-separator";
import "./sheet.js";
import "./tooltip.js";
import "@radix-ui/react-tooltip";
import "@radix-ui/react-avatar";
import "./dropdown-menu.js";
import "@radix-ui/react-dropdown-menu";
import "@tabler/icons-react";
import "./app-logo-icon.js";
import "next-themes";
import "sonner";
import "clsx";
import "tailwind-merge";
import "./constants2.js";
import "./index13.js";
import "./bank-transfer.js";
import "./form-section.js";
import "./card-form.js";
import "./check-form.js";
import "./input-view.js";
import "./lines-column-headers.js";
import "./lines-form.js";
import "./lines-line.js";
function Create({
  auth,
  customers,
  items,
  item,
  tax_receipts
}) {
  const t = useTranslation().trans;
  const currency = useNumber().currency;
  const [open, setOpen] = React__default.useState(false);
  const [openCancelConfirmation, setCancelConfirmation] = React__default.useState(false);
  const [openCheckout, setCheckout] = React__default.useState(false);
  const [isEditing, setEditing] = React__default.useState(false);
  const referenceInputRef = React__default.useRef(null);
  const qtyInputRef = React__default.useRef(null);
  const [search, setSearch] = React__default.useState("");
  const dedbouncedSearch = useDebounced(search, 500);
  const [amount, setAmount] = React__default.useState(0);
  const [invoiceForm, setInvoiceForm, removeInvoiceForm] = usePersistedState("invoice", defaultInvoiceForm);
  const [currentItem, setCurrentItem] = React__default.useState(void 0);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage().props;
  const { post, transform, processing, errors } = useForm({
    customer_id: 0,
    terms: 0,
    tax_receipt: 0,
    lines: [],
    date: /* @__PURE__ */ new Date(),
    discount: defaultDiscount
  });
  useEffect(() => setCurrentItem(item), [item]);
  const findCurrentItem = useCallback(() => {
    const exists = (element) => element.id === (currentItem == null ? void 0 : currentItem.id);
    const index = invoiceForm.lines.findIndex(exists);
    if (index >= 0) {
      setEditing(true);
      const line = invoiceForm.lines[index];
      setCurrentItem(line);
      qtyInputRef.current.value = line.qty.toString();
      setAmount(line.amount);
    }
  }, [currentItem, invoiceForm.lines]);
  useEffect(() => {
    var _a;
    if (currentItem) {
      findCurrentItem();
      (_a = qtyInputRef.current) == null ? void 0 : _a.focus();
    }
  }, [currentItem, findCurrentItem]);
  useEffect(() => {
    const searchCustomer = () => {
      router.reload({ only: ["customers"], data: { search: dedbouncedSearch }, preserveUrl: true });
    };
    if (dedbouncedSearch) {
      searchCustomer();
    }
  }, [dedbouncedSearch]);
  const searchItem = (search2) => {
    router.reload({
      only: ["item"],
      data: { search: search2 },
      preserveUrl: true,
      onSuccess: () => {
        qtyInputRef.current.value = "1";
      }
    });
  };
  const handleOnSelectedItem = (item2) => {
    setCurrentItem(item2);
    referenceInputRef.current.value = item2.name;
    qtyInputRef.current.value = "1";
  };
  const handleOnKeyDown = (event) => {
    if (event.key === "Enter" || event.key === "Tab") {
      event.preventDefault();
      if (event.currentTarget.name === "reference" && isNotEmpty(event.currentTarget.value)) {
        searchItem(event.currentTarget.value);
        return;
      }
      if (event.currentTarget.name === "qty" && currentItem != void 0) {
        processCurrentItem();
      }
    }
  };
  const processCurrentItem = () => {
    var _a, _b;
    const line = currentItem;
    if (isEditing) {
      const index = invoiceForm.lines.findIndex((element) => element.id === line.id);
      if (index >= 0) {
        invoiceForm.lines[index].qty = ((_a = qtyInputRef.current) == null ? void 0 : _a.valueAsNumber) || 0;
        invoiceForm.lines[index].amount = amount;
      }
      setEditing(false);
    } else {
      invoiceForm.lines.push({ ...line, qty: ((_b = qtyInputRef.current) == null ? void 0 : _b.valueAsNumber) || 0, amount, action: "added" });
    }
    setInvoiceForm(() => {
      return { ...invoiceForm, lines: [...invoiceForm.lines] };
    });
    resetInvoiceFormInput();
  };
  const resetInvoiceFormInput = () => {
    var _a;
    setCurrentItem(void 0);
    setAmount(0);
    referenceInputRef.current.value = "";
    qtyInputRef.current.value = "";
    (_a = referenceInputRef.current) == null ? void 0 : _a.focus();
  };
  const handleCustomerSelection = (customer) => {
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, customer } };
    });
    setOpen(false);
  };
  const handleDateChange = (date) => {
    invoiceForm.header.date = date;
    invoiceForm.header.due = void 0;
    if (invoiceForm.header.terms > 1) {
      invoiceForm.header.due = addDays(invoiceForm.header.date, invoiceForm.header.terms);
    }
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, date } };
    });
  };
  const handlePaymentTermsChange = (value) => {
    invoiceForm.header.terms = Number(value);
    if (invoiceForm.header.terms > 1 && invoiceForm.header.date) {
      invoiceForm.header.due = addDays(invoiceForm.header.date, invoiceForm.header.terms);
    } else {
      invoiceForm.header.due = void 0;
    }
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, terms: Number(value) } };
    });
  };
  const handleTaxReceiptChange = (value) => {
    invoiceForm.header.taxReceipt = Number(value);
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, taxReceipt: Number(value) } };
    });
  };
  const handleRemoveLine = (event) => {
    event.preventDefault();
    const index = parseInt(event.currentTarget.dataset.index || "-1");
    if (index < 0) return;
    const newItems = invoiceForm.lines.filter((_, i) => i !== index);
    setInvoiceForm(() => {
      return { ...invoiceForm, lines: newItems };
    });
  };
  const handleDiscountTypeChange = (value) => {
    invoiceForm.header.discount.type = value;
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, discount: { ...invoiceForm.header.discount, type: value } } };
    });
  };
  const handleDiscountValueChange = (event) => {
    invoiceForm.header.discount.value = event.target.valueAsNumber;
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, discount: { ...invoiceForm.header.discount, value: event.target.valueAsNumber } } };
    });
  };
  const performInvoiceCancelation = (event) => {
    event.preventDefault();
    removeInvoiceForm();
    router.get("/invoices");
  };
  const handleCheckout = (event) => {
    event.preventDefault();
    if (computeTotalAmount() === 0) return;
    if (invoiceForm.header.terms === 1) {
      Object.keys(propsErrors).forEach((key) => delete propsErrors[key]);
      setCheckout(true);
      return;
    }
    placedInvoice();
  };
  const placedInvoice = () => {
    transform((data) => {
      var _a;
      return {
        ...data,
        customer_id: (_a = invoiceForm.header.customer) == null ? void 0 : _a.id,
        date: invoiceForm.header.date,
        terms: invoiceForm.header.terms,
        tax_receipt: invoiceForm.header.taxReceipt,
        discount: invoiceForm.header.discount,
        notes: invoiceForm.header.notes || "",
        lines: invoiceForm.lines.map((line) => {
          return { id: line.id, qty: line.qty, unit: line.unit.id, price: line.price, rate: line.tax.rate, action: line.action };
        }),
        payment: invoiceForm.payment
      };
    });
    post("/invoices", {
      ...headers,
      preserveState: "errors",
      onSuccess: () => {
        removeInvoiceForm();
        router.get("/invoices");
      }
    });
  };
  const computeDiscount = () => {
    const discount = invoiceForm.header.discount;
    if (discount.type === "percentage") {
      const total = composeSubTotal;
      return total * (discount.value / 100);
    }
    return discount.value;
  };
  const composeSubTotal = invoiceForm.lines.reduce((acc, line) => {
    return acc + line.amount;
  }, 0);
  const composeTax = invoiceForm.lines.reduce((acc, line) => {
    let discount = invoiceForm.header.discount.value;
    if (invoiceForm.header.discount.type === "fixed") {
      discount = discount / composeSubTotal * 100;
    }
    const lineAmount = line.price * line.qty;
    const lineDiscount = lineAmount * (discount / 100);
    const tax = (lineAmount - lineDiscount) * (line.tax.rate / 100);
    return acc + tax;
  }, 0);
  const computeTotalAmount = () => {
    const discount = computeDiscount();
    return composeSubTotal - discount + composeTax;
  };
  const handleCheckoutChange = (method, form) => {
    setInvoiceForm(() => {
      return { ...invoiceForm, payment: { ...invoiceForm.payment, [method]: form } };
    });
  };
  return /* @__PURE__ */ jsxs(AppLayout, { user: auth.user, breadcrumbs: createBreadcrumbs, children: [
    /* @__PURE__ */ jsx(AppLayout.Actions, { children: /* @__PURE__ */ jsxs("div", { className: "flex justify-end gap-x-6", children: [
      /* @__PURE__ */ jsx(Button, { variant: "secondary", onClick: () => setCancelConfirmation(true), children: t("global.actions.cancel") }),
      /* @__PURE__ */ jsx(Button, { onClick: handleCheckout, disabled: processing || computeTotalAmount() === 0, children: invoiceForm.header.terms === 1 ? t("global.actions.checkout") : t("global.actions.save") })
    ] }) }),
    /* @__PURE__ */ jsxs("div", { className: "grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4 bg-gray-50/10", children: [
      !openCheckout && propsErrors.status && /* @__PURE__ */ jsx("div", { className: "col-span-12", children: /* @__PURE__ */ jsx(AlertDestructive, { description: propsErrors.status, onDestroy: () => delete propsErrors.status }) }),
      /* @__PURE__ */ jsxs("div", { className: "z-50 col-span-12 grid min-h-42 grid-cols-2 gap-x-6", children: [
        /* @__PURE__ */ jsx(
          CustomerSection,
          {
            customer: invoiceForm.header.customer,
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
                    className: cn("w-[280px] justify-start text-left font-normal", !invoiceForm.header.date && "text-muted-foreground"),
                    children: [
                      /* @__PURE__ */ jsx(CalendarIcon, {}),
                      invoiceForm.header.date ? format(invoiceForm.header.date, "PPP") : /* @__PURE__ */ jsx("span", { children: t("global.datePlaceholder") })
                    ]
                  }
                ) }),
                /* @__PURE__ */ jsx(PopoverContent, { className: "w-auto p-0", children: /* @__PURE__ */ jsx(
                  Calendar,
                  {
                    mode: "single",
                    defaultMonth: invoiceForm.header.date,
                    selected: invoiceForm.header.date,
                    onSelect: handleDateChange,
                    initialFocus: true
                  }
                ) })
              ] }),
              /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.date })
            ] }),
            /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-y-2", children: [
              /* @__PURE__ */ jsx(Label, { htmlFor: "date", children: t("global.dueDate") }),
              /* @__PURE__ */ jsx(Label, { className: "text-muted-foreground w-70 rounded-sm border p-2.5", children: invoiceForm.header.due ? format(invoiceForm.header.due, "PPP") : t("global.noAvailable.default") })
            ] })
          ] }),
          /* @__PURE__ */ jsxs("div", { className: "col-span-6 flex flex-col gap-y-6", children: [
            /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-y-2", children: [
              /* @__PURE__ */ jsx(Label, { htmlFor: "paymentTerms", children: t("invoices.paymentTerms") }),
              /* @__PURE__ */ jsxs(
                Select,
                {
                  name: "paymentTerms",
                  onValueChange: handlePaymentTermsChange,
                  defaultValue: "0",
                  value: String(invoiceForm.header.terms),
                  required: true,
                  children: [
                    /* @__PURE__ */ jsx(SelectTrigger, { className: "w-full", children: /* @__PURE__ */ jsx(SelectValue, { placeholder: "Select terms" }) }),
                    /* @__PURE__ */ jsx(SelectContent, { className: "", children: paymentTerms.map((term, index) => /* @__PURE__ */ jsx(SelectItem, { value: term.value.toString(), children: term.label }, index.toString())) })
                  ]
                }
              ),
              /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.terms })
            ] }),
            /* @__PURE__ */ jsxs("div", { className: "flex flex-col gap-y-2", children: [
              /* @__PURE__ */ jsx(Label, { htmlFor: "paymentTerms", children: t("invoices.taxReceipt") }),
              /* @__PURE__ */ jsxs(
                Select,
                {
                  name: "paymentTerms",
                  onValueChange: handleTaxReceiptChange,
                  defaultValue: "0",
                  value: String(invoiceForm.header.taxReceipt),
                  required: true,
                  children: [
                    /* @__PURE__ */ jsx(SelectTrigger, { className: "w-full", children: /* @__PURE__ */ jsx(SelectValue, { placeholder: "Select terms" }) }),
                    /* @__PURE__ */ jsx(SelectContent, { className: "", children: tax_receipts.map((receipt) => /* @__PURE__ */ jsxs(SelectItem, { value: String(receipt.id), disabled: !receipt.available, children: [
                      receipt.name,
                      !receipt.available && /* @__PURE__ */ jsx("span", { className: "text-red-500", children: t("global.limitReached") })
                    ] }, receipt.id)) })
                  ]
                }
              ),
              /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.tax_receipt })
            ] })
          ] })
        ] })
      ] }),
      /* @__PURE__ */ jsx("div", { className: "col-span-12", children: /* @__PURE__ */ jsx("div", { className: "flex flex-col", children: /* @__PURE__ */ jsx(
        Lines,
        {
          items,
          lines: invoiceForm.lines,
          lineError: errors.lines,
          currentItem,
          handleRemoveLine,
          handleKeyDown: handleOnKeyDown,
          handleOnSelected: handleOnSelectedItem,
          amount,
          setAmount,
          referenceInputRef,
          qtyInputRef
        }
      ) }) }),
      /* @__PURE__ */ jsx("div", { className: "col-span-12 min-h-48", children: /* @__PURE__ */ jsx("div", { className: "flex flex-col gap-y-2", children: /* @__PURE__ */ jsxs("div", { className: "grid grid-cols-12", children: [
        /* @__PURE__ */ jsxs("div", { className: "col-span-10 flex flex-col gap-y-2 p-2", children: [
          /* @__PURE__ */ jsx(Label, { className: "text-sm/6 font-medium", children: t("global.notes") }),
          /* @__PURE__ */ jsx(
            Textarea,
            {
              name: "notes",
              rows: 4,
              className: "focus:no-data-focus:outline-none block w-1/2 resize-none rounded-lg border px-3 py-1.5 text-sm/6 data-focus:outline-2 data-focus:-outline-offset-2 data-focus:outline-white/25",
              defaultValue: invoiceForm.header.notes,
              onChange: (e) => setInvoiceForm(() => {
                return { ...invoiceForm, header: { ...invoiceForm.header, notes: e.currentTarget.value } };
              })
            }
          )
        ] }),
        /* @__PURE__ */ jsx("div", { className: "col-span-2 flex flex-col gap-y-2 rounded-lg border border-gray-300/25 bg-gray-100/10", children: /* @__PURE__ */ jsxs("div", { className: "grid place-content-end gap-y-4 p-2", children: [
          /* @__PURE__ */ jsx(InputError, { message: errors["discount.value"] }),
          /* @__PURE__ */ jsxs("div", { className: "flex w-60 items-center justify-between", children: [
            /* @__PURE__ */ jsx("span", { className: "block text-base", children: t("global.subTotal") }),
            /* @__PURE__ */ jsx("span", { className: "block text-base", children: currency(composeSubTotal) })
          ] }),
          /* @__PURE__ */ jsxs("div", { className: "flex w-60 items-center justify-between", children: [
            /* @__PURE__ */ jsxs("span", { className: "block text-base", children: [
              t("global.discount"),
              invoiceForm.header.discount.type === "percentage" && /* @__PURE__ */ jsxs(Fragment, { children: [
                ": ",
                /* @__PURE__ */ jsx("span", { className: "text-muted-foreground text-xs", children: currency(computeDiscount()) })
              ] })
            ] }),
            /* @__PURE__ */ jsxs("div", { className: "flex w-40 justify-end", children: [
              /* @__PURE__ */ jsx(
                Input,
                {
                  type: "number",
                  min: 0,
                  defaultValue: invoiceForm.header.discount.value,
                  name: "discount",
                  className: "w-16 text-end",
                  onChange: handleDiscountValueChange
                }
              ),
              /* @__PURE__ */ jsxs(
                Select,
                {
                  name: "discountType",
                  onValueChange: handleDiscountTypeChange,
                  defaultValue: "percentage",
                  value: String(invoiceForm.header.discount.type),
                  required: true,
                  children: [
                    /* @__PURE__ */ jsx(SelectTrigger, { className: "w-16", children: /* @__PURE__ */ jsx(SelectValue, { placeholder: t("global.discount") }) }),
                    /* @__PURE__ */ jsxs(SelectContent, { className: "", children: [
                      /* @__PURE__ */ jsx(SelectItem, { value: "fixed", children: "$" }),
                      /* @__PURE__ */ jsx(SelectItem, { value: "percentage", children: "%" })
                    ] })
                  ]
                }
              )
            ] })
          ] }),
          /* @__PURE__ */ jsxs("div", { className: "flex w-60 items-center justify-between", children: [
            /* @__PURE__ */ jsx("span", { className: "block text-base", children: t("global.tax") }),
            /* @__PURE__ */ jsx("span", { className: "block text-base", children: currency(composeTax) })
          ] }),
          /* @__PURE__ */ jsx(Separator, {}),
          /* @__PURE__ */ jsxs("div", { className: "flex w-60 items-center justify-between", children: [
            /* @__PURE__ */ jsx("span", { className: "block text-xl", children: t("global.total") }),
            /* @__PURE__ */ jsx("span", { className: "block text-xl", children: currency(computeTotalAmount()) })
          ] })
        ] }) })
      ] }) }) }),
      /* @__PURE__ */ jsx(AlertDialog, { open: openCancelConfirmation, onOpenChange: setCancelConfirmation, children: /* @__PURE__ */ jsxs(AlertDialogContent, { children: [
        /* @__PURE__ */ jsxs(AlertDialogHeader, { children: [
          /* @__PURE__ */ jsx(AlertDialogTitle, { children: t("invoices.confirmsCancelation.title") }),
          /* @__PURE__ */ jsx(AlertDialogDescription, { children: t("invoices.confirmsCancelation.description") })
        ] }),
        /* @__PURE__ */ jsxs(AlertDialogFooter, { children: [
          /* @__PURE__ */ jsx(AlertDialogCancel, { children: t("global.cancel") }),
          /* @__PURE__ */ jsx(AlertDialogAction, { onClick: performInvoiceCancelation, children: t("invoices.confirmsCancelation.confirm") })
        ] })
      ] }) }),
      /* @__PURE__ */ jsx(
        CheckoutForm,
        {
          action: t("global.actions.save"),
          openCheckout,
          setCheckout,
          paymentForm: invoiceForm.payment,
          totalAmount: computeTotalAmount(),
          onCompleteCheckout: placedInvoice,
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
