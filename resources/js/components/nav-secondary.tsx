import * as React from 'react';

import { SidebarGroup, SidebarGroupContent, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { useActiveNav } from '@/hooks/use-active-nav';
import { useGate } from '@/hooks/use-gate';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { NavItem, PageProps } from '@/types';
import { Link, usePage } from '@inertiajs/react';
import { NavBadge, NavPill } from './nav-indicators';

export function NavSecondary({
  items,
  ...props
}: {
  items: NavItem[];
} & React.ComponentPropsWithoutRef<typeof SidebarGroup>) {
  const { auth, navBadges } = usePage<PageProps>().props;
  const t = useTranslation().trans;
  const { can } = useGate();

  const filterItems = (items: NavItem[]) => items.filter((item) => !item.requiredAbility || can(item.requiredAbility)).map((item) => ({ ...item }));

  const filtered = filterItems(items);
  const menuItems = useActiveNav(filtered);

  return (
    <SidebarGroup {...props}>
      <SidebarGroupContent>
        <SidebarMenu>
          {menuItems
            .map((item) => ({
              ...item,
              url: item.url.replace(/:account/i, auth.account.uuid),
            }))
            .map((item) => {
              const badgeCount = navBadges?.[item.url] ?? item.badge ?? 0;
              const isDisabled = item.disabled ?? item.pillVariant === 'soon';
              const hasBadge = badgeCount > 0 && !item.pill;
              const hasPill = !!item.pill;

              return (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton
                    asChild
                    tooltip={t(item.title)}
                    className={cn(
                      'cursor-pointer duration-200 ease-linear',
                      item.isActive &&
                        'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground active:text-primary-foreground',
                      isDisabled && 'cursor-not-allowed opacity-50',
                    )}
                  >
                    {/* SidebarMenuButton expects a single focusable child — use <a> for disabled, <Link> otherwise */}
                    {isDisabled ? (
                      <span role="link" aria-disabled="true" className="flex w-full items-center gap-2" onClick={(e) => e.preventDefault()}>
                        {item.icon && <item.icon className="h-4 w-4 shrink-0" />}
                        <span className="flex-1 truncate">{t(item.title)}</span>
                        {hasPill && <NavPill label={item.pill!} variant={item.pillVariant} />}
                      </span>
                    ) : (
                      <Link href={item.url} className="flex w-full items-center gap-2">
                        {item.icon && <item.icon className="h-4 w-4 shrink-0" />}
                        <span className="flex-1 truncate">{t(item.title)}</span>
                        {hasBadge && <NavBadge count={badgeCount} />}
                        {hasPill && <NavPill label={item.pill!} variant={item.pillVariant} />}
                      </Link>
                    )}
                  </SidebarMenuButton>
                </SidebarMenuItem>
              );
            })}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  );
}
