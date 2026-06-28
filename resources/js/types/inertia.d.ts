import { PageProps as InertiaPageProps } from '@inertiajs/core';

export interface NavBadgeCounts {
  '/orders'?: number;
  '/invoices'?: number;
  '/purchases/orders'?: number;
  '/purchases/vendor-bills'?: number;
  '/payments'?: number;
  [key: string]: number | undefined;
}

declare module '@inertiajs/core' {
  interface PageProps extends InertiaPageProps {
    navBadges?: NavBadgeCounts;
  }
}
