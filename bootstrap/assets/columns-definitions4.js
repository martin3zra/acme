import { jsx, jsxs, Fragment } from "react/jsx-runtime";
import { C as Checkbox, T as TextCell, H as HeaderCell, a as CurrencyCell, D as DateCell } from "./checkbox.js";
import { L as LinkCell } from "./link-cell.js";
import { S as StatusBadge } from "./status-badge.js";
import { B as Button } from "./button.js";
import { D as DropdownMenu, a as DropdownMenuTrigger, b as DropdownMenuContent, d as DropdownMenuLabel, e as DropdownMenuSeparator, f as DropdownMenuItem } from "./dropdown-menu.js";
import { T as TooltipProvider, a as Tooltip, b as TooltipTrigger, c as TooltipContent } from "./tooltip.js";
import { MessageCircleMore, ArrowUpDown, MoreHorizontal } from "lucide-react";
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
import "@radix-ui/react-tooltip";
const NumericCell = ({ value, columnWidth }) => {
  return /* @__PURE__ */ jsx("div", { className: "p-1.5 text-end", title: value, style: { width: columnWidth }, children: value });
};
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
        const hasNotes = !!props.row.original.notes;
        return /* @__PURE__ */ jsxs("div", { className: "[&_[data-slot=has-notes]]:-px-6 relative flex [&_[data-slot=has-notes]]:block [&_[data-slot=has-notes]]:text-red-500", children: [
          /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() }),
          /* @__PURE__ */ jsx(TooltipProvider, { children: /* @__PURE__ */ jsxs(Tooltip, { children: [
            /* @__PURE__ */ jsx(TooltipTrigger, { children: /* @__PURE__ */ jsx(
              MessageCircleMore,
              {
                className: "absolute inset-0 -top-0 left-[62%] hidden size-5 -translate-x-1/2 -translate-y-1/2 transform cursor-pointer",
                "data-slot": hasNotes ? "has-notes" : "default"
              }
            ) }),
            /* @__PURE__ */ jsx(TooltipContent, { children: props.row.original.notes })
          ] }) })
        ] });
      }
    },
    {
      accessorKey: "customer.name",
      id: "customer.name",
      meta: t("global.customer"),
      header: ({ column }) => {
        return /* @__PURE__ */ jsxs(Button, { className: "uppercase", variant: "ghost", onClick: () => column.toggleSorting(column.getIsSorted() === "asc"), children: [
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
      accessorKey: "customer.amount_due",
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
      accessorKey: "date",
      meta: t("global.date"),
      // size: 880,
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
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.amount"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "invoices",
      meta: t("global.navMain.invoices"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.navMain.invoices"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(NumericCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "status",
      meta: t("global.status"),
      size: 70,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.status"), alignment: "center", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(StatusBadge, { type: "payment", status: props.row.original.status });
      }
    },
    {
      id: "actions",
      enableHiding: false,
      cell: (props) => {
        const disabled = false;
        return /* @__PURE__ */ jsxs(DropdownMenu, { children: [
          /* @__PURE__ */ jsx(DropdownMenuTrigger, { asChild: true, children: /* @__PURE__ */ jsxs(Button, { variant: "ghost", className: "size-8 p-0", children: [
            /* @__PURE__ */ jsx("span", { className: "sr-only", children: t("global.openMenu") }),
            /* @__PURE__ */ jsx(MoreHorizontal, {})
          ] }) }),
          /* @__PURE__ */ jsxs(DropdownMenuContent, { align: "end", children: [
            /* @__PURE__ */ jsx(DropdownMenuLabel, { children: t("global.actions.title") }),
            /* @__PURE__ */ jsx(DropdownMenuSeparator, {}),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "view"), children: t("payments.viewPayment.title") }),
            props.row.original.status !== "void" && /* @__PURE__ */ jsxs(Fragment, { children: [
              /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "edit"), disabled, children: t("payments.editPayment.title") }),
              /* @__PURE__ */ jsx(DropdownMenuSeparator, {}),
              /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "void"), disabled, children: t("payments.voidPayment.title") })
            ] })
          ] })
        ] });
      }
    }
  ];
};
export {
  getColumns
};
