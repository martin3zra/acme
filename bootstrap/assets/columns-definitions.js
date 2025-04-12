import { jsx, jsxs } from "react/jsx-runtime";
import { Slot } from "@radix-ui/react-slot";
import { cva } from "class-variance-authority";
import { c as cn, B as Button } from "./button.js";
import * as CheckboxPrimitive from "@radix-ui/react-checkbox";
import { CheckIcon, ArrowUpDown, MoreHorizontal } from "lucide-react";
import { D as DropdownMenu, a as DropdownMenuTrigger, b as DropdownMenuContent, d as DropdownMenuLabel, e as DropdownMenuSeparator, f as DropdownMenuItem } from "./dropdown-menu.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-dropdown-menu";
function useNumber() {
  return { currency };
}
function currency(value, precision = 2, inCent = true) {
  const formatter = new Intl.NumberFormat("en-US", {
    minimumFractionDigits: precision,
    maximumFractionDigits: precision
  });
  if (typeof value === "string") {
    return `$${formatter.format(inCent ? parseInt(value) / 100 : parseFloat(value))}`;
  }
  return `$${formatter.format(inCent ? value / 100 : value)}`;
}
const CurrencyCell = ({ value, columnWidth }) => {
  const { currency: currency2 } = useNumber();
  const formattedValue = value !== void 0 ? currency2(value) : "";
  return /* @__PURE__ */ jsx("div", { className: "truncate p-1.5 text-right", title: formattedValue, style: { width: columnWidth }, children: value });
};
const DateCell = ({ value, columnWidth, locale }) => {
  const formattedValue = value !== void 0 ? new Intl.DateTimeFormat(locale, {
    year: "numeric",
    month: "short",
    weekday: "short",
    day: "numeric"
  }).format(new Date(value)) : "";
  return /* @__PURE__ */ jsx("div", { className: "truncate p-1.5 text-right tabular-nums", title: formattedValue, style: { width: columnWidth }, children: formattedValue });
};
const HeaderCell = ({ title, columnWidth, alignment }) => {
  return /* @__PURE__ */ jsx("div", { className: `p-1.5 font-semibold text-${alignment}`, style: { width: columnWidth }, children: title });
};
const TextCell = ({ value, columnWidth }) => {
  return /* @__PURE__ */ jsx("div", { className: "p-1.5", title: value, style: { width: columnWidth }, children: value });
};
const badgeVariants = cva(
  "inline-flex items-center justify-center rounded-md border px-2 py-0.5 text-xs font-medium w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive transition-[color,box-shadow] overflow-hidden",
  {
    variants: {
      variant: {
        default: "border-transparent bg-primary text-primary-foreground [a&]:hover:bg-primary/90",
        secondary: "border-transparent bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/90",
        destructive: "border-transparent bg-destructive text-white [a&]:hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/70",
        outline: "text-foreground [a&]:hover:bg-accent [a&]:hover:text-accent-foreground"
      }
    },
    defaultVariants: {
      variant: "default"
    }
  }
);
function Badge({
  className,
  variant,
  asChild = false,
  ...props
}) {
  const Comp = asChild ? Slot : "span";
  return /* @__PURE__ */ jsx(
    Comp,
    {
      "data-slot": "badge",
      className: cn(badgeVariants({ variant }), className),
      ...props
    }
  );
}
function Checkbox({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    CheckboxPrimitive.Root,
    {
      "data-slot": "checkbox",
      className: cn(
        "peer border-input dark:bg-input/30 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground dark:data-[state=checked]:bg-primary data-[state=checked]:border-primary focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive size-4 shrink-0 rounded-[4px] border shadow-xs transition-shadow outline-none focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50",
        className
      ),
      ...props,
      children: /* @__PURE__ */ jsx(
        CheckboxPrimitive.Indicator,
        {
          "data-slot": "checkbox-indicator",
          className: "flex items-center justify-center text-current transition-none",
          children: /* @__PURE__ */ jsx(CheckIcon, { className: "size-3.5" })
        }
      )
    }
  );
}
const getColumns = ({ onDidClick }) => {
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
      header: ({ column }) => {
        return /* @__PURE__ */ jsxs(Button, { variant: "ghost", onClick: () => column.toggleSorting(column.getIsSorted() === "asc"), children: [
          "Name ",
          /* @__PURE__ */ jsx(ArrowUpDown, {})
        ] });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "contact_name",
      meta: "Contact",
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: "Contact", alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "phone",
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: "Phone", alignment: "left", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "email",
      header: ({ column }) => {
        return /* @__PURE__ */ jsxs(Button, { variant: "ghost", onClick: () => column.toggleSorting(column.getIsSorted() === "asc"), children: [
          "Email ",
          /* @__PURE__ */ jsx(ArrowUpDown, {})
        ] });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(TextCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "status",
      size: 70,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: "Status", alignment: "center", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        const isEnabled = props.row.original.status === "enabled";
        const capitalized = props.row.original.status.charAt(0).toUpperCase() + props.row.original.status.slice(1);
        return /* @__PURE__ */ jsx("div", { className: `pl-1.5 w-[${props.column.getSize()}px]`, children: /* @__PURE__ */ jsx(Badge, { variant: `${isEnabled ? "default" : "destructive"}`, children: capitalized }) });
      }
    },
    {
      accessorKey: "amount_due",
      meta: "Amount Due",
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: "Amount Due", alignment: "right", columnWidth: props.column.getSize() });
      },
      cell: (props) => {
        return /* @__PURE__ */ jsx(CurrencyCell, { columnWidth: props.column.getSize(), value: props.getValue() });
      }
    },
    {
      accessorKey: "created_at",
      meta: "Added On",
      // size: 880,
      header: (props) => {
        return /* @__PURE__ */ jsx(HeaderCell, { title: "Added On", alignment: "right", columnWidth: props.column.getSize() });
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
            /* @__PURE__ */ jsx("span", { className: "sr-only", children: "Open menu" }),
            /* @__PURE__ */ jsx(MoreHorizontal, {})
          ] }) }),
          /* @__PURE__ */ jsxs(DropdownMenuContent, { align: "end", children: [
            /* @__PURE__ */ jsx(DropdownMenuLabel, { children: "Actions" }),
            /* @__PURE__ */ jsx(DropdownMenuSeparator, {}),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "view"), children: "View" }),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "edit"), children: "Edit" }),
            /* @__PURE__ */ jsx(DropdownMenuItem, { onClick: () => onDidClick(props.row.original, "trash"), children: "Delete" })
          ] })
        ] });
      }
    }
  ];
};
export {
  getColumns
};
