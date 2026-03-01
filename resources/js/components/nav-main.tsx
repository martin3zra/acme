'use client';

import { SidebarGroup, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { useActiveNav } from '@/hooks/use-active-nav';
import { useGate } from '@/hooks/use-gate';
import { useTranslation } from '@/hooks/use-translation';
import { NavItem } from '@/types';
import { Link } from '@inertiajs/react';

interface NavGroup {
  group: string;
  items: NavItem[];
}

export function NavMain({ groups }: { groups: NavGroup[] }) {
  const t = useTranslation().trans;
  const { can } = useGate();

  const filterItems = (items: NavItem[]) => items.filter((item) => !item.requiredAbility || can(item.requiredAbility));

  return (
    <>
      {groups.map((group) => {
        const filtered = filterItems(group.items);
        const menuItems = useActiveNav(filtered);

        if (menuItems.length === 0) return null;

        return (
          <SidebarGroup key={group.group}>
            <SidebarGroupLabel>{t(group.group)}</SidebarGroupLabel>
            <SidebarMenu>
              {menuItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <Link href={item.url}>
                    <SidebarMenuButton
                      asChild
                      tooltip={t(item.title)}
                      className={`${
                        item.isActive
                          ? 'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:text-primary-foreground duration-200 ease-linear'
                          : ''
                      } cursor-pointer`}
                    >
                      <div className="flex items-center gap-2">
                        {item.icon && <item.icon />}
                        <span>{t(item.title)}</span>
                      </div>
                    </SidebarMenuButton>
                  </Link>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroup>
        );
      })}
    </>
  );
}
