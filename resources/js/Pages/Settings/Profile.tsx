import AppLayout from '@/layouts/app-layout';
import SettingsProfileLayout from '@/layouts/settings/profile-layout';
import { BreadcrumbItem, PageProps } from '@/types';
import { Head, usePage } from '@inertiajs/react';
const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Profile Settings',
    href: '/settings/profile',
  },
];
export default function Profile({ mustVerifyEmail, status }: { mustVerifyEmail: boolean; status?: string }) {
  const { auth } = usePage<PageProps>().props;
  return (
    <AppLayout breadcrumbs={breadcrumbs} user={auth.user}>
      <Head title="Profile Settings" />
      <SettingsProfileLayout>
        <h1>Settings</h1>
      </SettingsProfileLayout>
    </AppLayout>
  );
}
