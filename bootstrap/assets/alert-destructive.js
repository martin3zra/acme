import { jsx, jsxs } from "react/jsx-runtime";
import { AlertCircle } from "lucide-react";
import { cva } from "class-variance-authority";
import { c as cn } from "./utils.js";
import { useState, useEffect } from "react";
const alertVariants = cva(
  "relative w-full rounded-lg border px-4 py-3 text-sm grid has-[>svg]:grid-cols-[calc(var(--spacing)*4)_1fr] grid-cols-[0_1fr] has-[>svg]:gap-x-3 gap-y-0.5 items-start [&>svg]:size-4 [&>svg]:translate-y-0.5 [&>svg]:text-current",
  {
    variants: {
      variant: {
        default: "bg-card text-card-foreground",
        destructive: "text-destructive bg-card [&>svg]:text-current *:data-[slot=alert-description]:text-destructive/90"
      }
    },
    defaultVariants: {
      variant: "default"
    }
  }
);
function Alert({
  className,
  variant,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "alert",
      role: "alert",
      className: cn(alertVariants({ variant }), className),
      ...props
    }
  );
}
function AlertTitle({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "alert-title",
      className: cn(
        "col-start-2 line-clamp-1 min-h-4 font-medium tracking-tight",
        className
      ),
      ...props
    }
  );
}
function AlertDescription({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "alert-description",
      className: cn(
        "text-muted-foreground col-start-2 grid justify-items-start gap-1 text-sm [&_p]:leading-relaxed",
        className
      ),
      ...props
    }
  );
}
function AlertDestructive({ title = "Whoops", description, duration = 3e3, destroyable = true, onDestroy }) {
  const [isVisible, setIsVisible] = useState(true);
  useEffect(() => {
    if (isVisible && destroyable) {
      const timer = setTimeout(() => {
        setIsVisible(false);
        if (onDestroy) onDestroy();
      }, duration);
      return () => clearTimeout(timer);
    }
  }, [isVisible, duration, onDestroy, destroyable]);
  if (!isVisible) {
    return null;
  }
  return /* @__PURE__ */ jsxs(Alert, { variant: "destructive", className: "border-red-400 bg-red-100/50", children: [
    /* @__PURE__ */ jsx(AlertCircle, { className: "h-4 w-4" }),
    /* @__PURE__ */ jsx(AlertTitle, { children: title }),
    /* @__PURE__ */ jsx(AlertDescription, { children: description })
  ] });
}
export {
  AlertDestructive as A
};
