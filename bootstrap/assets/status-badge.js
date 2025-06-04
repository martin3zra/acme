import { jsxs, jsx } from "react/jsx-runtime";
import { u as useTranslation } from "./use-translation.js";
import { c as cn } from "./utils.js";
import { CircleSlash, XCircle, CircleCheckBig, Clock, ShieldBan, ShieldCheck, Wallet, Slash, CheckCircle2, AlertTriangle, Eye, Send, FileText, FileCheck, CircleDollarSign, Trash2, CircleDashed, CheckCircle } from "lucide-react";
const statusConfig = {
  paid: {
    paid: {
      label: "Paid",
      bg: "bg-green-100",
      border: "border-green-500",
      text: "text-green-800",
      Icon: CheckCircle
    },
    unpaid: {
      label: "Unpaid",
      bg: "bg-red-100",
      border: "border-red-500",
      text: "text-red-800",
      Icon: XCircle
    },
    partial: {
      label: "Partial",
      bg: "bg-yellow-100",
      border: "border-yellow-500",
      text: "text-yellow-800",
      Icon: CircleDashed
    },
    removed: {
      label: "Removed",
      bg: "bg-gray-200",
      border: "border-gray-500",
      text: "text-gray-700",
      Icon: Trash2
    },
    overpaid: {
      label: "Overpaid",
      bg: "bg-blue-100",
      border: "border-blue-500",
      text: "text-blue-800",
      Icon: CircleDollarSign
    },
    pending: {
      label: "Pending",
      bg: "bg-orange-100",
      border: "border-orange-500",
      text: "text-orange-800",
      Icon: Clock
    }
  },
  invoice: {
    open: {
      label: "Open",
      bg: "bg-yellow-100",
      border: "border-yellow-500",
      text: "text-yellow-700",
      Icon: FileCheck
    },
    draft: {
      label: "Draft",
      bg: "bg-gray-100",
      border: "border-gray-500",
      text: "text-gray-700",
      Icon: FileText
    },
    sent: {
      label: "Sent",
      bg: "bg-blue-100",
      border: "border-blue-500",
      text: "text-blue-800",
      Icon: Send
    },
    viewed: {
      label: "Viewed",
      bg: "bg-indigo-100",
      border: "border-indigo-500",
      text: "text-indigo-800",
      Icon: Eye
    },
    overdue: {
      label: "Overdue",
      bg: "bg-orange-100",
      border: "border-orange-500",
      text: "text-orange-800",
      Icon: AlertTriangle
    },
    completed: {
      label: "Completed",
      bg: "bg-green-100",
      border: "border-green-500",
      text: "text-green-800",
      Icon: CheckCircle2
    },
    void: {
      label: "Void",
      bg: "bg-red-100",
      border: "border-red-500",
      text: "text-red-800",
      Icon: Slash
    },
    partial: {
      label: "Partial",
      bg: "bg-teal-100",
      border: "border-teal-500",
      text: "text-teal-800",
      Icon: Wallet
    }
  },
  status: {
    enabled: {
      label: "Enabled",
      bg: "bg-green-100",
      border: "border-green-500",
      text: "text-green-800",
      Icon: ShieldCheck
    },
    disabled: {
      label: "Disabled",
      bg: "bg-gray-200",
      border: "border-gray-500",
      text: "text-gray-700",
      Icon: ShieldBan
    }
  },
  payment: {
    pending: {
      label: "Pending",
      bg: "bg-yellow-100",
      border: "border-yellow-500",
      text: "text-yellow-800",
      Icon: Clock
    },
    completed: {
      label: "Completed",
      bg: "bg-blue-100",
      border: "border-blue-500",
      text: "text-blue-800",
      Icon: CircleCheckBig
    },
    failed: {
      label: "Failed",
      bg: "bg-red-100",
      border: "border-red-500",
      text: "text-red-800",
      Icon: XCircle
    },
    void: {
      label: "Voided",
      bg: "bg-gray-100",
      border: "border-gray-400",
      text: "text-gray-700",
      Icon: CircleSlash
    }
  }
};
const StatusBadge = ({ type, status, className = "", variant = "badge", prefix }) => {
  const t = useTranslation().trans;
  const config = statusConfig[type][status];
  if (!config) return null;
  const { Icon, bg, border, text } = config;
  const resolveLabel = () => {
    return {
      invoice: t("invoices.statuses." + status),
      status: t("global.statuses." + status),
      paid: t("global.paidStatuses." + status),
      payment: t("global.paidStatuses." + status)
    }[type];
  };
  return /* @__PURE__ */ jsxs(
    "div",
    {
      className: cn(
        `inline-flex items-center gap-2 text-sm font-medium ${bg} ${text} ${className}`,
        variant === "badge" && "rounded-full px-3 py-1",
        variant === "alert" && `rounded-md border px-4 py-3 ${border}`
      ),
      children: [
        /* @__PURE__ */ jsx(Icon, { className: `h-4 w-4 ${text}` }),
        /* @__PURE__ */ jsxs("span", { children: [
          prefix && /* @__PURE__ */ jsx("span", { className: "pe-1", children: prefix }),
          resolveLabel()
        ] })
      ]
    }
  );
};
export {
  StatusBadge as S
};
