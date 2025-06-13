'use client';

import { SidebarGroup, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { useGate } from '@/hooks/use-gate';
import { useTranslation } from '@/hooks/use-translation';
import { NavItem } from '@/types';
import { Link, usePage } from '@inertiajs/react';

export function NavMain({ items }: { items: NavItem[] }) {
  const { component } = usePage();
  const t = useTranslation().trans;
  const { can } = useGate();
  const filterItems = (items: NavItem[]) =>
    items
      .filter((item) => !item.requiredAbility || can(item.requiredAbility))
      .map((item) => ({
        ...item,
        //   children: item.children?.filter((child) => !child.requiredAbility || hasAbility(child.requiredAbility)),
      }));

  const filtered = filterItems(items);

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{t('global.platform')}</SidebarGroupLabel>
      <SidebarMenu>
        {filtered.map((item) => (
          <SidebarMenuItem key={item.title}>
            <Link href={item.url}>
              <SidebarMenuButton
                asChild
                tooltip={t(item.title)}
                className={`${item.components.includes(component) ? 'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:text-primary-foreground duration-200 ease-linear' : ''} cursor-pointer`}
              >
                <div>
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
}
