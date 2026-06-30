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
      // Or match by kind for the shared sales component (estimate/order/invoice).
      // Must be the exact sales path so kind "order" doesn't also light up
      // "/purchases/orders".
      (kind && item.url === `/${kind}s`) ||
      // prefix match (if defined)
      (item.match && item.match.some((prefix: string) => url.startsWith(prefix)));

    return { ...item, isActive };
  });
}
