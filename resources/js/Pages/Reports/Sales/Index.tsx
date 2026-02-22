import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
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
    title: 'reports.sales',
    href: '/reports/sales',
  },
];

const invoicesAllowed: string[] = ['sales_by_customer', 'sales_by_date'];

export default function Index({
  auth,
  csrf_token,
  initialRange,
  initialReportType,
  initialPreset,
}: PageProps<{ initialRange: DateRange; initialPreset: string; initialReportType: string }>) {
  const initialDateRange: DateRange | undefined = initialRange
    ? { from: initialRange.from ? new Date(initialRange.from) : undefined, to: initialRange.to ? new Date(initialRange.to) : undefined }
    : undefined;

  const [request, setRequest] = useState<ReportRequest>({
    endpoint: 'sales',
    reportType: initialReportType,
    dateRange: initialDateRange,
    presetKey: initialPreset,
    showInvoices: false,
    csrfToken: csrf_token,
  });

  const t = useTranslation().trans;

  function updateRequest<K extends keyof ReportRequest>(key: K, value: ReportRequest[K]) {
    setRequest((prev) => ({
      ...prev,
      [key]: value,
    }));
  }

  return (
    <ReportLayout user={auth.user} breadcrumbs={breadcrumbs} trans={t} activeTab="sales" request={request}>
      <ReportLayout.FilterSection>
        <div className="flex flex-col space-y-4 gap-y-2">
          <div className="flex flex-col space-y-2">
            <Label>{t('global.dateRangePresets')}</Label>
            <DateRangeQuickSelect
              initialPreset={initialPreset}
              onChange={(presetKey, range) => {
                updateRequest('dateRange', range);
                updateRequest('presetKey', presetKey);
              }}
            />
          </div>
          <div className="flex flex-col space-y-2">
            <Label htmlFor="date">{t('global.dateRange')}</Label>
            <DateRangePicker
              disabled={request.presetKey !== 'custom'}
              dateRange={request.dateRange}
              setDateRange={(range) => updateRequest('dateRange', range)}
            />
          </div>
        </div>
      </ReportLayout.FilterSection>
      <ReportLayout.ContentSection>
        <div>
          <div className="flex flex-col space-y-2">
            <Label>{t('reports.filters.type.title')}</Label>
            <Select onValueChange={(value) => updateRequest('reportType', value)} defaultValue={request.reportType}>
              <SelectTrigger className="w-70">
                <SelectValue placeholder={t('reports.filters.type.title')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="sales_by_customer">{t('reports.filters.type.sales_by_customer')}</SelectItem>
                <SelectItem value="sales_by_item">{t('reports.filters.type.sales_by_item')}</SelectItem>
                <SelectItem value="sales_by_date">{t('reports.filters.type.sales_by_date')}</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="py-4">
            <Label className="hover:bg-accent/50 has-aria-checked:border-primary flex cursor-pointer items-start gap-3 rounded-lg border p-3 has-[:disabled]:cursor-not-allowed has-[[aria-checked=true]]:bg-blue-50 dark:has-[[aria-checked=true]]:border-blue-900 dark:has-[[aria-checked=true]]:bg-blue-950">
              <Checkbox
                id="showInvoices"
                checked={request.showInvoices as boolean}
                disabled={!invoicesAllowed.includes(request.reportType)}
                onCheckedChange={(value) => {
                  const show = value === true;
                  updateRequest('showInvoices', show);
                }}
                className="data-[state=checked]:border-primary data-[state=checked]:bg-primary dark:data-[state=checked]:bg-primary dark:data-[state=checked]:border-primary disabled:cursor-not-allowed data-[state=checked]:text-white"
              />
              <div className="grid gap-1.5 font-normal">
                <p className="text-sm leading-none font-medium">{t('reports.filters.showInvoice.title')}</p>
                {!invoicesAllowed.includes(request.reportType) ? (
                  <p className="text-muted-foreground text-sm italic">{t('reports.filters.showInvoice.descriptionWhenNotAvailable')}</p>
                ) : (
                  <p className="text-muted-foreground text-sm">{t('reports.filters.showInvoice.descriptionWhenAvailable')}</p>
                )}
              </div>
            </Label>
          </div>
        </div>
      </ReportLayout.ContentSection>
    </ReportLayout>
  );
}
