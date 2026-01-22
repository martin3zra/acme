import * as React from 'react';

import { SidebarGroup, SidebarGroupContent, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { useActiveNav } from '@/hooks/use-active-nav';
import { useGate } from '@/hooks/use-gate';
import { useTranslation } from '@/hooks/use-translation';
import { NavItem, PageProps } from '@/types';
import { Link, usePage } from '@inertiajs/react';

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
  const menuItems = useActiveNav(filtered);
  return (
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {menuItems
            .map((item) => {
              return { ...item, url: item.url.replace(/:account/i, auth.account.uuid) };
            })
            .map((item) => (
              <SidebarMenuItem key={item.title}>
                <Link href={item.url}>
                  <SidebarMenuButton
                    asChild
                    tooltip={t(item.title)}
                    className={`${item.isActive ? 'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:text-primary-foreground duration-200 ease-linear' : ''} cursor-pointer`}
                  >
                    <div>
                      {item.icon && <item.icon />}
                      <span>{t(item.title)}</span>
                    </div>
                  </SidebarMenuButton>
                </Link>
                {/* <SidebarMenuButton asChild size="sm">
                  <a href={item.url}>
                    {item.icon && <item.icon />}
                    <span>{t(item.title)}</span>
                  </a>
                </SidebarMenuButton> */}
              </SidebarMenuItem>
            ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
