import { jsx, jsxs, Fragment } from "react/jsx-runtime";
import { C as Checkbox, T as TextCell, H as HeaderCell, D as DateCell, a as CurrencyCell } from "./checkbox.js";
import { L as LinkCell } from "./link-cell.js";
import { S as StatusBadge } from "./status-badge.js";
import { B as Button } from "./button.js";
import { D as DropdownMenu, a as DropdownMenuTrigger, b as DropdownMenuContent, d as DropdownMenuLabel, e as DropdownMenuSeparator, f as DropdownMenuItem } from "./dropdown-menu.js";
import { ArrowUpDown, MoreHorizontal } from "lucide-react";
import "./use-number.js";
import "date-fns";
import "@radix-ui/react-checkbox";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "@inertiajs/react";
import "./use-translation.js";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "@radix-ui/react-dropdown-menu";
const getColumns = ({ onDidClick, t }) => {
  return [
    {
      id: "select",
      header: ({ table }) => /* @__PURE__ */ jsx(
        Checkbox,
        {
          checked: table.getIsAllPageRowsSelected() || table.getIsSomePageRowsSelected() && "indeterminate",
          onCheckedChange: (value) => table.toggleAllPageRowsSelected(!!value),
          "aria-label": "Select all"
        }
      ),
      cell: ({ row }) => /* @__PURE__ */ jsx(Checkbox, { checked: row.getIsSelected(), onCheckedChange: (value) => row.toggleSelected(!!value), "aria-label": "Select row" }),
      enableSorting: false,
      enableHiding: false
    },
    {
      accessorKey: "number",
      meta: t("global.number"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.number"), alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "ncf",
      meta: "NCF",
      enableHiding: true,
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: "NCF", alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "customer.name",
      id: "customer.name",
      meta: t("global.customer"),
      size: 200,
      header: ({ column }) => {
        return /* @__PURE__ */ jsxs(Button, { className: "font-semibold uppercase", variant: "ghost", onClick: () => column.toggleSorting(column.getIsSorted() === "asc"), children: [
          t("global.customer"),
          " ",
          /* @__PURE__ */ jsx(ArrowUpDown, {})
        ] });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(
          LinkCell,
          {
            href: `/customers?id=${props.row.original.customer.uuid}`,
            columnWidth: props.column.getSize(),
            value: props.getValue()
          }
        );
      }
    },
    {
      accessorKey: "date",
      meta: t("global.date"),
      size: 100,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.date"), alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(DateCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "amount",
      meta: t("global.amount"),
      size: 100,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.amount"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "discount",
      meta: t("global.discount"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.discount"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        const discount = props.row.getValue("discount");
        const suffix = discount.type === "percentage" ? "%" : void 0;
        return /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), suffix, value: String(discount.value) });
      }
    },
    {
      accessorKey: "tax",
      meta: t("global.tax"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.tax"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "total",
      meta: t("global.total"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.total"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "amount_due",
      meta: t("global.balance"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.balance"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "status",
      size: 70,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.status"), alignment: "center", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(StatusBadge, { type: "invoice", status: props.row.original.status });
      }
    },
    {
      accessorKey: "paid_status",
      size: 70,
      meta: t("invoices.paidStatus"),
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("invoices.paidStatus"), alignment: "center", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(StatusBadge, { type: "paid", status: props.row.original.paid_status });
      }
    },
    {
      id: "actions",
      enableHiding: false,
      cell: (props) => {
        const disabled = props.row.original.status === "void";
        const canRecordPayment = props.row.original.paid_status === "unpaid" || props.row.original.paid_status === "partial";
        return /* @__PURE__ */ jsxs(DropdownMenu, { children: [
          /* @__PURE__ */ jsx(DropdownMenuTrigger, { asChild: true, children: /* @__PURE__ */ jsxs(Button, { variant: "ghost", className: "size-8 p-0", children: [
            /* @__PURE__ */ jsx("span", { className: "sr-only", children: t("global.openMenu") }),
            /* @__PURE__ */ jsx(MoreHorizontal, {})
          ] }) }),
          /* @__PURE__ */ jsxs(DropdownMenuContent, { align: "end", children: [
            /* @__PURE__ */ jsx(DropdownMenuLabel, { children: t("global.actions.title") }),
            /* @__PURE__ */ jsx(DropdownMenuSeparator, {}),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "view"), children: t("invoices.viewInvoice.title") }),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "edit"), disabled, children: t("invoices.editInvoice.title") }),
            /* @__PURE__ */ jsx(DropdownMenuSeparator, {}),
            canRecordPayment && /* @__PURE__ */ jsxs(Fragment, { children: [
              /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "record-payment"), children: t("global.actions.recordPayment") }),
              /* @__PURE__ */ jsx(DropdownMenuSeparator, {})
            ] }),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "void"), disabled, children: t("invoices.voidInvoice.title") })
          ] })
        ] });
      }
    }
  ];
};
export {
  getColumns
};
