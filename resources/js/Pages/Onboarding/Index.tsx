import AppLogoIcon from '@/components/app-logo-icon';
import CreateCompanyForm from '@/components/create-company-form';
import { Separator } from '@/components/ui/separator';
import { useTranslation } from '@/hooks/use-translation';
import AppSimpleLayout from '@/layouts/app-simple-layout';
import { PageProps } from '@/types';
import { Link } from '@inertiajs/react';
import Congrats from './Shared/congrats';

export default function Index({ csrf_token, status }: PageProps<{ status: string }>) {
  const t = useTranslation().trans;
  return (
    <AppSimpleLayout>
      <div className="grid grid-cols-12 p-10">
        <div className="col-span-3 flex flex-col gap-4">
          <div className="size-12">
            <AppLogoIcon className="size-4 fill-current text-white dark:text-black" />
          </div>
          <h1 className="text-4xl font-normal">{t('onboarding.title')}</h1>
          <p className="text-base font-normal">{t('onboarding.description')}</p>
          <Separator className="max-w-xs" />
          <Link
            href="/logout"
            method="post"
            headers={{ 'X-CSRF-Token': csrf_token }}
            as="button"
            className="cursor-pointer text-start text-sm text-gray-600 underline hover:text-gray-900"
          >
            {t('global.navUser.logout')}
          </Link>
        </div>
        <div className="col-span-9 py-16">{status === 'success' ? <Congrats /> : <CreateCompanyForm />}</div>
      </div>
    </AppSimpleLayout>
  );
}
