import { jsx, jsxs } from "react/jsx-runtime";
import { C as Checkbox, T as TextCell, H as HeaderCell, D as DateCell, a as CurrencyCell } from "./checkbox.js";
import { useState } from "react";
import { B as Button } from "./button.js";
import { D as DropdownMenu, a as DropdownMenuTrigger, b as DropdownMenuContent, d as DropdownMenuLabel, e as DropdownMenuSeparator, f as DropdownMenuItem } from "./dropdown-menu.js";
import { Link } from "@inertiajs/react";
import { ArrowUpDown, MoreHorizontal } from "lucide-react";
import "./use-number.js";
import "date-fns";
import "@radix-ui/react-checkbox";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "@radix-ui/react-dropdown-menu";
function EditableCell({
  getValue,
  row,
  column,
  table,
  identifier,
  inputType = "text"
}) {
  const initialValue = getValue();
  const [value, setValue] = useState(initialValue);
  const onBlur = () => {
    var _a;
    (_a = table.options.meta) == null ? void 0 : _a.updateData(identifier, row.index, column.id, value);
  };
  return /* @__PURE__ */ jsx(
    "input",
    {
      id: identifier,
      type: inputType,
      value,
      onChange: (e) => {
        const val = inputType === "number" ? Number(e.target.value) : e.target.value;
        setValue(val);
      },
      onBlur,
      "data-type": inputType,
      className: "h-10 w-full px-2 py-0"
    }
  );
}
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
      meta: t("payments.single.invoice"),
      header: ({ column }) => {
        return /* @__PURE__ */ jsxs(Button, { className: "uppercase", variant: "ghost", onClick: () => column.toggleSorting(column.getIsSorted() === "asc"), children: [
          t("payments.single.invoice"),
          " ",
          /* @__PURE__ */ jsx(ArrowUpDown, {})
        ] });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "ncf",
      meta: "NCF",
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: "NCF", alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
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
      accessorKey: "due_on",
      meta: t("global.dueDate"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.dueDate"), alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(DateCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "total",
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
      accessorKey: "balance",
      meta: t("global.balance"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.balance"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), value: props.row.original.amount_due })
    },
    {
      accessorKey: "payment",
      meta: t("global.payment"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.payment"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(EditableCell, { ...props, identifier: props.row.original.uuid, inputType: "number" });
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
        return /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "remaining",
      meta: t("global.balance"),
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: t("global.balance"), alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        const remaining = props.row.original.amount_due - props.row.original.payment;
        return /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), value: remaining });
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
            /* @__PURE__ */ jsx(DropdownMenuItem, { children: /* @__PURE__ */ jsx(Link, { href: `/invoices?id=${props.row.original.uuid}`, children: t("payments.single.viewInvoice") }) }),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "trash"), children: t("global.actions.delete") })
          ] })
        ] });
      }
    }
  ];
};
export {
  getColumns
};
