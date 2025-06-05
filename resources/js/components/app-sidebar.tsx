import { NavMain } from '@/components/nav-main';
import { NavSecondary } from '@/components/nav-secondary';
import { NavUser } from '@/components/nav-user';
import { Sidebar, SidebarContent, SidebarFooter, SidebarHeader, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { PageProps } from '@/types';
import { Link, usePage } from '@inertiajs/react';
import { ClipboardList, CreditCard, HelpCircleIcon, LayoutDashboardIcon, LayoutListIcon, SearchIcon, SettingsIcon, UsersIcon } from 'lucide-react';
import * as React from 'react';
import AppLogoIcon from './app-logo-icon';

const data = {
  navMain: [
    {
      title: 'global.navMain.dashboard',
      url: '/home',
      icon: LayoutDashboardIcon,
      components: ['Home/Index'],
    },
    {
      title: 'global.navMain.invoices',
      url: '/invoices',
      icon: ClipboardList,
      components: ['Invoices/Index', 'Invoices/Create'],
    },
    {
      title: 'global.navMain.customers',
      url: '/customers',
      icon: UsersIcon,
      components: ['Customers/Index'],
    },
    {
      title: 'global.navMain.items',
      url: '/items',
      icon: LayoutListIcon,
      components: ['Items/Index'],
    },
    {
      title: 'global.navMain.payments',
      url: '/payments',
      icon: CreditCard,
      components: ['Payments/Index', 'Payments/Create'],
    },
  ],
  navSecondary: [
    {
      title: 'global.navSecondary.settings',
      url: '/settings/:account/profile',
      icon: SettingsIcon,
    },
    {
      title: 'global.navSecondary.get-help',
      url: '#',
      icon: HelpCircleIcon,
    },
    {
      title: 'global.navSecondary.search',
      url: '#',
      icon: SearchIcon,
    },
  ],
};

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { auth } = usePage<PageProps>().props;
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
                  <span className="truncate font-medium">{auth.company.name}</span>
                  <span className="truncate text-xs">Enterprise</span>
                </div>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={props.user} />
      </SidebarFooter>
    </Sidebar>
  );
}
