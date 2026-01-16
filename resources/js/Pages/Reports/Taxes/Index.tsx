import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { useTranslation } from '@/hooks/use-translation';
import ReportLayout from '@/layouts/reports/layout';
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

export default function Index({ auth, initialRange, initialPreset }: PageProps<{ initialRange: DateRange; initialPreset: string }>) {
  const initialDateRange: DateRange | undefined = initialRange
    ? { from: initialRange.from ? new Date(initialRange.from) : undefined, to: initialRange.to ? new Date(initialRange.to) : undefined }
    : undefined;
  const [dateRange, setDateRange] = useState<DateRange | undefined>(initialDateRange);
  const t = useTranslation().trans;

  const handleSelect = (range: DateRange | undefined) => {
    if (range?.from) {
      setDateRange(range);
    }
  };
  return (
    <ReportLayout user={auth.user} breadcrumbs={breadcrumbs} trans={t} activeTab="taxes">
      <ReportLayout.FilterSection>
        <div className="flex flex-col space-y-4 gap-y-2">
          <div className="flex flex-col space-y-2">
            <Label>{t('global.dateRangePresets')}</Label>
            <DateRangeQuickSelect initialPreset={initialPreset} onChange={setDateRange} />
          </div>
          <div className="flex flex-col space-y-2">
            <Label htmlFor="date">{t('global.dateRange')}</Label>
            <DateRangePicker dateRange={dateRange} setDateRange={handleSelect} />
          </div>
        </div>
      </ReportLayout.FilterSection>
      <ReportLayout.ContentSection>Nothing to see here yet.</ReportLayout.ContentSection>
      <ReportLayout.ActionSection>
        <Button>Generate report</Button>
      </ReportLayout.ActionSection>
    </ReportLayout>
  );
}
