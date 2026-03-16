import { FeatureNotImplemented } from '@/components/feature-not-implemented';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { makeBreadcrumbs } from './constants';

export default function Index({ auth }: PageProps) {
  return (
    <AppLayout user={auth.user} breadcrumbs={makeBreadcrumbs('purchases.vendor-bills')}>
      <FeatureNotImplemented />
    </AppLayout>
  );
}
