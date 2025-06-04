import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';

export default function Index({ auth }: PageProps) {
  return <AppLayout user={auth.user}>Error Page</AppLayout>;
}
