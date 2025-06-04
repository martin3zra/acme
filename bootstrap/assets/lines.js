import { jsxs, Fragment, jsx } from "react/jsx-runtime";
import { A as AlertDestructive } from "./alert-destructive.js";
import { C as CommandDialog, a as CommandInput, b as Command, c as CommandList, d as CommandGroup, e as CommandItem } from "./command.js";
import { u as useNumber } from "./use-number.js";
import { u as useDebounced } from "./use-debounced.js";
import { u as useTranslation } from "./use-translation.js";
import { router } from "@inertiajs/react";
import { LayoutListIcon } from "lucide-react";
import { useState, useEffect } from "react";
import LinesColumnHeaders from "./lines-column-headers.js";
import LinesForm from "./lines-form.js";
import Line from "./lines-line.js";
import "class-variance-authority";
import "./utils.js";
import "clsx";
import "tailwind-merge";
import "cmdk";
import "./dialog.js";
import "@radix-ui/react-dialog";
import "./input.js";
import "./label.js";
import "@radix-ui/react-label";
import "./button.js";
import "@radix-ui/react-slot";
const Lines = ({
  lineError,
  referenceInputRef,
  qtyInputRef,
  lines,
  currentItem,
  handleRemoveLine,
  handleKeyDown,
  handleOnSelected,
  amount,
  setAmount,
  items
}) => {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const dedbouncedSearch = useDebounced(search, 500);
  const currency = useNumber().currency;
  const computedItemAmount = (qty) => {
    setAmount(qty * ((currentItem == null ? void 0 : currentItem.price) || 0));
  };
  useEffect(() => {
    const searchItems = () => {
      router.reload({ only: ["items"], data: { search: dedbouncedSearch }, preserveUrl: true });
    };
    if (dedbouncedSearch) {
      searchItems();
    }
  }, [dedbouncedSearch]);
  const handleOnKeyDown = (event) => {
    if (event.currentTarget.name === "reference") {
      if (event.key === "k" && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        setOpen(true);
        return;
      }
    }
    handleKeyDown(event);
  };
  return /* @__PURE__ */ jsxs(Fragment, { children: [
    /* @__PURE__ */ jsxs("table", { className: "w-full table-auto", children: [
      /* @__PURE__ */ jsxs("thead", { children: [
        /* @__PURE__ */ jsx(
          LinesForm,
          {
            currentItem,
            amount,
            currency,
            handleOnKeyDown,
            computedItemAmount,
            referenceInputRef,
            qtyInputRef
          }
        ),
        /* @__PURE__ */ jsx(LinesColumnHeaders, {})
      ] }),
      /* @__PURE__ */ jsx("tbody", { children: lines && lines.filter((l) => l.action !== "deleted").map((line, index) => /* @__PURE__ */ jsx(Line, { line, index, currency, handleRemoveLine }, index)) }),
      lineError && /* @__PURE__ */ jsx("tfoot", { children: /* @__PURE__ */ jsx("tr", { children: /* @__PURE__ */ jsx("td", { colSpan: 7, children: /* @__PURE__ */ jsx("div", { className: "py-3", children: /* @__PURE__ */ jsx(AlertDestructive, { description: lineError, destroyable: false }) }) }) }) })
    ] }),
    /* @__PURE__ */ jsxs(CommandDialog, { open, onOpenChange: setOpen, children: [
      /* @__PURE__ */ jsx(CommandInput, { placeholder: t("invoices.line.form.search"), value: typeof search === "string" ? search : "", onValueChange: setSearch }),
      /* @__PURE__ */ jsx(Command, { children: /* @__PURE__ */ jsx(CommandList, { className: "min-h-40", children: /* @__PURE__ */ jsx(CommandGroup, { className: "max-h-60 min-h-40 overflow-y-auto", children: items && items.map((item) => /* @__PURE__ */ jsx(
        CommandItem,
        {
          asChild: true,
          value: String(item.id),
          onSelect: () => {
            handleOnSelected(item);
            setOpen(false);
          },
          children: /* @__PURE__ */ jsxs("div", { className: "flex w-full items-center justify-between", children: [
            /* @__PURE__ */ jsxs("div", { className: "flex w-full flex-col items-start justify-start gap-y-0", children: [
              /* @__PURE__ */ jsx("div", { children: item.name }),
              /* @__PURE__ */ jsx("div", { className: "text-muted-foreground text-xs", children: item.description })
            ] }),
            /* @__PURE__ */ jsx("div", { className: "text-xl font-medium", children: currency(item.price) })
          ] })
        },
        String(item.id)
      )) }) }) }),
      /* @__PURE__ */ jsx("div", { className: "flex w-full items-center justify-center rounded-b-lg border bg-gray-100/25 py-2", children: /* @__PURE__ */ jsxs("button", { className: "flex cursor-pointer items-center justify-center gap-x-2 text-indigo-400", onClick: () => alert("Create new item"), children: [
        /* @__PURE__ */ jsx(LayoutListIcon, { className: "size-4" }),
        " ",
        t("invoices.line.form.addNew")
      ] }) })
    ] })
  ] });
};
export {
  Lines
};
