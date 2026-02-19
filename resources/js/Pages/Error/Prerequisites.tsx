import { useTranslation } from '@/hooks/use-translation';
import { PageProps } from '@/types';
import { Link, usePage } from '@inertiajs/react';

type MissingItem = {
  key: string;
  message: string;
  url: string;
};

type Props = {
  resource: string;
  missing: MissingItem[];
};

export default function Prerequisites({ resource, missing }: Props) {
  const t = useTranslation().trans;
  const { auth } = usePage<PageProps>().props;

  return (
    <>
      <div className="mx-auto mt-20 max-w-2xl rounded-xl bg-white p-6 shadow">
        <h1 className="text-xl font-semibold text-red-600">{t('global.prerequisites.title')}</h1>

        <p className="mt-2 text-gray-600">{t('global.prerequisites.description', { resource: t(`global.prerequisites.resources.${resource}`) })}</p>

        <ul className="mt-4 space-y-2">
          {missing.map((item) => (
            <li key={item.key} className="flex items-start gap-2 rounded border border-red-200 bg-red-50 p-3">
              <Link href={item.url}>
                <span className="text-red-600">•</span>
                <span>{t(`global.prerequisites.missing.${item.key}`)}</span>
              </Link>
            </li>
          ))}
        </ul>

        <div className="mt-6 flex gap-3">
          <a href={`/settings/${auth.account.uuid}/profile`} className="bg-primary text-primary-foreground rounded px-4 py-2">
            {t('global.prerequisites.actions.goToSettings')}
          </a>

          <a href="/home" className="rounded border px-4 py-2">
            {t('global.prerequisites.actions.goToHome')}
          </a>
        </div>
      </div>
    </>
  );
}
