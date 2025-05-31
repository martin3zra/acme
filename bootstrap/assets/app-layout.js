import { jsx, jsxs, Fragment } from "react/jsx-runtime";
import * as React from "react";
import React__default, { Fragment as Fragment$1 } from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva } from "class-variance-authority";
import { PanelLeftIcon, ChevronsUpDown, Sparkles, BadgeCheck, CreditCard, Bell, LayoutDashboardIcon, ClipboardList, UsersIcon, LayoutListIcon, SettingsIcon, HelpCircleIcon, SearchIcon, ChevronRight } from "lucide-react";
import { c as cn } from "./utils.js";
import { B as Button } from "./button.js";
import { S as Sheet, a as SheetContent, b as SheetHeader, c as SheetTitle, d as SheetDescription } from "./sheet.js";
import { T as TooltipProvider, a as Tooltip, b as TooltipTrigger, c as TooltipContent } from "./tooltip.js";
import { u as useTranslation } from "./use-translation.js";
import { usePage, Link } from "@inertiajs/react";
import * as AvatarPrimitive from "@radix-ui/react-avatar";
import { D as DropdownMenu, a as DropdownMenuTrigger, b as DropdownMenuContent, d as DropdownMenuLabel, e as DropdownMenuSeparator, g as DropdownMenuGroup, f as DropdownMenuItem } from "./dropdown-menu.js";
import { IconLogout } from "@tabler/icons-react";
import { A as AppLogoIcon } from "./app-logo-icon.js";
import { S as Separator } from "./separator.js";
import { useTheme } from "next-themes";
import { Toaster as Toaster$1 } from "sonner";
const MOBILE_BREAKPOINT = 768;
function useIsMobile() {
  const [isMobile, setIsMobile] = React.useState(void 0);
  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`);
    const onChange = () => {
      setIsMobile(window.innerWidth < MOBILE_BREAKPOINT);
    };
    mql.addEventListener("change", onChange);
    setIsMobile(window.innerWidth < MOBILE_BREAKPOINT);
    return () => mql.removeEventListener("change", onChange);
  }, []);
  return !!isMobile;
}
const SIDEBAR_COOKIE_NAME = "sidebar_state";
const SIDEBAR_COOKIE_MAX_AGE = 60 * 60 * 24 * 7;
const SIDEBAR_WIDTH = "16rem";
const SIDEBAR_WIDTH_MOBILE = "18rem";
const SIDEBAR_WIDTH_ICON = "3rem";
const SIDEBAR_KEYBOARD_SHORTCUT = "b";
const SidebarContext = React.createContext(null);
function useSidebar() {
  const context = React.useContext(SidebarContext);
  if (!context) {
    throw new Error("useSidebar must be used within a SidebarProvider.");
  }
  return context;
}
function SidebarProvider({
  defaultOpen = true,
  open: openProp,
  onOpenChange: setOpenProp,
  className,
  style,
  children,
  ...props
}) {
  const isMobile = useIsMobile();
  const [openMobile, setOpenMobile] = React.useState(false);
  const [_open, _setOpen] = React.useState(defaultOpen);
  const open = openProp ?? _open;
  const setOpen = React.useCallback(
    (value) => {
      const openState = typeof value === "function" ? value(open) : value;
      if (setOpenProp) {
        setOpenProp(openState);
      } else {
        _setOpen(openState);
      }
      document.cookie = `${SIDEBAR_COOKIE_NAME}=${openState}; path=/; max-age=${SIDEBAR_COOKIE_MAX_AGE}`;
    },
    [setOpenProp, open]
  );
  const toggleSidebar = React.useCallback(() => {
    return isMobile ? setOpenMobile((open2) => !open2) : setOpen((open2) => !open2);
  }, [isMobile, setOpen, setOpenMobile]);
  React.useEffect(() => {
    const handleKeyDown = (event) => {
      if (event.key === SIDEBAR_KEYBOARD_SHORTCUT && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        toggleSidebar();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [toggleSidebar]);
  const state = open ? "expanded" : "collapsed";
  const contextValue = React.useMemo(
    () => ({
      state,
      open,
      setOpen,
      isMobile,
      openMobile,
      setOpenMobile,
      toggleSidebar
    }),
    [state, open, setOpen, isMobile, openMobile, setOpenMobile, toggleSidebar]
  );
  return /* @__PURE__ */ jsx(SidebarContext.Provider, { value: contextValue, children: /* @__PURE__ */ jsx(TooltipProvider, { delayDuration: 0, children: /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-wrapper",
      style: {
        "--sidebar-width": SIDEBAR_WIDTH,
        "--sidebar-width-icon": SIDEBAR_WIDTH_ICON,
        ...style
      },
      className: cn(
        "group/sidebar-wrapper has-data-[variant=inset]:bg-sidebar flex min-h-svh w-full",
        className
      ),
      ...props,
      children
    }
  ) }) });
}
function Sidebar({
  side = "left",
  variant = "sidebar",
  collapsible = "offcanvas",
  className,
  children,
  ...props
}) {
  const { isMobile, state, openMobile, setOpenMobile } = useSidebar();
  if (collapsible === "none") {
    return /* @__PURE__ */ jsx(
      "div",
      {
        "data-slot": "sidebar",
        className: cn(
          "bg-sidebar text-sidebar-foreground flex h-full w-(--sidebar-width) flex-col",
          className
        ),
        ...props,
        children
      }
    );
  }
  if (isMobile) {
    return /* @__PURE__ */ jsx(Sheet, { open: openMobile, onOpenChange: setOpenMobile, ...props, children: /* @__PURE__ */ jsxs(
      SheetContent,
      {
        "data-sidebar": "sidebar",
        "data-slot": "sidebar",
        "data-mobile": "true",
        className: "bg-sidebar text-sidebar-foreground w-(--sidebar-width) p-0 [&>button]:hidden",
        style: {
          "--sidebar-width": SIDEBAR_WIDTH_MOBILE
        },
        side,
        children: [
          /* @__PURE__ */ jsxs(SheetHeader, { className: "sr-only", children: [
            /* @__PURE__ */ jsx(SheetTitle, { children: "Sidebar" }),
            /* @__PURE__ */ jsx(SheetDescription, { children: "Displays the mobile sidebar." })
          ] }),
          /* @__PURE__ */ jsx("div", { className: "flex h-full w-full flex-col", children })
        ]
      }
    ) });
  }
  return /* @__PURE__ */ jsxs(
    "div",
    {
      className: "group peer text-sidebar-foreground hidden md:block",
      "data-state": state,
      "data-collapsible": state === "collapsed" ? collapsible : "",
      "data-variant": variant,
      "data-side": side,
      "data-slot": "sidebar",
      children: [
        /* @__PURE__ */ jsx(
          "div",
          {
            "data-slot": "sidebar-gap",
            className: cn(
              "relative w-(--sidebar-width) bg-transparent transition-[width] duration-200 ease-linear",
              "group-data-[collapsible=offcanvas]:w-0",
              "group-data-[side=right]:rotate-180",
              variant === "floating" || variant === "inset" ? "group-data-[collapsible=icon]:w-[calc(var(--sidebar-width-icon)+(--spacing(4)))]" : "group-data-[collapsible=icon]:w-(--sidebar-width-icon)"
            )
          }
        ),
        /* @__PURE__ */ jsx(
          "div",
          {
            "data-slot": "sidebar-container",
            className: cn(
              "fixed inset-y-0 z-10 hidden h-svh w-(--sidebar-width) transition-[left,right,width] duration-200 ease-linear md:flex",
              side === "left" ? "left-0 group-data-[collapsible=offcanvas]:left-[calc(var(--sidebar-width)*-1)]" : "right-0 group-data-[collapsible=offcanvas]:right-[calc(var(--sidebar-width)*-1)]",
              // Adjust the padding for floating and inset variants.
              variant === "floating" || variant === "inset" ? "p-2 group-data-[collapsible=icon]:w-[calc(var(--sidebar-width-icon)+(--spacing(4))+2px)]" : "group-data-[collapsible=icon]:w-(--sidebar-width-icon) group-data-[side=left]:border-r group-data-[side=right]:border-l",
              className
            ),
            ...props,
            children: /* @__PURE__ */ jsx(
              "div",
              {
                "data-sidebar": "sidebar",
                "data-slot": "sidebar-inner",
                className: "bg-sidebar group-data-[variant=floating]:border-sidebar-border flex h-full w-full flex-col group-data-[variant=floating]:rounded-lg group-data-[variant=floating]:border group-data-[variant=floating]:shadow-sm",
                children
              }
            )
          }
        )
      ]
    }
  );
}
function SidebarTrigger({
  className,
  onClick,
  ...props
}) {
  const { toggleSidebar } = useSidebar();
  return /* @__PURE__ */ jsxs(
    Button,
    {
      "data-sidebar": "trigger",
      "data-slot": "sidebar-trigger",
      variant: "ghost",
      size: "icon",
      className: cn("size-7", className),
      onClick: (event) => {
        onClick == null ? void 0 : onClick(event);
        toggleSidebar();
      },
      ...props,
      children: [
        /* @__PURE__ */ jsx(PanelLeftIcon, {}),
        /* @__PURE__ */ jsx("span", { className: "sr-only", children: "Toggle Sidebar" })
      ]
    }
  );
}
function SidebarInset({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "main",
    {
      "data-slot": "sidebar-inset",
      className: cn(
        "bg-background relative flex w-full flex-1 flex-col",
        "md:peer-data-[variant=inset]:m-2 md:peer-data-[variant=inset]:ml-0 md:peer-data-[variant=inset]:rounded-xl md:peer-data-[variant=inset]:shadow-sm md:peer-data-[variant=inset]:peer-data-[state=collapsed]:ml-2",
        className
      ),
      ...props
    }
  );
}
function SidebarHeader({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-header",
      "data-sidebar": "header",
      className: cn("flex flex-col gap-2 p-2", className),
      ...props
    }
  );
}
function SidebarFooter({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-footer",
      "data-sidebar": "footer",
      className: cn("flex flex-col gap-2 p-2", className),
      ...props
    }
  );
}
function SidebarContent({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-content",
      "data-sidebar": "content",
      className: cn(
        "flex min-h-0 flex-1 flex-col gap-2 overflow-auto group-data-[collapsible=icon]:overflow-hidden",
        className
      ),
      ...props
    }
  );
}
function SidebarGroup({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-group",
      "data-sidebar": "group",
      className: cn("relative flex w-full min-w-0 flex-col p-2", className),
      ...props
    }
  );
}
function SidebarGroupLabel({
  className,
  asChild = false,
  ...props
}) {
  const Comp = asChild ? Slot : "div";
  return /* @__PURE__ */ jsx(
    Comp,
    {
      "data-slot": "sidebar-group-label",
      "data-sidebar": "group-label",
      className: cn(
        "text-sidebar-foreground/70 ring-sidebar-ring flex h-8 shrink-0 items-center rounded-md px-2 text-xs font-medium outline-hidden transition-[margin,opacity] duration-200 ease-linear focus-visible:ring-2 [&>svg]:size-4 [&>svg]:shrink-0",
        "group-data-[collapsible=icon]:-mt-8 group-data-[collapsible=icon]:opacity-0",
        className
      ),
      ...props
    }
  );
}
function SidebarGroupContent({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    "div",
    {
      "data-slot": "sidebar-group-content",
      "data-sidebar": "group-content",
      className: cn("w-full text-sm", className),
      ...props
    }
  );
}
function SidebarMenu({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "ul",
    {
      "data-slot": "sidebar-menu",
      "data-sidebar": "menu",
      className: cn("flex w-full min-w-0 flex-col gap-1", className),
      ...props
    }
  );
}
function SidebarMenuItem({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "li",
    {
      "data-slot": "sidebar-menu-item",
      "data-sidebar": "menu-item",
      className: cn("group/menu-item relative", className),
      ...props
    }
  );
}
const sidebarMenuButtonVariants = cva(
  "peer/menu-button flex w-full items-center gap-2 overflow-hidden rounded-md p-2 text-left text-sm outline-hidden ring-sidebar-ring transition-[width,height,padding] hover:bg-sidebar-accent hover:text-sidebar-accent-foreground focus-visible:ring-2 active:bg-sidebar-accent active:text-sidebar-accent-foreground disabled:pointer-events-none disabled:opacity-50 group-has-data-[sidebar=menu-action]/menu-item:pr-8 aria-disabled:pointer-events-none aria-disabled:opacity-50 data-[active=true]:bg-sidebar-accent data-[active=true]:font-medium data-[active=true]:text-sidebar-accent-foreground data-[state=open]:hover:bg-sidebar-accent data-[state=open]:hover:text-sidebar-accent-foreground group-data-[collapsible=icon]:size-8! group-data-[collapsible=icon]:p-2! [&>span:last-child]:truncate [&>svg]:size-4 [&>svg]:shrink-0",
  {
    variants: {
      variant: {
        default: "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
        outline: "bg-background shadow-[0_0_0_1px_hsl(var(--sidebar-border))] hover:bg-sidebar-accent hover:text-sidebar-accent-foreground hover:shadow-[0_0_0_1px_hsl(var(--sidebar-accent))]"
      },
      size: {
        default: "h-8 text-sm",
        sm: "h-7 text-xs",
        lg: "h-12 text-sm group-data-[collapsible=icon]:p-0!"
      }
    },
    defaultVariants: {
      variant: "default",
      size: "default"
    }
  }
);
function SidebarMenuButton({
  asChild = false,
  isActive = false,
  variant = "default",
  size = "default",
  tooltip,
  className,
  ...props
}) {
  const Comp = asChild ? Slot : "button";
  const { isMobile, state } = useSidebar();
  const button = /* @__PURE__ */ jsx(
    Comp,
    {
      "data-slot": "sidebar-menu-button",
      "data-sidebar": "menu-button",
      "data-size": size,
      "data-active": isActive,
      className: cn(sidebarMenuButtonVariants({ variant, size }), className),
      ...props
    }
  );
  if (!tooltip) {
    return button;
  }
  if (typeof tooltip === "string") {
    tooltip = {
      children: tooltip
    };
  }
  return /* @__PURE__ */ jsxs(Tooltip, { children: [
    /* @__PURE__ */ jsx(TooltipTrigger, { asChild: true, children: button }),
    /* @__PURE__ */ jsx(
      TooltipContent,
      {
        side: "right",
        align: "center",
        hidden: state !== "collapsed" || isMobile,
        ...tooltip
      }
    )
  ] });
}
function NavMain({
  items
}) {
  const { component } = usePage();
  const t = useTranslation().trans;
  return /* @__PURE__ */ jsxs(SidebarGroup, { children: [
    /* @__PURE__ */ jsx(SidebarGroupLabel, { children: t("global.platform") }),
    /* @__PURE__ */ jsx(SidebarMenu, { children: items.map((item) => /* @__PURE__ */ jsx(SidebarMenuItem, { children: /* @__PURE__ */ jsx(Link, { href: item.url, children: /* @__PURE__ */ jsx(
      SidebarMenuButton,
      {
        asChild: true,
        tooltip: t(item.title),
        className: `${item.components.includes(component) ? "bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:text-primary-foreground duration-200 ease-linear" : ""} cursor-pointer`,
        children: /* @__PURE__ */ jsxs("div", { children: [
          /* @__PURE__ */ jsx(item.icon, {}),
          /* @__PURE__ */ jsx("span", { children: t(item.title) })
        ] })
      }
    ) }) }, item.title)) })
  ] });
}
function NavSecondary({
  items,
  ...props
}) {
  const t = useTranslation().trans;
  return /* @__PURE__ */ jsx(SidebarGroup, { ...props, children: /* @__PURE__ */ jsx(SidebarGroupContent, { children: /* @__PURE__ */ jsx(SidebarMenu, { children: items.map((item) => /* @__PURE__ */ jsx(SidebarMenuItem, { children: /* @__PURE__ */ jsx(SidebarMenuButton, { asChild: true, size: "sm", children: /* @__PURE__ */ jsxs("a", { href: item.url, children: [
    /* @__PURE__ */ jsx(item.icon, {}),
    /* @__PURE__ */ jsx("span", { children: t(item.title) })
  ] }) }) }, item.title)) }) }) });
}
function Avatar({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    AvatarPrimitive.Root,
    {
      "data-slot": "avatar",
      className: cn(
        "relative flex size-8 shrink-0 overflow-hidden rounded-full",
        className
      ),
      ...props
    }
  );
}
function AvatarImage({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    AvatarPrimitive.Image,
    {
      "data-slot": "avatar-image",
      className: cn("aspect-square size-full", className),
      ...props
    }
  );
}
function AvatarFallback({
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    AvatarPrimitive.Fallback,
    {
      "data-slot": "avatar-fallback",
      className: cn(
        "bg-muted flex size-full items-center justify-center rounded-full",
        className
      ),
      ...props
    }
  );
}
function useInitials() {
  const getInitials = (fullName) => {
    const names = fullName.trim().split(" ");
    if (names.length === 0) return "";
    if (names.length === 1) return names[0].charAt(0).toUpperCase();
    return `${names[0].charAt(0)}${names[names.length - 1].charAt(0)}`.toUpperCase();
  };
  return getInitials;
}
function NavUser({ user }) {
  const { isMobile } = useSidebar();
  const props = usePage().props;
  const t = useTranslation().trans;
  const getInitials = useInitials();
  return /* @__PURE__ */ jsx(SidebarMenu, { children: /* @__PURE__ */ jsx(SidebarMenuItem, { children: /* @__PURE__ */ jsxs(DropdownMenu, { children: [
    /* @__PURE__ */ jsx(DropdownMenuTrigger, { asChild: true, children: /* @__PURE__ */ jsxs(SidebarMenuButton, { size: "lg", className: "data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground", children: [
      /* @__PURE__ */ jsxs(Avatar, { className: "h-8 w-8 rounded-lg", children: [
        /* @__PURE__ */ jsx(AvatarImage, { src: user.avatar, alt: user.name }),
        /* @__PURE__ */ jsx(AvatarFallback, { className: "rounded-lg", children: "CN" })
      ] }),
      /* @__PURE__ */ jsxs("div", { className: "grid flex-1 text-left text-sm leading-tight", children: [
        /* @__PURE__ */ jsx("span", { className: "truncate font-medium", children: user.name }),
        /* @__PURE__ */ jsx("span", { className: "truncate text-xs", children: user.email })
      ] }),
      /* @__PURE__ */ jsx(ChevronsUpDown, { className: "ml-auto size-4" })
    ] }) }),
    /* @__PURE__ */ jsxs(
      DropdownMenuContent,
      {
        className: "w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg",
        side: isMobile ? "bottom" : "right",
        align: "end",
        sideOffset: 4,
        children: [
          /* @__PURE__ */ jsx(DropdownMenuLabel, { className: "p-0 font-normal", children: /* @__PURE__ */ jsxs("div", { className: "flex items-center gap-2 px-1 py-1.5 text-left text-sm", children: [
            /* @__PURE__ */ jsxs(Avatar, { className: "h-8 w-8 rounded-lg", children: [
              /* @__PURE__ */ jsx(AvatarImage, { src: user.avatar, alt: `${user.first_name} ${user.last_name}` }),
              /* @__PURE__ */ jsx(AvatarFallback, { className: "rounded-lg", children: getInitials(`${user.first_name} ${user.last_name}`) })
            ] }),
            /* @__PURE__ */ jsxs("div", { className: "grid flex-1 text-left text-sm leading-tight", children: [
              /* @__PURE__ */ jsx("span", { className: "truncate font-medium", children: user.name }),
              /* @__PURE__ */ jsx("span", { className: "truncate text-xs", children: user.email })
            ] })
          ] }) }),
          /* @__PURE__ */ jsx(DropdownMenuSeparator, {}),
          /* @__PURE__ */ jsx(DropdownMenuGroup, { children: /* @__PURE__ */ jsxs(DropdownMenuItem, { children: [
            /* @__PURE__ */ jsx(Sparkles, {}),
            t("global.navUser.upgrade-to-pro")
          ] }) }),
          /* @__PURE__ */ jsx(DropdownMenuSeparator, {}),
          /* @__PURE__ */ jsxs(DropdownMenuGroup, { children: [
            /* @__PURE__ */ jsxs(DropdownMenuItem, { children: [
              /* @__PURE__ */ jsx(BadgeCheck, {}),
              t("global.navUser.account")
            ] }),
            /* @__PURE__ */ jsxs(DropdownMenuItem, { children: [
              /* @__PURE__ */ jsx(CreditCard, {}),
              t("global.navUser.billing")
            ] }),
            /* @__PURE__ */ jsxs(DropdownMenuItem, { children: [
              /* @__PURE__ */ jsx(Bell, {}),
              t("global.navUser.notifications")
            ] })
          ] }),
          /* @__PURE__ */ jsx(DropdownMenuSeparator, {}),
          /* @__PURE__ */ jsxs(DropdownMenuItem, { children: [
            /* @__PURE__ */ jsx(IconLogout, {}),
            /* @__PURE__ */ jsx(Link, { href: "/logout", method: "post", headers: { "X-CSRF-Token": props.csrf_token }, className: "cursor-pointer", as: "button", children: t("global.navUser.logout") })
          ] })
        ]
      }
    )
  ] }) }) });
}
const data = {
  navMain: [
    {
      title: "global.navMain.dashboard",
      url: "/home",
      icon: LayoutDashboardIcon,
      components: ["Home/Index"]
    },
    {
      title: "global.navMain.invoices",
      url: "/invoices",
      icon: ClipboardList,
      components: ["Invoices/Index", "Invoices/Create"]
    },
    {
      title: "global.navMain.customers",
      url: "/customers",
      icon: UsersIcon,
      components: ["Customers/Index"]
    },
    {
      title: "global.navMain.items",
      url: "/items",
      icon: LayoutListIcon,
      components: ["Items/Index"]
    },
    {
      title: "global.navMain.payments",
      url: "/payments",
      icon: CreditCard,
      components: ["Payments/Index", "Payments/Create"]
    }
  ],
  navSecondary: [
    {
      title: "global.navSecondary.settings",
      url: "#",
      icon: SettingsIcon
    },
    {
      title: "global.navSecondary.get-help",
      url: "#",
      icon: HelpCircleIcon
    },
    {
      title: "global.navSecondary.search",
      url: "#",
      icon: SearchIcon
    }
  ]
};
function AppSidebar({ ...props }) {
  const { auth } = usePage().props;
  return /* @__PURE__ */ jsxs(Sidebar, { variant: "inset", ...props, children: [
    /* @__PURE__ */ jsx(SidebarHeader, { children: /* @__PURE__ */ jsx(SidebarMenu, { children: /* @__PURE__ */ jsx(SidebarMenuItem, { children: /* @__PURE__ */ jsx(SidebarMenuButton, { size: "lg", asChild: true, children: /* @__PURE__ */ jsxs(Link, { href: "/home", children: [
      /* @__PURE__ */ jsx("div", { className: "bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg", children: /* @__PURE__ */ jsx(AppLogoIcon, { className: "size-4" }) }),
      /* @__PURE__ */ jsxs("div", { className: "grid flex-1 text-left text-sm leading-tight", children: [
        /* @__PURE__ */ jsx("span", { className: "truncate font-medium", children: auth.company.name }),
        /* @__PURE__ */ jsx("span", { className: "truncate text-xs", children: "Enterprise" })
      ] })
    ] }) }) }) }) }),
    /* @__PURE__ */ jsxs(SidebarContent, { children: [
      /* @__PURE__ */ jsx(NavMain, { items: data.navMain }),
      /* @__PURE__ */ jsx(NavSecondary, { items: data.navSecondary, className: "mt-auto" })
    ] }),
    /* @__PURE__ */ jsx(SidebarFooter, { children: /* @__PURE__ */ jsx(NavUser, { user: props.user }) })
  ] });
}
function Breadcrumb({ ...props }) {
  return /* @__PURE__ */ jsx("nav", { "aria-label": "breadcrumb", "data-slot": "breadcrumb", ...props });
}
function BreadcrumbList({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "ol",
    {
      "data-slot": "breadcrumb-list",
      className: cn(
        "text-muted-foreground flex flex-wrap items-center gap-1.5 text-sm break-words sm:gap-2.5",
        className
      ),
      ...props
    }
  );
}
function BreadcrumbItem({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "li",
    {
      "data-slot": "breadcrumb-item",
      className: cn("inline-flex items-center gap-1.5", className),
      ...props
    }
  );
}
function BreadcrumbLink({
  asChild,
  className,
  ...props
}) {
  const Comp = asChild ? Slot : "a";
  return /* @__PURE__ */ jsx(
    Comp,
    {
      "data-slot": "breadcrumb-link",
      className: cn("hover:text-foreground transition-colors", className),
      ...props
    }
  );
}
function BreadcrumbPage({ className, ...props }) {
  return /* @__PURE__ */ jsx(
    "span",
    {
      "data-slot": "breadcrumb-page",
      role: "link",
      "aria-disabled": "true",
      "aria-current": "page",
      className: cn("text-foreground font-normal", className),
      ...props
    }
  );
}
function BreadcrumbSeparator({
  children,
  className,
  ...props
}) {
  return /* @__PURE__ */ jsx(
    "li",
    {
      "data-slot": "breadcrumb-separator",
      role: "presentation",
      "aria-hidden": "true",
      className: cn("[&>svg]:size-3.5", className),
      ...props,
      children: children ?? /* @__PURE__ */ jsx(ChevronRight, {})
    }
  );
}
function Breadcrumbs({ breadcrumbs }) {
  const t = useTranslation().trans;
  return /* @__PURE__ */ jsx(Fragment, { children: breadcrumbs && breadcrumbs.length > 0 && /* @__PURE__ */ jsx(Breadcrumb, { children: /* @__PURE__ */ jsx(BreadcrumbList, { children: breadcrumbs.map((item, index) => {
    const isLast = index === breadcrumbs.length - 1;
    return /* @__PURE__ */ jsxs(Fragment$1, { children: [
      /* @__PURE__ */ jsx(BreadcrumbItem, { children: isLast ? /* @__PURE__ */ jsx(BreadcrumbPage, { children: t(item.title) }) : /* @__PURE__ */ jsx(BreadcrumbLink, { asChild: true, children: /* @__PURE__ */ jsx(Link, { href: item.href, children: t(item.title) }) }) }),
      !isLast && /* @__PURE__ */ jsx(BreadcrumbSeparator, {})
    ] }, index);
  }) }) }) });
}
const Toaster = ({ ...props }) => {
  const { theme = "system" } = useTheme();
  return /* @__PURE__ */ jsx(
    Toaster$1,
    {
      theme,
      className: "toaster group",
      style: {
        "--normal-bg": "var(--popover)",
        "--normal-text": "var(--popover-foreground)",
        "--normal-border": "var(--border)"
      },
      ...props
    }
  );
};
function Actions({ children }) {
  return /* @__PURE__ */ jsx(Fragment, { children });
}
const _AppLayout = class _AppLayout extends React__default.Component {
  render() {
    const { user, breadcrumbs = [], children } = this.props;
    const actions = React__default.Children.toArray(children).find((children2) => children2.type === Actions);
    const content = React__default.Children.toArray(children).filter((children2) => children2.type !== Actions);
    return /* @__PURE__ */ jsxs(SidebarProvider, { children: [
      /* @__PURE__ */ jsx(AppSidebar, { user }),
      /* @__PURE__ */ jsxs(SidebarInset, { children: [
        /* @__PURE__ */ jsxs("header", { className: "flex h-16 shrink-0 items-center justify-between gap-2 border-b px-4", children: [
          /* @__PURE__ */ jsxs("div", { className: "flex grow", children: [
            /* @__PURE__ */ jsx(SidebarTrigger, { className: "-ml-1" }),
            /* @__PURE__ */ jsx(Separator, { orientation: "vertical", className: "mr-2 h-4" }),
            /* @__PURE__ */ jsx(Breadcrumbs, { breadcrumbs })
          ] }),
          /* @__PURE__ */ jsx("div", { className: "min-w-md", children: actions ? actions : null })
        ] }),
        /* @__PURE__ */ jsx("div", { className: "flex flex-1 flex-col gap-4 p-4", children: /* @__PURE__ */ jsx("div", { className: "grid min-h-[100vh] grid-flow-col grid-cols-1 grid-rows-1 rounded-xl md:min-h-min", children: content }) }),
        /* @__PURE__ */ jsx(Toaster, { position: "top-right", richColors: true })
      ] })
    ] });
  }
};
_AppLayout.Actions = Actions;
let AppLayout = _AppLayout;
export {
  AppLayout as A
};
