import * as React from 'react';

import { SidebarGroup, SidebarGroupContent, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { useGate } from '@/hooks/use-gate';
import { useTranslation } from '@/hooks/use-translation';
import { NavItem, PageProps } from '@/types';
import { usePage } from '@inertiajs/react';

export function NavSecondary({
  items,
  ...props
}: {
  items: NavItem[];
} & React.ComponentPropsWithoutRef<typeof SidebarGroup>) {
  const { auth } = usePage<PageProps>().props;
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
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {filtered
            .map((item) => {
              return { ...item, url: item.url.replace(/:account/i, auth.account.uuid) };
            })
            .map((item) => (
              <SidebarMenuItem key={item.title}>
                <SidebarMenuButton asChild size="sm">
                  <a href={item.url}>
                    {item.icon && <item.icon />}
                    <span>{t(item.title)}</span>
                  </a>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
