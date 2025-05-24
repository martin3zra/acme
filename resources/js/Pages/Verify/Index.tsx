import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import Guest from '@/layouts/guest-layout';
import { PageProps } from '@/types';
import { Head, Link, useForm } from '@inertiajs/react';
import { FormEventHandler } from 'react';

interface EmailVerificationForm {
  email: string;
}

export default function Verify({ status, csrf_token }: PageProps<{ status: string }>) {
  const { headers } = useHeader();
  const t = useTranslation().trans;
  const { data, setData, errors, post, processing } = useForm<Required<EmailVerificationForm>>({
    email: '',
  });

  const showForm = status !== 'verification-link-sent' && status !== 'account-verified' && status !== 'already-verified';

  const submit: FormEventHandler = (e) => {
    e.preventDefault();

    post('/email/verification-notification', { ...headers });
  };
  return (
    <Guest>
      <Head title="Email Verification" />

      {status !== 'verification-link-sent' && (
        <div
          className="mb-4 text-base text-gray-600 dark:text-gray-400"
          dangerouslySetInnerHTML={{ __html: t(`verify.${status}`).replace(/\n/g, '<br>') }}
        />
      )}

      {status === 'already-verified' && (
        <Link href="/home">
          {t('global.visit', { to: 'tu' })}
          {t('global.navMain.dashboard')}
        </Link>
      )}

      {status === 'verification-link-sent' && (
        <div className="mb-4 text-base font-medium text-green-600 dark:text-green-400">{t(`verify.${status}`)}</div>
      )}

      {status === 'account-verified' && (
        <>
          <Link href="/login">{t('verify.login')}</Link>

          <Link
            href="/logout"
            method="post"
            headers={{ 'X-CSRF-Token': csrf_token }}
            as="button"
            className="rounded-md text-sm text-gray-600 underline hover:text-gray-900 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:outline-none"
          >
            Log Out
          </Link>
        </>
      )}

      {showForm && (
        <form onSubmit={submit}>
          <div className="grid gap-2">
            <Label htmlFor="email">{t(`global.email`)}</Label>
            <Input
              id="email"
              type="email"
              required
              autoFocus
              tabIndex={1}
              autoComplete="email"
              value={data.email}
              onChange={(e) => setData('email', e.target.value)}
            />
            <InputError message={errors.email} />
          </div>
          <div className="mt-4 flex items-center justify-between">
            <Button disabled={processing}>{t(`verify.resend-verification-email`)}</Button>
          </div>
        </form>
      )}
    </Guest>
  );
}
