import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';
import ReportLayout, { ReportRequest } from '@/layouts/reports/layout';
import { BreadcrumbItem, defaultBreadcrumbs, PageProps } from '@/types';
import { useState } from 'react';
import { DateRange } from 'react-day-picker';
import DateRangePicker from '../Shared/date-range-picker';
import { DateRangeQuickSelect } from '../Shared/date-range-quick-selector';
export const breadcrumbs: BreadcrumbItem[] = [
  ...defaultBreadcrumbs,
  {
    title: 'reports.title',
    href: '/reports',
  },
  {
    title: 'reports.taxes',
    href: '/reports/taxes',
  },
];

export default function Index({ auth, csrf_token, initialRange, initialPreset }: PageProps<{ initialRange: DateRange; initialPreset: string }>) {
  const t = useTranslation().trans;
  const initialDateRange: DateRange | undefined = initialRange
    ? { from: initialRange.from ? new Date(initialRange.from) : undefined, to: initialRange.to ? new Date(initialRange.to) : undefined }
    : undefined;

  const [request, setRequest] = useState<ReportRequest>({
    endpoint: 'taxes',
    reportType: 'taxes',
    dateRange: initialDateRange,
    csrfToken: csrf_token,
  });

  function updateRequest<K extends keyof ReportRequest>(key: K, value: ReportRequest[K]) {
    setRequest((prev) => ({
      ...prev,
      [key]: value,
    }));
  }

  return (
    <ReportLayout user={auth.user} breadcrumbs={breadcrumbs} activeTab="taxes" request={request} trans={t}>
      <ReportLayout.FilterSection>
        <div className="flex flex-col space-y-4 gap-y-2">
          <div className="flex flex-col space-y-2">
            <Label>{t('global.dateRangePresets')}</Label>
            <DateRangeQuickSelect initialPreset={initialPreset} onChange={(range) => updateRequest('dateRange', range)} />
          </div>
          <div className="flex flex-col space-y-2">
            <Label htmlFor="date">{t('global.dateRange')}</Label>
            <DateRangePicker dateRange={request.dateRange} setDateRange={(range) => updateRequest('dateRange', range)} />
          </div>
        </div>
      </ReportLayout.FilterSection>
    </ReportLayout>
  );
}
