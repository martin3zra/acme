import AppLayout from '@/layouts/app-layout';
import SettingsLayout from '@/layouts/settings/layout';
import { BreadcrumbItem, PageProps } from '@/types';
import { Head, usePage } from '@inertiajs/react';
const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Preferences',
    href: '',
  },
];
export default function Preferences() {
  const { auth } = usePage<PageProps>().props;
  return (
    <AppLayout breadcrumbs={breadcrumbs} user={auth.user}>
      <Head title="Preferences" />
      <SettingsLayout>
        <h1>Preferences</h1>
      </SettingsLayout>
    </AppLayout>
  );
}
