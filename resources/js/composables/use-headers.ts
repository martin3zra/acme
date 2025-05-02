import { PageProps } from '@/types';
import { usePage } from '@inertiajs/react';

export function useHeader() {
  const { csrf_token } = usePage<PageProps>().props;
  return { headers: { headers: { 'X-CSRF-Token': csrf_token } } };
}
