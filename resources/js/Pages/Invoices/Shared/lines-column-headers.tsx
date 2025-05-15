import { useTranslation } from '@/hooks/use-translation';

export default function LinesColumnHeaders() {
  const t = useTranslation().trans;
  return (
    <tr>
      <th scope="col" className="w-60 border border-gray-300 px-1 text-start">
        {t('global.reference')}
      </th>
      <th scope="col" className="w-auto border border-gray-300 px-1 text-start">
        {t('global.description')}
      </th>
      <th scope="col" className="w-36 border border-gray-300 px-1 text-start">
        {t('global.unit')}
      </th>
      <th scope="col" className="w-36 border border-gray-300 px-1 text-end">
        {t('global.qty')}
      </th>
      <th scope="col" className="w-36 border border-gray-300 px-1 text-end">
        {t('global.price')}
      </th>
      <th scope="col" className="w-36 border border-gray-300 px-1 text-end">
        {t('global.amount')}
      </th>
      <th scope="col" className="w-6 gap-2 border border-gray-300 px-5 text-end whitespace-nowrap"></th>
    </tr>
  );
}
