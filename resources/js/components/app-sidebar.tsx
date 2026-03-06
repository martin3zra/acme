import { NavMain } from '@/components/nav-main';
import { NavSecondary } from '@/components/nav-secondary';
import { NavUser } from '@/components/nav-user';
import { Sidebar, SidebarContent, SidebarFooter, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { buildNavGroups } from '@/lib/utils';
import { NavItem, PageProps } from '@/types';
import { Link, usePage } from '@inertiajs/react';
import {
  ChartArea,
  ClipboardList,
  ClipboardPenLineIcon,
  CreditCard,
  LayoutDashboardIcon,
  Package,
  Receipt,
  SettingsIcon,
  ShoppingCartIcon,
  SlidersHorizontal,
  UsersIcon,
  Warehouse,
} from 'lucide-react';
import * as React from 'react';
import AppLogoIcon from './app-logo-icon';

const navMain: NavItem[] = [
  {
    title: 'global.navMain.dashboard',
    url: '/home',
    icon: LayoutDashboardIcon,
    requiredAbility: 'viewAny:dashboard',
  },
  {
    title: 'global.navMain.invoices',
    url: '/invoices',
    icon: ClipboardList,
    requiredAbility: 'viewAny:invoice',
  },
  {
    title: 'global.navMain.estimates',
    url: '/estimates',
    icon: ClipboardPenLineIcon,
    requiredAbility: 'viewAny:estimate',
  },
  {
    title: 'global.navMain.orders',
    url: '/orders',
    icon: ShoppingCartIcon,
    requiredAbility: 'viewAny:order',
  },
  {
    title: 'global.navMain.customers',
    url: '/customers',
    icon: UsersIcon,
    requiredAbility: 'viewAny:customer',
  },
  {
    title: 'global.navMain.warehouses',
    url: '/warehouses',
    icon: Warehouse,
    requiredAbility: 'viewAny:warehouse',
  },
  {
    title: 'global.navMain.attributes',
    url: '/attributes',
    icon: SlidersHorizontal,
    requiredAbility: 'viewAny:attribute',
  },
  {
    title: 'global.navMain.items',
    url: '/items',
    icon: Package,
    requiredAbility: 'viewAny:item',
  },
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
  // {
  //   title: 'global.navSecondary.get-help',
  //   url: '#',
  //   icon: HelpCircleIcon,
  // },
  // {
  //   title: 'global.navSecondary.search',
  //   url: '#',
  //   icon: SearchIcon,
  // },
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
