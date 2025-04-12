import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { PageProps } from '@/types';

export default function Index({ auth }: PageProps) {
  return <AuthenticatedLayout user={auth.user}>Error Page</AuthenticatedLayout>;
}
