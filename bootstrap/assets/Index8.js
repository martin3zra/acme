import { jsxs, jsx } from "react/jsx-runtime";
import { B as Button } from "./button.js";
import { D as DropdownMenu, a as DropdownMenuTrigger, b as DropdownMenuContent, c as DropdownMenuCheckboxItem } from "./dropdown-menu.js";
import { I as Input } from "./input.js";
import { T as Table, a as TableHeader, b as TableRow, c as TableHead, d as TableBody, e as TableCell } from "./table.js";
import { u as useTranslation } from "./use-translation.js";
import { useReactTable, getFilteredRowModel, getSortedRowModel, getPaginationRowModel, getCoreRowModel, flexRender } from "@tanstack/react-table";
import { ChevronDown } from "lucide-react";
import { useState } from "react";
import { getColumns } from "./columns-definitions3.js";
import "@radix-ui/react-slot";
import "class-variance-authority";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "@radix-ui/react-dropdown-menu";
import "@inertiajs/react";
import "./checkbox.js";
import "./use-number.js";
import "date-fns";
import "@radix-ui/react-checkbox";
import "./status-badge.js";
const List = ({ data, onSelectItem }) => {
  var _a, _b;
  const t = useTranslation().trans;
  const [sorting, setSorting] = useState([]);
  const [columnFilters, setColumnFilters] = useState([]);
  const [columnVisibility, setColumnVisibility] = useState({});
  const [rowSelection, setRowSelection] = useState({});
  const columns = getColumns({ onDidClick: onSelectItem, t });
  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection
    }
  });
  return /* @__PURE__ */ jsxs("div", { children: [
    /* @__PURE__ */ jsxs("div", { className: "flex items-center py-4", children: [
      /* @__PURE__ */ jsx(
        Input,
        {
          placeholder: "Filter names...",
          value: ((_a = table.getColumn("name")) == null ? void 0 : _a.getFilterValue()) ?? "",
          onChange: (event) => {
            var _a2;
            return (_a2 = table.getColumn("name")) == null ? void 0 : _a2.setFilterValue(event.target.value);
          },
          className: "max-w-sm"
        }
      ),
      /* @__PURE__ */ jsxs(DropdownMenu, { children: [
        /* @__PURE__ */ jsx(DropdownMenuTrigger, { asChild: true, children: /* @__PURE__ */ jsxs(Button, { variant: "outline", className: "ml-auto", children: [
          t("global.columns"),
          " ",
          /* @__PURE__ */ jsx(ChevronDown, {})
        ] }) }),
        /* @__PURE__ */ jsx(DropdownMenuContent, { align: "end", children: table.getAllColumns().filter((column) => column.getCanHide()).map((column) => {
          return /* @__PURE__ */ jsx(
            DropdownMenuCheckboxItem,
            {
              className: "capitalize",
              checked: column.getIsVisible(),
              onCheckedChange: (value) => column.toggleVisibility(!!value),
              children: column.columnDef.meta || column.id
            },
            column.id
          );
        }) })
      ] })
    ] }),
    /* @__PURE__ */ jsx("div", { className: "rounded-md border", children: /* @__PURE__ */ jsxs(Table, { children: [
      /* @__PURE__ */ jsx(TableHeader, { children: table.getHeaderGroups().map((headerGroup) => /* @__PURE__ */ jsx(TableRow, { children: headerGroup.headers.map((header) => {
        return /* @__PURE__ */ jsx(TableHead, { children: header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext()) }, header.id);
      }) }, headerGroup.id)) }),
      /* @__PURE__ */ jsx(TableBody, { children: ((_b = table.getRowModel().rows) == null ? void 0 : _b.length) ? table.getRowModel().rows.map((row) => /* @__PURE__ */ jsx(TableRow, { "data-state": row.getIsSelected() && "selected", children: row.getVisibleCells().map((cell) => /* @__PURE__ */ jsx(TableCell, { children: flexRender(cell.column.columnDef.cell, cell.getContext()) }, cell.id)) }, row.id)) : /* @__PURE__ */ jsx(TableRow, { children: /* @__PURE__ */ jsx(TableCell, { colSpan: columns.length, className: "h-24 text-center", children: t("global.noResults") }) }) })
    ] }) }),
    /* @__PURE__ */ jsxs("div", { className: "flex items-center justify-end space-x-2 py-4", children: [
      /* @__PURE__ */ jsx("div", { className: "text-muted-foreground flex-1 text-sm", children: t("global.pagination.selectedOf", {
        rows: table.getFilteredSelectedRowModel().rows.length,
        total: table.getFilteredRowModel().rows.length
      }) }),
      /* @__PURE__ */ jsxs("div", { className: "space-x-2", children: [
        /* @__PURE__ */ jsx(Button, { value: "outline", size: "sm", onClick: () => table.previousPage(), disabled: !table.getCanPreviousPage(), children: t("global.pagination.previous") }),
        /* @__PURE__ */ jsx(Button, { value: "outline", size: "sm", onClick: () => table.nextPage(), disabled: !table.getCanNextPage(), children: t("global.pagination.next") })
      ] })
    ] })
  ] });
};
export {
  List
};
