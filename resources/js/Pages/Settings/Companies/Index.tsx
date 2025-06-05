import AppLayout from '@/layouts/app-layout';
import SettingsLayout from '@/layouts/settings/layout';
import { BreadcrumbItem, PageProps } from '@/types';
import { Head, usePage } from '@inertiajs/react';
const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Companies Settings',
    href: '',
  },
];
export default function Index() {
  const { auth } = usePage<PageProps>().props;
  return (
    <AppLayout breadcrumbs={breadcrumbs} user={auth.user}>
      <Head title="Companies Settings" />
      <SettingsLayout>
        <h1>Companies</h1>
      </SettingsLayout>
    </AppLayout>
  );
}
