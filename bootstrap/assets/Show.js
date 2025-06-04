import { jsxs, jsx, Fragment } from "react/jsx-runtime";
import { S as StatusBadge } from "./status-badge.js";
import { L as Label } from "./label.js";
import { S as Select, a as SelectTrigger, b as SelectValue, c as SelectContent, d as SelectItem } from "./select.js";
import { S as Separator } from "./separator.js";
import { u as useNumber } from "./use-number.js";
import { u as useTranslation } from "./use-translation.js";
import { c as cn, i as isNotEmpty } from "./utils.js";
import { P as PaidStatuses } from "./index13.js";
import { format } from "date-fns";
import { UserPen, Calendar1, CircleDollarSignIcon, CreditCardIcon, CircleCheckIcon } from "lucide-react";
import PaymentSummary from "./payment-summary.js";
import "@radix-ui/react-label";
import "@radix-ui/react-select";
import "@radix-ui/react-separator";
import "@inertiajs/react";
import "clsx";
import "tailwind-merge";
function Show({ invoice, auth }) {
  const t = useTranslation().trans;
  const { currency } = useNumber();
  return /* @__PURE__ */ jsxs("div", { className: "grid grid-cols-12 gap-x-4", children: [
    /* @__PURE__ */ jsxs("div", { className: "col-span-12 pb-6", children: [
      /* @__PURE__ */ jsx(Separator, {}),
      /* @__PURE__ */ jsxs("div", { className: "flex justify-between py-6 [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-medium", children: [
        /* @__PURE__ */ jsxs("div", { className: "col-span-6 flex items-center gap-x-6 [&>div]:flex [&>div]:gap-x-2", children: [
          /* @__PURE__ */ jsxs("div", { children: [
            /* @__PURE__ */ jsx(Label, { children: t("invoices.single.title") }),
            /* @__PURE__ */ jsxs(Label, { children: [
              "#",
              invoice.header.number
            ] })
          ] }),
          /* @__PURE__ */ jsxs("div", { children: [
            /* @__PURE__ */ jsx(Label, { children: "NCF" }),
            /* @__PURE__ */ jsx(Label, { children: invoice.header.ncf })
          ] })
        ] }),
        /* @__PURE__ */ jsxs("div", { className: "col-span-6 flex items-center gap-x-2 [&_[data-slot=label]]:font-normal", children: [
          /* @__PURE__ */ jsx(Label, { children: t("global.date") }),
          /* @__PURE__ */ jsx(Label, { className: "", children: format(invoice.header.date, "dd-MM-yyyy") })
        ] })
      ] }),
      /* @__PURE__ */ jsx(Separator, {})
    ] }),
    /* @__PURE__ */ jsxs("div", { className: "col-span-8 grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4", children: [
      /* @__PURE__ */ jsxs("div", { className: "col-span-12 grid grid-cols-12", children: [
        /* @__PURE__ */ jsx("div", { className: "col-span-6", children: /* @__PURE__ */ jsxs("div", { className: "w-56", children: [
          /* @__PURE__ */ jsxs(Label, { className: "font-bold", children: [
            t("invoices.single.payTo"),
            ":"
          ] }),
          /* @__PURE__ */ jsxs("div", { className: "pt-2", children: [
            /* @__PURE__ */ jsx("span", { className: "text-sm font-semibold", children: auth.company.name }),
            /* @__PURE__ */ jsx("address", { className: "text-muted-foreground text-sm font-normal", children: auth.company.address })
          ] }),
          /* @__PURE__ */ jsx("div", { className: "text-sm font-medium", children: invoice.header.customer.email })
        ] }) }),
        /* @__PURE__ */ jsx("div", { className: "col-span-6 place-items-end", children: /* @__PURE__ */ jsxs("div", { className: "w-56", children: [
          /* @__PURE__ */ jsxs(Label, { className: "font-bold", children: [
            t("invoices.single.invoiceTo"),
            ":"
          ] }),
          /* @__PURE__ */ jsxs("div", { className: "pt-2", children: [
            /* @__PURE__ */ jsx("span", { className: "text-sm font-semibold", children: invoice.header.customer.name }),
            /* @__PURE__ */ jsx("address", { className: "text-muted-foreground text-sm font-normal", children: invoice.header.customer.address })
          ] }),
          /* @__PURE__ */ jsx("div", { className: "text-sm font-medium", children: invoice.header.customer.email })
        ] }) })
      ] }),
      /* @__PURE__ */ jsx("div", { className: "col-span-12", children: /* @__PURE__ */ jsxs(
        "table",
        {
          className: cn(
            "w-full table-auto",
            "[&_th]:border-none [&_th]:border-gray-200 [&_th]:bg-gray-50/25 [&_th]:text-sm [&_th]:font-semibold [&_th]:uppercase",
            "[&_th]:p-2 [&_th]:text-start [&_th[data-format=number]]:text-end",
            "[&_td]:border-y [&_td]:p-2 [&_td]:text-start [&_td[data-format=number]]:w-36 [&_td[data-format=number]]:text-end"
          ),
          children: [
            /* @__PURE__ */ jsx("thead", { children: /* @__PURE__ */ jsxs("tr", { className: "bg-gray-50/50", children: [
              /* @__PURE__ */ jsx("th", { scope: "col", className: "text-start", children: "ID" }),
              /* @__PURE__ */ jsx("th", { scope: "col", className: "text-start", children: t("global.name") }),
              /* @__PURE__ */ jsx("th", { scope: "col", className: "w-20 text-start", children: t("global.unit") }),
              /* @__PURE__ */ jsx("th", { scope: "col", "data-format": "number", className: "!w-20", children: t("global.qty") }),
              /* @__PURE__ */ jsx("th", { scope: "col", "data-format": "number", children: t("global.price") }),
              /* @__PURE__ */ jsx("th", { scope: "col", "data-format": "number", children: t("global.tax") }),
              /* @__PURE__ */ jsx("th", { scope: "col", "data-format": "number", children: t("global.amount") })
            ] }) }),
            /* @__PURE__ */ jsx("tbody", { children: invoice.lines.map((line) => /* @__PURE__ */ jsxs("tr", { children: [
              /* @__PURE__ */ jsx("td", { children: line.id }),
              /* @__PURE__ */ jsx("td", { children: line.name }),
              /* @__PURE__ */ jsx("td", { children: line.unit.name }),
              /* @__PURE__ */ jsx("td", { "data-format": "number", className: "!w-16", children: line.qty }),
              /* @__PURE__ */ jsx("td", { "data-format": "number", children: currency(line.price) }),
              /* @__PURE__ */ jsx("td", { "data-format": "number", children: currency(line.tax.amount) }),
              /* @__PURE__ */ jsx("td", { "data-format": "number", children: currency(line.amount) })
            ] }, line.id)) })
          ]
        }
      ) }),
      /* @__PURE__ */ jsxs("div", { className: "col-span-12 grid grid-cols-12", children: [
        /* @__PURE__ */ jsx("div", { className: "col-span-7", children: /* @__PURE__ */ jsxs("div", { className: "max-w-sm rounded-md border p-4", children: [
          /* @__PURE__ */ jsxs(Label, { className: "text-sm/6 font-medium", children: [
            t("global.notes"),
            ":"
          ] }),
          /* @__PURE__ */ jsx("div", { className: "text-muted-foreground text-sm", children: isNotEmpty(invoice.header.notes) ? invoice.header.notes : t("global.noNotesWasLeft") })
        ] }) }),
        /* @__PURE__ */ jsxs("div", { className: "col-span-5 rounded-md border p-4", children: [
          /* @__PURE__ */ jsx(Label, { children: t("invoices.single.totalSummary") }),
          /* @__PURE__ */ jsxs(
            "div",
            {
              className: cn(
                "flex flex-col gap-y-3 py-4",
                "[&_[data-slot=label]]:text-muted-foreground [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-normal",
                "[&_[data-slot=label-value]]:text-lg [&_[data-slot=label-value]]:font-medium"
              ),
              children: [
                /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-between", children: [
                  /* @__PURE__ */ jsx(Label, { children: t("global.subTotal") }),
                  /* @__PURE__ */ jsx(Label, { "data-slot": "label-value", children: currency(invoice.header.amount) })
                ] }),
                /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-between", children: [
                  /* @__PURE__ */ jsx(Label, { children: t("global.discount") }),
                  /* @__PURE__ */ jsxs(Label, { "data-slot": "label-value", children: [
                    invoice.header.discount.type === "percentage" && /* @__PURE__ */ jsx("span", { className: "text-muted-foreground text-xs", children: currency(invoice.header.amount * (invoice.header.discount.value / 100)) }),
                    invoice.header.discount.type === "fixed" ? currency(invoice.header.discount.value) : `${invoice.header.discount.value}%`
                  ] })
                ] }),
                /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-between", children: [
                  /* @__PURE__ */ jsx(Label, { children: t("global.tax") }),
                  /* @__PURE__ */ jsx(Label, { "data-slot": "label-value", children: currency(invoice.header.tax) })
                ] }),
                /* @__PURE__ */ jsx(Separator, {}),
                /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-between", children: [
                  /* @__PURE__ */ jsx(Label, { children: t("global.total") }),
                  /* @__PURE__ */ jsx(Label, { "data-slot": "label-value", children: currency(invoice.header.total) })
                ] }),
                invoice.header.due_on && /* @__PURE__ */ jsxs(Fragment, { children: [
                  /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-between", children: [
                    /* @__PURE__ */ jsx(Label, { children: t("global.totalPaymentApplied") }),
                    /* @__PURE__ */ jsx(Label, { "data-slot": "label-value", children: currency(invoice.header.total - invoice.header.amount_due) })
                  ] }),
                  /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-between", children: [
                    /* @__PURE__ */ jsx(Label, { children: t("global.balance") }),
                    /* @__PURE__ */ jsx(Label, { "data-slot": "label-value", children: currency(invoice.header.amount_due) })
                  ] })
                ] })
              ]
            }
          )
        ] })
      ] })
    ] }),
    /* @__PURE__ */ jsxs("div", { className: "col-span-4 flex flex-col gap-y-3 rounded-lg border p-3", children: [
      /* @__PURE__ */ jsx(StatusBadge, { type: "paid", variant: "alert", prefix: `${t("invoices.paidStatus")}:`, status: invoice.header.paid_status }),
      /* @__PURE__ */ jsx(Label, { children: t("invoices.single.description") }),
      /* @__PURE__ */ jsx(Separator, {}),
      /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-between", children: [
        /* @__PURE__ */ jsx(Label, { className: "text-lg", children: currency(invoice.header.total) }),
        /* @__PURE__ */ jsxs(Select, { name: "paid_status", defaultValue: "0", value: invoice.header.paid_status, required: true, disabled: invoice.header.status === "void", children: [
          /* @__PURE__ */ jsx(SelectTrigger, { className: "w-46", children: /* @__PURE__ */ jsx(SelectValue, { placeholder: "Paid status" }) }),
          /* @__PURE__ */ jsx(SelectContent, { className: "w-46", children: PaidStatuses.map((status) => /* @__PURE__ */ jsx(SelectItem, { value: status, children: t(`global.paidStatuses.${status}`) }, status)) })
        ] })
      ] }),
      /* @__PURE__ */ jsx(Separator, {}),
      /* @__PURE__ */ jsxs("div", { className: "flex items-center gap-x-1 text-sm", children: [
        /* @__PURE__ */ jsx(UserPen, { size: 14 }),
        /* @__PURE__ */ jsxs("span", { className: "font-medium", children: [
          t("global.createdBy"),
          ":"
        ] }),
        /* @__PURE__ */ jsx("span", { className: "text-muted-foreground", children: "Jane Doe" })
      ] }),
      invoice.header.due_on && /* @__PURE__ */ jsxs("div", { className: "flex items-center gap-x-1 text-sm", children: [
        /* @__PURE__ */ jsx(Calendar1, { size: 14 }),
        /* @__PURE__ */ jsxs("span", { className: "font-medium", children: [
          t("global.dueDate"),
          ":"
        ] }),
        /* @__PURE__ */ jsx("span", { className: "text-muted-foreground", children: format(invoice.header.due_on, "PPP") })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "flex items-center gap-x-1 text-sm", children: [
        /* @__PURE__ */ jsx(CircleDollarSignIcon, { size: 14 }),
        /* @__PURE__ */ jsxs("span", { className: "font-medium", children: [
          t("global.currency"),
          ":"
        ] }),
        /* @__PURE__ */ jsx("span", { className: "text-muted-foreground", children: "Domincan Peso" })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "flex items-center gap-x-1 text-sm", children: [
        /* @__PURE__ */ jsx(CreditCardIcon, { size: 14 }),
        /* @__PURE__ */ jsx("span", { className: "font-medium", children: t("global.paymentSummary") })
      ] }),
      invoice.header.due_on !== null && /* @__PURE__ */ jsx("span", { className: "text-muted-foreground -m-1.5 block px-1.5 text-sm", children: t("global.noAvailable.yet") }),
      /* @__PURE__ */ jsx(PaymentSummary, { paymentData: invoice.header.payment }),
      /* @__PURE__ */ jsx(Separator, {}),
      /* @__PURE__ */ jsx("div", { className: "px-1 py-4", children: /* @__PURE__ */ jsxs(
        "ol",
        {
          className: cn(
            "relative flex flex-col justify-between gap-y-2 border-s border-gray-200",
            "[&_li]:ms-4 [&_li]:flex [&_li]:items-center [&_li]:justify-between [&_li]:text-sm",
            "[&_li>div]:flex [&_li>div]:items-center [&_li>div]:gap-x-1 [&_svg]:absolute [&_svg]:-start-2.5 [&_svg]:size-5",
            "[&_svg]:fill-green-400 [&_svg]:stroke-white",
            "[&_[data-status=pending]]:text-muted-foreground [&_[data-status=pending]]:fill-gray-200",
            "[&_[data-slot=label]]:text-muted-foreground [&_[data-slot=label]]:text-sm/1 [&_[data-slot=label]]:font-normal"
          ),
          children: [
            /* @__PURE__ */ jsxs("li", { children: [
              /* @__PURE__ */ jsxs("div", { children: [
                /* @__PURE__ */ jsx(CircleCheckIcon, {}),
                /* @__PURE__ */ jsx("div", { children: "Invoice created" })
              ] }),
              /* @__PURE__ */ jsx(Label, { children: format(invoice.header.date, "P") })
            ] }),
            /* @__PURE__ */ jsxs("li", { children: [
              /* @__PURE__ */ jsxs("div", { children: [
                /* @__PURE__ */ jsx(CircleCheckIcon, {}),
                /* @__PURE__ */ jsx("div", { children: "Invoice sent" })
              ] }),
              /* @__PURE__ */ jsx(Label, { children: format(invoice.header.date, "P") })
            ] }),
            /* @__PURE__ */ jsxs("li", { children: [
              /* @__PURE__ */ jsxs("div", { children: [
                /* @__PURE__ */ jsx(CircleCheckIcon, { "data-status": "pending" }),
                /* @__PURE__ */ jsx("div", { children: "Invoice paid" })
              ] }),
              /* @__PURE__ */ jsx(Label, { children: format(invoice.header.date, "P") })
            ] })
          ]
        }
      ) })
    ] })
  ] });
}
export {
  Show as default
};
