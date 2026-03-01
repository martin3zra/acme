import { NavItem } from '@/types';
import { usePage } from '@inertiajs/react';

export function useActiveNav(navItems: NavItem[]) {
  const { url, props } = usePage();
  // If you pass `kind` from the backend (invoice, estimate, order)
  const kind = props.kind as string | undefined;

  return navItems.map((item) => {
    const isActive =
      // Match by URL path
      url.startsWith(item.url) ||
      // Or match by kind if using shared component
      (kind && item.url.includes(kind)) ||
      // prefix match (if defined)
      (item.match && item.match.some((prefix: string) => url.startsWith(prefix)));

    return { ...item, isActive };
  });
}
