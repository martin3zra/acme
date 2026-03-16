import { FeatureNotImplemented } from '@/components/feature-not-implemented';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { breadcrumbs } from './constants';

export default function Index({ auth }: PageProps) {
  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <FeatureNotImplemented />
    </AppLayout>
  );
}
