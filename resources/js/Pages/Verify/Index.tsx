import { Button } from '@/components/ui/button';
import Guest from '@/layouts/guest-layout';
import { PageProps } from '@/types';
import { Head } from '@inertiajs/react';

export default function Verify({ status }: PageProps<{ status: string }>) {
  const submit: FormEventHandler = (e) => {
    e.preventDefault();
  };
  return (
    <Guest>
      <Head title="Email Verification" />
      <div className="mb-4 text-sm text-gray-600 dark:text-gray-400">
        Thanks for signing up! Before getting started, could you verify your email address by clicking on the link we just emailed to you? If you
        didn't receive the email, we will gladly send you another.
      </div>
      <div>{status}</div>
      {/* {status === 'verification-link-sent' && ( */}
      <div className="mb-4 text-sm font-medium text-green-600 dark:text-green-400">
        A new verification link has been sent to the email address you provided during registration.
      </div>
      {/* )} */}

      <form onSubmit={submit}>
        <div className="mt-4 flex items-center justify-between">
          <Button disabled={true}>Resend Verification Email</Button>
        </div>
      </form>
    </Guest>
  );
}
