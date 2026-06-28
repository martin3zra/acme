'use client';

import { SidebarGroup, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@/components/ui/sidebar';
import { useActiveNav } from '@/hooks/use-active-nav';
import { useGate } from '@/hooks/use-gate';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { NavItem, PageProps } from '@/types';
import { Link, usePage } from '@inertiajs/react';
import NavBadge from './nav-badge';
import NavPill from './nav-pill';

interface NavGroup {
  group: string;
  items: NavItem[];
}

// ── NavGroupSection — isolated component so hooks are called at component level
function NavGroupSection({ group, navBadges }: { group: NavGroup; navBadges: Record<string, number | undefined> }) {
  const t = useTranslation().trans;
  const { can } = useGate();

  const filtered = group.items.filter((item) => !item.requiredAbility || can(item.requiredAbility));

  // ✅ Hook called at component level — not inside a .map()
  const menuItems = useActiveNav(filtered);

  if (menuItems.length === 0) return null;

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{t(group.group)}</SidebarGroupLabel>
      <SidebarMenu>
        {menuItems.map((item) => {
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
                {isDisabled ? (
                  // ✅ Non-navigable span for soon/disabled items
                  <span role="link" aria-disabled="true" className="flex w-full items-center gap-2" onClick={(e) => e.preventDefault()}>
                    {item.icon && <item.icon className="h-4 w-4 shrink-0" />}
                    <span className="flex-1 truncate">{t(item.title)}</span>
                    {hasPill && <NavPill label={item.pill!} variant={item.pillVariant} />}
                  </span>
                ) : (
                  // ✅ Link is the direct asChild target — not wrapped in an extra div
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
    </SidebarGroup>
  );
}

export function NavMain({ groups }: { groups: NavGroup[] }) {
  const { navBadges } = usePage<PageProps>().props;

  return (
    <>
      {groups.map((group) => (
        <NavGroupSection key={group.group} group={group} navBadges={navBadges ?? ({} as Record<string, number | undefined>)} />
      ))}
    </>
  );
}
