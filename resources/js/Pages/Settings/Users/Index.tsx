import AppLayout from '@/layouts/app-layout';
import SettingsLayout from '@/layouts/settings/layout';
import { BreadcrumbItem, PageProps } from '@/types';
import { Head, usePage } from '@inertiajs/react';
const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Users Settings',
    href: '',
  },
];
export default function Index() {
  const { auth } = usePage<PageProps>().props;
  return (
    <AppLayout breadcrumbs={breadcrumbs} user={auth.user}>
      <Head title="Users Settings" />
      <SettingsLayout>
        <h1>Users</h1>
      </SettingsLayout>
    </AppLayout>
  );
}
