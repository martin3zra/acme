'use client';

import { SidebarGroup, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { Link, usePage } from '@inertiajs/react';
import { type LucideIcon } from 'lucide-react';

export function NavMain({
  items,
}: {
  items: {
    title: string;
    url: string;
    icon: LucideIcon;
    components: string[];
  }[];
}) {
  const { component } = usePage();
  return (
    <SidebarGroup>
      <SidebarGroupLabel>Platform</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) => (
          <SidebarMenuItem key={item.title}>
            <Link href={item.url}>
              <SidebarMenuButton
                asChild
                tooltip={item.title}
                className={`${item.components.includes(component) ? 'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:text-primary-foreground duration-200 ease-linear' : ''} cursor-pointer`}
              >
                <div>
                  <item.icon />
                  <span>{item.title}</span>
                </div>
              </SidebarMenuButton>
            </Link>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  );
}
