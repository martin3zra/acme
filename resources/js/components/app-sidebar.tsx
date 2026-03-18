import { NavMain } from '@/components/nav-main';
import { NavSecondary } from '@/components/nav-secondary';
import { NavUser } from '@/components/nav-user';
import { Sidebar, SidebarContent, SidebarFooter, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { buildNavGroups } from '@/lib/utils';
import { NavItem, PageProps } from '@/types';
import { Link, usePage } from '@inertiajs/react';
import {
  ArrowLeftRight,
  Boxes,
  ChartArea,
  ClipboardCheck,
  ClipboardEdit,
  ClipboardPenLine,
  CreditCard,
  FileText,
  HelpCircleIcon,
  LayoutDashboard,
  PackageCheck,
  Receipt,
  ReceiptText,
  SearchIcon,
  SettingsIcon,
  ShoppingCart,
  Truck,
  Users,
  Warehouse,
} from 'lucide-react';
import * as React from 'react';
import AppLogoIcon from './app-logo-icon';

const navMain: NavItem[] = [
  {
    title: 'global.navMain.dashboard',
    url: '/home',
    icon: LayoutDashboard,
    requiredAbility: 'viewAny:dashboard',
  },

  // SALES
  {
    title: 'global.navMain.estimates',
    url: '/estimates',
    icon: ClipboardPenLine,
    requiredAbility: 'viewAny:estimate',
  },
  {
    title: 'global.navMain.orders',
    url: '/orders',
    icon: ShoppingCart,
    requiredAbility: 'viewAny:order',
  },
  {
    title: 'global.navMain.invoices',
    url: '/invoices',
    icon: FileText,
    requiredAbility: 'viewAny:invoice',
  },

  // PURCHASING
  {
    title: 'global.navMain.purchaseOrders',
    url: '/purchases/orders',
    icon: ClipboardCheck,
    requiredAbility: 'viewAny:purchase',
  },
  {
    title: 'global.navMain.purchaseReceipts',
    url: '/purchases/receipts',
    icon: PackageCheck,
    requiredAbility: 'viewAny:purchase',
  },
  {
    title: 'global.navMain.vendorBills',
    url: '/purchases/vendor-bills',
    icon: ReceiptText,
    requiredAbility: 'viewAny:purchase',
  },

  // INVENTORY
  {
    title: 'global.navMain.warehouses',
    url: '/inventories/warehouses',
    icon: Warehouse,
    requiredAbility: 'viewAny:warehouse',
  },
  {
    title: 'global.navMain.stock',
    url: '/inventories/stocks',
    icon: Boxes,
    requiredAbility: 'viewAny:stock',
    pill: 'Soon',
    pillVariant: 'soon',
    disabled: true,
  },
  {
    title: 'global.navMain.transfers',
    url: '/inventories/transfers',
    icon: ArrowLeftRight,
    requiredAbility: 'viewAny:stock',
    pill: 'Soon',
    pillVariant: 'soon',
    disabled: true,
  },
  {
    title: 'global.navMain.adjustments',
    url: '/inventories/adjustments',
    icon: ClipboardEdit,
    requiredAbility: 'viewAny:stock',
    pill: 'Soon',
    pillVariant: 'soon',
    disabled: true,
  },

  // CATALOG
  {
    title: 'global.navMain.customers',
    url: '/customers',
    icon: Users,
    requiredAbility: 'viewAny:customer',
  },
  {
    title: 'global.navMain.vendors',
    url: '/vendors',
    icon: Truck,
    requiredAbility: 'viewAny:vendor',
  },
  {
    title: 'global.navMain.items',
    url: '/items',
    icon: Boxes,
    requiredAbility: 'viewAny:item',
  },

  // FINANCE
  {
    title: 'global.navMain.payments',
    url: '/payments',
    icon: CreditCard,
    requiredAbility: 'viewAny:payment',
  },
  {
    title: 'global.navMain.expenses',
    url: '/expenses',
    icon: Receipt,
    requiredAbility: 'viewAny:expense',
  },

  // ANALYTICS
  {
    title: 'global.navMain.reports',
    url: '/reports/sales',
    icon: ChartArea,
    match: ['/reports'],
    requiredAbility: 'viewAny:reports',
  },
];

const navSecondary: NavItem[] = [
  {
    title: 'global.navSecondary.settings',
    url: '/settings/:account/profile',
    match: ['/settings'],
    icon: SettingsIcon,
  },
  {
    title: 'global.navSecondary.get-help',
    url: '#',
    icon: HelpCircleIcon,
    pill: 'Soon',
    pillVariant: 'soon',
    disabled: true,
  },
  {
    title: 'global.navSecondary.search',
    url: '#',
    icon: SearchIcon,
    pill: 'Soon',
    pillVariant: 'soon',
    disabled: true,
  },
];

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { auth } = usePage<PageProps>().props;
  const groups = buildNavGroups(navMain);
  return (
    <Sidebar variant="inset" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <Link href="/home">
                <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                  <AppLogoIcon className="size-4" />
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-medium">{auth.company?.name}</span>
                  <span className="truncate text-xs">Enterprise</span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain groups={groups} />
        <NavSecondary items={navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={props.user} />
      </SidebarFooter>
    </Sidebar>
  );
}
