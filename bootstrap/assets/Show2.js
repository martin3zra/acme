import { jsxs, jsx } from "react/jsx-runtime";
import { S as StatusBadge } from "./status-badge.js";
import { L as Label } from "./label.js";
import { S as Separator } from "./separator.js";
import { u as useNumber } from "./use-number.js";
import { u as useTranslation } from "./use-translation.js";
import { c as cn, i as isNotEmpty } from "./utils.js";
import { Link } from "@inertiajs/react";
import { formatDate, format } from "date-fns/format";
import { Eye, UserPen, CircleDollarSignIcon, CreditCardIcon, CircleCheckIcon } from "lucide-react";
import PaymentSummary from "./payment-summary.js";
import "@radix-ui/react-label";
import "@radix-ui/react-separator";
import "clsx";
import "tailwind-merge";
function Show({ payment, auth }) {
  const t = useTranslation().trans;
  const { currency } = useNumber();
  return /* @__PURE__ */ jsxs("div", { className: "grid grid-cols-12 gap-x-4", children: [
    /* @__PURE__ */ jsxs("div", { className: "col-span-12 pb-6", children: [
      /* @__PURE__ */ jsx(Separator, {}),
      /* @__PURE__ */ jsxs("div", { className: "flex justify-between py-6 [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-medium", children: [
        /* @__PURE__ */ jsx("div", { className: "col-span-6 flex items-center gap-x-6 [&>div]:flex [&>div]:gap-x-2", children: /* @__PURE__ */ jsxs("div", { children: [
          /* @__PURE__ */ jsx(Label, { children: t("payments.single.title") }),
          /* @__PURE__ */ jsxs(Label, { children: [
            "#",
            payment.header.number
          ] })
        ] }) }),
        /* @__PURE__ */ jsxs("div", { className: "col-span-6 flex items-center gap-x-2 [&_[data-slot=label]]:font-normal", children: [
          /* @__PURE__ */ jsx(Label, { children: t("global.date") }),
          /* @__PURE__ */ jsx(Label, { className: "", children: formatDate(payment.header.date, "dd-MM-yyyy") })
        ] })
      ] }),
      /* @__PURE__ */ jsx(Separator, {})
    ] }),
    /* @__PURE__ */ jsxs("div", { className: "col-span-9 grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4", children: [
      /* @__PURE__ */ jsxs("div", { className: "col-span-12 grid grid-cols-12", children: [
        /* @__PURE__ */ jsx("div", { className: "col-span-6", children: /* @__PURE__ */ jsxs("div", { className: "w-56", children: [
          /* @__PURE__ */ jsxs(Label, { className: "font-bold", children: [
            t("payments.single.payTo"),
            ":"
          ] }),
          /* @__PURE__ */ jsxs("div", { className: "pt-2", children: [
            /* @__PURE__ */ jsx("span", { className: "text-sm font-semibold", children: auth.company.name }),
            /* @__PURE__ */ jsx("address", { className: "text-muted-foreground text-sm font-normal", children: auth.company.address })
          ] }),
          /* @__PURE__ */ jsx("div", { className: "text-sm font-medium", children: payment.header.customer.email })
        ] }) }),
        /* @__PURE__ */ jsx("div", { className: "col-span-6 place-items-end", children: /* @__PURE__ */ jsxs("div", { className: "w-56", children: [
          /* @__PURE__ */ jsxs(Label, { className: "font-bold", children: [
            t("payments.single.paymentFrom"),
            ":"
          ] }),
          /* @__PURE__ */ jsxs("div", { className: "pt-2", children: [
            /* @__PURE__ */ jsx("span", { className: "text-sm font-semibold", children: payment.header.customer.name }),
            /* @__PURE__ */ jsx("address", { className: "text-muted-foreground text-sm font-normal", children: payment.header.customer.address })
          ] }),
          /* @__PURE__ */ jsx("div", { className: "text-sm font-medium", children: payment.header.customer.email })
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
              /* @__PURE__ */ jsx("th", { scope: "col", className: "text-start", children: t("payments.single.invoice") }),
              /* @__PURE__ */ jsx("th", { scope: "col", className: "text-start", children: t("global.date") }),
              /* @__PURE__ */ jsx("th", { scope: "col", "data-format": "number", className: "!min-w-20", children: t("global.amount") }),
              /* @__PURE__ */ jsx("th", { scope: "col", "data-format": "number", className: "!min-w-20", children: t("global.balance") }),
              /* @__PURE__ */ jsx("th", { scope: "col", className: "text-start", children: t("global.payDate") }),
              /* @__PURE__ */ jsx("th", { scope: "col", "data-format": "number", className: "!min-w-20", children: t("global.payment") }),
              /* @__PURE__ */ jsx("th", { scope: "col", className: "text-start", children: t("payments.paidStatus") })
            ] }) }),
            /* @__PURE__ */ jsx("tbody", { children: payment.lines.map((line) => /* @__PURE__ */ jsxs("tr", { children: [
              /* @__PURE__ */ jsx("td", { children: line.invoice.number }),
              /* @__PURE__ */ jsx("td", { children: formatDate(line.invoice.date, "dd-MM-yyyy") }),
              /* @__PURE__ */ jsx("td", { "data-format": "number", className: "!min-w-20", children: currency(line.invoice.amount) }),
              /* @__PURE__ */ jsx("td", { "data-format": "number", className: "!min-w-20", children: currency(line.invoice.amount_due) }),
              /* @__PURE__ */ jsx("td", { children: formatDate(payment.header.date, "dd-MM-yyyy") }),
              /* @__PURE__ */ jsxs("td", { "data-format": "number", children: [
                " ",
                currency(line.payment)
              ] }),
              /* @__PURE__ */ jsxs("td", { className: "flex items-center gap-x-1", children: [
                /* @__PURE__ */ jsx(StatusBadge, { type: "paid", status: line.invoice.paid_status }),
                /* @__PURE__ */ jsx(Link, { href: `/invoices?id=${line.invoice.uuid}`, children: /* @__PURE__ */ jsx(Eye, { className: "text-muted-foreground" }) })
              ] })
            ] }, line.id)) })
          ]
        }
      ) }),
      /* @__PURE__ */ jsxs("div", { className: "col-span-12 grid grid-cols-12", children: [
        /* @__PURE__ */ jsx("div", { className: "col-span-8", children: /* @__PURE__ */ jsxs("div", { className: "max-w-sm rounded-md border p-4", children: [
          /* @__PURE__ */ jsxs(Label, { className: "text-sm/6 font-medium", children: [
            t("global.notes"),
            ":"
          ] }),
          /* @__PURE__ */ jsx("div", { className: "text-muted-foreground text-sm", children: isNotEmpty(payment.header.notes) ? payment.header.notes : "No notes left" })
        ] }) }),
        /* @__PURE__ */ jsxs("div", { className: "col-span-4 rounded-md border p-4", children: [
          /* @__PURE__ */ jsx(Label, { children: t("payments.single.totalSummary") }),
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
                  /* @__PURE__ */ jsx(Label, { "data-slot": "label-value", children: currency(payment.header.amount) })
                ] }),
                /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-between", children: [
                  /* @__PURE__ */ jsx(Label, { children: t("global.discount") }),
                  /* @__PURE__ */ jsx(Label, { "data-slot": "label-value", children: currency(0) })
                ] }),
                /* @__PURE__ */ jsx(Separator, {}),
                /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-between", children: [
                  /* @__PURE__ */ jsx(Label, { children: t("global.total") }),
                  /* @__PURE__ */ jsx(Label, { "data-slot": "label-value", children: currency(payment.header.amount) })
                ] })
              ]
            }
          )
        ] })
      ] })
    ] }),
    /* @__PURE__ */ jsxs("div", { className: "col-span-3 flex flex-col gap-y-3 rounded-lg border p-3", children: [
      /* @__PURE__ */ jsx(Label, { children: t("payments.single.description") }),
      /* @__PURE__ */ jsx(Separator, {}),
      /* @__PURE__ */ jsxs("div", { className: "flex items-center gap-x-1 text-sm", children: [
        /* @__PURE__ */ jsx(UserPen, { size: 14 }),
        /* @__PURE__ */ jsxs("span", { className: "font-medium", children: [
          t("global.createdBy"),
          ":"
        ] }),
        /* @__PURE__ */ jsx("span", { className: "text-muted-foreground", children: "Jane Doe" })
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
      /* @__PURE__ */ jsx(PaymentSummary, { paymentData: payment.header.payment }),
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
                /* @__PURE__ */ jsx("div", { children: "payment created" })
              ] }),
              /* @__PURE__ */ jsx(Label, { children: format(payment.header.date, "P") })
            ] }),
            /* @__PURE__ */ jsxs("li", { children: [
              /* @__PURE__ */ jsxs("div", { children: [
                /* @__PURE__ */ jsx(CircleCheckIcon, {}),
                /* @__PURE__ */ jsx("div", { children: "payment sent" })
              ] }),
              /* @__PURE__ */ jsx(Label, { children: format(payment.header.date, "P") })
            ] }),
            /* @__PURE__ */ jsxs("li", { children: [
              /* @__PURE__ */ jsxs("div", { children: [
                /* @__PURE__ */ jsx(CircleCheckIcon, { "data-status": "pending" }),
                /* @__PURE__ */ jsx("div", { children: "payment paid" })
              ] }),
              /* @__PURE__ */ jsx(Label, { children: format(payment.header.date, "P") })
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
