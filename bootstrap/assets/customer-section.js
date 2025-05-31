import { jsx, jsxs, Fragment } from "react/jsx-runtime";
import { I as InputError } from "./input-error.js";
import { c as cn } from "./utils.js";
import * as PopoverPrimitive from "@radix-ui/react-popover";
import { PopoverClose } from "@radix-ui/react-popover";
import React__default from "react";
import { b as Command, a as CommandInput, c as CommandList, d as CommandGroup, e as CommandItem } from "./command.js";
import { u as useTranslation } from "./use-translation.js";
import { Link } from "@inertiajs/react";
import { UserPlus, User, Eye, XCircleIcon } from "lucide-react";
function Popover({
  ...props
}) {
  return /* @__PURE__ */ jsx(PopoverPrimitive.Root, { "data-slot": "popover", ...props });
}
function PopoverTrigger({
  ...props
}) {
  return /* @__PURE__ */ jsx(PopoverPrimitive.Trigger, { "data-slot": "popover-trigger", ...props });
}
function PopoverContent({
  className,
  align = "center",
  sideOffset = 4,
  ...props
}) {
  return /* @__PURE__ */ jsx(PopoverPrimitive.Portal, { children: /* @__PURE__ */ jsx(
    PopoverPrimitive.Content,
    {
      "data-slot": "popover-content",
      align,
      sideOffset,
      className: cn(
        "bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 w-72 origin-(--radix-popover-content-transform-origin) rounded-md border p-4 shadow-md outline-hidden",
        className
      ),
      ...props
    }
  ) });
}
function Actions({ children }) {
  return /* @__PURE__ */ jsx(Fragment, { children });
}
function Trigger({ children }) {
  return /* @__PURE__ */ jsx(Fragment, { children });
}
const _InputSearchable = class _InputSearchable extends React__default.Component {
  constructor(props) {
    super(props);
    this.handleOnValueChange = (value) => {
      var _a, _b;
      this.setState({ search: value });
      (_b = (_a = this.props).onChange) == null ? void 0 : _b.call(_a, value);
    };
    this.handleOnTrigger = (event) => {
      if (this.props.value) {
        event.preventDefault();
      }
    };
    this.state = {
      search: props.search
    };
  }
  render() {
    const { children, onOpenChange, title, options, value, valueKey, renderText, renderEmptyText, onSelected, open } = this.props;
    const { search } = this.state;
    const actions = React__default.Children.toArray(children).find((children2) => children2.type === Actions);
    const trigger = React__default.Children.toArray(children).find((children2) => children2.type === Trigger);
    return /* @__PURE__ */ jsxs(Popover, { modal: true, onOpenChange, children: [
      /* @__PURE__ */ jsx(PopoverTrigger, { asChild: true, onClick: this.handleOnTrigger, children: /* @__PURE__ */ jsx("div", { className: cn("h-fit", value === void 0 && !open && "h-full", value !== void 0 && open && "h-fit", !open && "h-full"), children: trigger }) }),
      /* @__PURE__ */ jsx(PopoverContent, { className: "w-[var(--radix-popover-trigger-width)] p-0", onOpenAutoFocus: (e) => e.preventDefault(), children: /* @__PURE__ */ jsxs(Command, { shouldFilter: false, className: "w-full", children: [
        /* @__PURE__ */ jsx(CommandInput, { placeholder: `${title}`, value: typeof search === "string" ? search : "", onValueChange: this.handleOnValueChange }),
        /* @__PURE__ */ jsx(CommandList, { children: /* @__PURE__ */ jsxs(CommandGroup, { className: "max-h-60 overflow-y-auto", children: [
          /* @__PURE__ */ jsx(PopoverClose, { asChild: true, children: /* @__PURE__ */ jsx("div", { children: options && options.map((option) => /* @__PURE__ */ jsx(CommandItem, { value: option[valueKey], onSelect: () => onSelected == null ? void 0 : onSelected(option), children: renderText(option) }, option[valueKey])) }) }),
          options && options.length === 0 && search && /* @__PURE__ */ jsx(CommandItem, { children: renderEmptyText }),
          actions && /* @__PURE__ */ jsx(CommandItem, { children: actions })
        ] }) })
      ] }) })
    ] });
  }
};
_InputSearchable.Actions = Actions;
_InputSearchable.Trigger = Trigger;
let InputSearchable = _InputSearchable;
const CustomerSection = ({
  customer,
  customers,
  errors,
  handleCustomerSelection,
  setSearch,
  open,
  setOpen,
  dedbouncedSearch
}) => {
  const t = useTranslation().trans;
  const handleOnCloseClick = (event) => {
    event.preventDefault();
    handleCustomerSelection(void 0);
    setOpen(false);
  };
  const EmptyCard = () => {
    return /* @__PURE__ */ jsxs(
      "div",
      {
        "data-slot": `${(errors == null ? void 0 : errors.customer_id) ? "customer-error" : "default"}`,
        className: cn(
          "flex h-full w-full flex-col items-center justify-center px-2 pb-2 [&_svg]:text-white",
          "data-[slot=customer-error]:rounded-lg data-[slot=customer-error]:border data-[slot=customer-error]:bg-red-100/50",
          "data-[slot=customer-error]:border-red-500 data-[slot=customer-error]:[&_[data-label=true]]:text-red-500 data-[slot=customer-error]:[&_svg]:text-red-500"
        ),
        children: [
          /* @__PURE__ */ jsxs("button", { onClick: () => setOpen(true), className: "flex h-full w-full cursor-pointer items-center justify-center gap-2", children: [
            /* @__PURE__ */ jsx("div", { className: "flex size-10 items-center justify-center rounded-full bg-gray-200", children: /* @__PURE__ */ jsx(User, { className: "size-6 *:data-[slot=customer-error]:text-red-500" }) }),
            /* @__PURE__ */ jsx("div", { "data-label": "true", className: "text-lg", children: t("global.customer") })
          ] }),
          /* @__PURE__ */ jsx(InputError, { className: "mt-2", message: errors.customer_id })
        ]
      }
    );
  };
  const CustomerCard = () => {
    return /* @__PURE__ */ jsxs("div", { className: "flex h-full flex-col overflow-y-hidden p-2", children: [
      /* @__PURE__ */ jsxs("div", { className: "flex w-full items-center justify-between", children: [
        /* @__PURE__ */ jsx("div", { children: customer == null ? void 0 : customer.name }),
        /* @__PURE__ */ jsxs("div", { className: "flex items-center gap-x-1.5 p-1", children: [
          /* @__PURE__ */ jsx(Link, { href: `/customers?id=${customer == null ? void 0 : customer.uuid}`, children: /* @__PURE__ */ jsx(Eye, { className: "text-muted-foreground size-8 stroke-1" }) }),
          /* @__PURE__ */ jsx("button", { onClick: handleOnCloseClick, className: "cursor-pointer", children: /* @__PURE__ */ jsx(XCircleIcon, {}) })
        ] })
      ] }),
      /* @__PURE__ */ jsx("div", { children: customer == null ? void 0 : customer.email }),
      /* @__PURE__ */ jsx("div", { children: customer == null ? void 0 : customer.phone }),
      /* @__PURE__ */ jsx("div", { children: customer == null ? void 0 : customer.address })
    ] });
  };
  return /* @__PURE__ */ jsx("div", { className: "rounded-lg bg-white shadow", children: /* @__PURE__ */ jsxs(
    InputSearchable,
    {
      open,
      title: t("global.customerSection.search"),
      valueKey: "id",
      renderText: (customer2) => `${customer2.name}`,
      renderEmptyText: t("global.customerSection.emptyText"),
      onSelected: handleCustomerSelection,
      onChange: (value) => setSearch(value),
      onOpenChange: setOpen,
      value: customer,
      options: customers,
      search: dedbouncedSearch,
      children: [
        /* @__PURE__ */ jsx(InputSearchable.Trigger, { children: /* @__PURE__ */ jsxs("div", { className: "h-full", children: [
          !open && !customer && /* @__PURE__ */ jsx(EmptyCard, {}),
          customer && /* @__PURE__ */ jsx(CustomerCard, {})
        ] }) }),
        /* @__PURE__ */ jsx(InputSearchable.Actions, { children: /* @__PURE__ */ jsx("div", { className: "flex w-full items-center justify-center rounded-b-lg border bg-gray-100/25 py-2", children: /* @__PURE__ */ jsxs("button", { className: "flex cursor-pointer items-center justify-center gap-x-2 text-indigo-400", onClick: () => alert("Create new customer"), children: [
          /* @__PURE__ */ jsx(UserPlus, { className: "size-4" }),
          " ",
          t("global.customerSection.addNew")
        ] }) }) })
      ]
    }
  ) });
};
const customerSection = /* @__PURE__ */ Object.freeze(/* @__PURE__ */ Object.defineProperty({
  __proto__: null,
  CustomerSection
}, Symbol.toStringTag, { value: "Module" }));
export {
  CustomerSection as C,
  Popover as P,
  PopoverTrigger as a,
  PopoverContent as b,
  customerSection as c
};
