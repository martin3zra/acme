import { jsx, jsxs } from "react/jsx-runtime";
import { A as AlertDestructive } from "./alert-destructive.js";
import { B as Button } from "./button.js";
import { S as Separator } from "./separator.js";
import { S as Sheet, a as SheetContent, b as SheetHeader, c as SheetTitle, d as SheetDescription, e as SheetFooter } from "./sheet.js";
import { e as defaultPaymentMethods } from "./constants2.js";
import { s as subtractFloats, c as cn } from "./utils.js";
import React__default from "react";
import { BankTransferFormView } from "./bank-transfer.js";
import { CardFormView } from "./card-form.js";
import { CheckFormView } from "./check-form.js";
import { InputView } from "./input-view.js";
import "lucide-react";
import "class-variance-authority";
import "@radix-ui/react-slot";
import "@radix-ui/react-separator";
import "@radix-ui/react-dialog";
import "clsx";
import "tailwind-merge";
import "./form-section.js";
import "./input.js";
import "./label.js";
import "@radix-ui/react-label";
import "./use-translation.js";
import "@inertiajs/react";
import "./select.js";
import "@radix-ui/react-select";
class CheckoutForm extends React__default.Component {
  constructor(props) {
    super(props);
    this.hydratePaymentMethods = () => {
      const { paymentForm } = this.props;
      const methods = defaultPaymentMethods;
      methods.filter((p) => p.value == "cash")[0].amount = paymentForm.cash.amount;
      methods.filter((p) => p.value == "ck")[0].amount = paymentForm.ck.amount;
      methods.filter((p) => p.value == "card")[0].amount = paymentForm.card.amount;
      methods.filter((p) => p.value == "bt")[0].amount = paymentForm.bt.amount;
      return methods;
    };
    this.onPaymentChange = (method, value) => {
      this.state.paymentMethods.filter((p) => p.value == method)[0].amount = value;
    };
    this.handleOnChangeInputView = (method, value) => {
      this.setState({ activePaymentForm: method });
      if (typeof value === "number" && method === "cash") {
        const givenValue = isNaN(value) ? 0 : value;
        this.setState(
          (state) => ({ cashForm: { ...state.cashForm, amount: givenValue } }),
          () => this.computeTotals("cash", this.state.cashForm)
        );
        this.onPaymentChange("cash", givenValue);
        return;
      }
    };
    this.handleOnChangeCheckFormView = (value) => {
      if (typeof value === "number") {
        const givenValue = isNaN(value) ? 0 : value;
        this.setState(
          (state) => ({ ckForm: { ...state.ckForm, amount: givenValue } }),
          () => this.computeTotals("ck", this.state.ckForm)
        );
        this.onPaymentChange("ck", givenValue);
        return;
      }
      this.setState(
        (state) => ({ ckForm: { ...state.ckForm, reference: value } }),
        () => this.props.onCheckoutChange("ck", this.state.ckForm)
      );
    };
    this.handleOnChangeCardFormView = (value, key) => {
      if (typeof value === "number" && key === "last4") {
        const givenValue = isNaN(value) ? 0 : value;
        this.setState(
          (state) => ({ cardForm: { ...state.cardForm, last4: givenValue } }),
          () => this.props.onCheckoutChange("card", this.state.cardForm)
        );
        return;
      }
      if (key === "amount") {
        this.setState(
          (state) => ({ cardForm: { ...state.cardForm, [key]: Number(value) } }),
          () => this.computeTotals("card", this.state.cardForm)
        );
        this.onPaymentChange("card", Number(value));
        return;
      }
      this.setState(
        (state) => ({ cardForm: { ...state.cardForm, [key]: value } }),
        () => this.props.onCheckoutChange("card", this.state.cardForm)
      );
    };
    this.handleOnChangeBTFormView = (value) => {
      if (typeof value === "number") {
        const givenValue = isNaN(value) ? 0 : value;
        this.setState(
          (state) => ({ btForm: { ...state.btForm, amount: givenValue } }),
          () => this.computeTotals("bt", this.state.btForm)
        );
        this.onPaymentChange("bt", givenValue);
        return;
      }
      this.setState(
        (state) => ({ btForm: { ...state.btForm, reference: value } }),
        () => this.props.onCheckoutChange("bt", this.state.btForm)
      );
    };
    this.computeTotals = (method, form) => {
      this.setState(
        (state) => {
          const receivedAmount2 = state.paymentMethods.reduce((accumulator, method2) => accumulator + method2.amount, 0);
          return {
            receivedAmount: receivedAmount2,
            remainingBalance: subtractFloats(this.props.totalAmount, receivedAmount2)
          };
        },
        () => this.props.onCheckoutChange(method, form)
      );
    };
    this.renderPaymentMethodForm = () => {
      const { activePaymentForm, ckForm, cardForm, btForm } = this.state;
      return {
        cash: null,
        ck: /* @__PURE__ */ jsx(CheckFormView, { ...ckForm, onChange: this.handleOnChangeCheckFormView }),
        card: /* @__PURE__ */ jsx(CardFormView, { ...cardForm, onChange: this.handleOnChangeCardFormView }),
        bt: /* @__PURE__ */ jsx(BankTransferFormView, { ...btForm, onChange: this.handleOnChangeBTFormView })
      }[activePaymentForm];
    };
    const paymentMethods = this.hydratePaymentMethods();
    const receivedAmount = paymentMethods.reduce((accumulator, method) => accumulator + method.amount, 0);
    const remainingBalance = props.totalAmount - receivedAmount;
    this.state = {
      activePaymentForm: "cash",
      paymentMethods,
      receivedAmount,
      remainingBalance,
      cashForm: props.paymentForm.cash,
      ckForm: props.paymentForm.ck,
      cardForm: props.paymentForm.card,
      btForm: props.paymentForm.bt
    };
  }
  render() {
    const { receivedAmount, remainingBalance } = this.state;
    const { action, openCheckout, setCheckout, errors, totalAmount, onCompleteCheckout, processing, setCancelConfirmation, currency, t } = this.props;
    return /* @__PURE__ */ jsx(Sheet, { open: openCheckout, onOpenChange: setCheckout, children: /* @__PURE__ */ jsxs(SheetContent, { side: "right", className: "m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl", children: [
      /* @__PURE__ */ jsxs(SheetHeader, { children: [
        /* @__PURE__ */ jsxs(SheetTitle, { children: [
          t("global.actions.checkout"),
          ": ",
          t(`global.paymentMethods.${this.state.activePaymentForm}.title`)
        ] }),
        /* @__PURE__ */ jsx(SheetDescription, { className: "text-[12px]", children: t("global.checkoutProcess") })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "grid gap-4 px-4", children: [
        errors.status && /* @__PURE__ */ jsx(AlertDestructive, { description: errors.status, onDestroy: () => delete errors.status }),
        Object.keys(errors).map((e) => /* @__PURE__ */ jsx(AlertDestructive, { description: errors[e], destroyable: false })),
        /* @__PURE__ */ jsx("div", { className: "flex w-full items-center justify-between", children: /* @__PURE__ */ jsxs("table", { className: "w-full table-auto", children: [
          /* @__PURE__ */ jsx("thead", { children: /* @__PURE__ */ jsx("tr", { children: this.state.paymentMethods.map((method) => /* @__PURE__ */ jsx(
            "th",
            {
              "data-slot": `${method.value === this.state.activePaymentForm ? "current" : "default"}`,
              scope: "col",
              className: cn(
                "w-60 border border-gray-300 px-7 text-end",
                "data-[slot=current]:bg-primary data-[slot=current]:text-primary-foreground data-[slot=current]:border-foreground"
              ),
              children: t(`global.paymentMethods.${method.value}.title`)
            },
            method.value
          )) }) }),
          /* @__PURE__ */ jsx("tbody", { children: /* @__PURE__ */ jsx("tr", { children: this.state.paymentMethods.map((method) => /* @__PURE__ */ jsx("td", { className: "border border-gray-300 px-1 text-start", children: /* @__PURE__ */ jsx(
            InputView,
            {
              value: method.amount,
              method: method.value,
              onChange: this.handleOnChangeInputView,
              onFocus: (pm) => this.setState({ activePaymentForm: pm })
            },
            method.value
          ) }, method.value)) }) })
        ] }) }),
        /* @__PURE__ */ jsx("div", { className: "pb-6", children: this.renderPaymentMethodForm() }),
        /* @__PURE__ */ jsx(Separator, { className: "" }),
        /* @__PURE__ */ jsxs("div", { children: [
          /* @__PURE__ */ jsxs("div", { className: "flex w-80 items-center justify-between", children: [
            /* @__PURE__ */ jsx("span", { className: "block text-2xl", children: t("global.totalToBeCollected") }),
            /* @__PURE__ */ jsx("span", { className: "block text-2xl", children: currency(totalAmount) })
          ] }),
          /* @__PURE__ */ jsxs("div", { className: "flex w-80 items-center justify-between", children: [
            /* @__PURE__ */ jsx("span", { className: "block text-2xl", children: t("global.totalReceived") }),
            /* @__PURE__ */ jsx("span", { className: "block text-2xl", children: currency(receivedAmount) })
          ] }),
          /* @__PURE__ */ jsxs("div", { className: "flex w-80 items-center justify-between", children: [
            /* @__PURE__ */ jsx("span", { className: "block text-2xl", children: t("global.balance") }),
            /* @__PURE__ */ jsx("span", { className: "block text-2xl font-medium text-red-600", children: currency(remainingBalance) })
          ] })
        ] })
      ] }),
      /* @__PURE__ */ jsx(SheetFooter, { children: /* @__PURE__ */ jsxs("div", { className: "flex justify-end gap-x-6", children: [
        /* @__PURE__ */ jsx(Button, { variant: "secondary", onClick: () => setCancelConfirmation(true), children: t("global.actions.cancel") }),
        /* @__PURE__ */ jsx(Button, { onClick: onCompleteCheckout, disabled: processing || remainingBalance !== 0, children: action })
      ] }) })
    ] }) });
  }
}
export {
  CheckoutForm as default
};
