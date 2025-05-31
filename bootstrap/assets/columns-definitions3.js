import { jsx, jsxs } from "react/jsx-runtime";
import { C as Checkbox, T as TextCell, H as HeaderCell, a as CurrencyCell, D as DateCell } from "./checkbox.js";
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
import "./use-translation.js";
import "@inertiajs/react";
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
      accessorKey: "name",
      meta: t("global.name"),
      header: ({ column }) => {
        return /* @__PURE__ */ jsxs(Button, { className: "uppercase", variant: "ghost", onClick: () => column.toggleSorting(column.getIsSorted() === "asc"), children: [
          t("global.name"),
          " ",
          /* @__PURE__ */ jsx(ArrowUpDown, {})
        ] });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "description",
      meta: t("global.description"),
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.description"), alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "unit.name",
      meta: t("global.unit"),
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.unit"), alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "tax.name",
      meta: t("global.taxRate"),
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.taxRate"), alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
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
        return /* @__PURE__ */ jsx(StatusBadge, { type: "status", status: props.row.original.status });
      }
    },
    {
      accessorKey: "price",
      meta: t("global.price"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.price"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "created_at",
      meta: t("global.addedAt"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.addedAt"), alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(DateCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      id: "actions",
      enableHiding: false,
      cell: (props) => {
        return /* @__PURE__ */ jsxs(DropdownMenu, { children: [
          /* @__PURE__ */ jsx(DropdownMenuTrigger, { asChild: true, children: /* @__PURE__ */ jsxs(Button, { variant: "ghost", className: "size-8 p-0", children: [
            /* @__PURE__ */ jsx("span", { className: "sr-only", children: t("global.openMenu") }),
            /* @__PURE__ */ jsx(MoreHorizontal, {})
          ] }) }),
          /* @__PURE__ */ jsxs(DropdownMenuContent, { align: "end", children: [
            /* @__PURE__ */ jsx(DropdownMenuLabel, { children: t("global.actions.title") }),
            /* @__PURE__ */ jsx(DropdownMenuSeparator, {}),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "view"), children: t("items.viewItem.title") }),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "edit"), children: t("items.editItem.title") }),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "trash"), children: t("items.trashItem.title") })
          ] })
        ] });
      }
    }
  ];
};
export {
  getColumns
};
