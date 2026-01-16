import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
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
    title: 'reports.sales',
    href: '/reports/sales',
  },
];

export default function Index({
  auth,
  csrf_token,
  initialRange,
  initialReportType,
  initialPreset,
}: PageProps<{ initialRange: DateRange; initialPreset: string; initialReportType: string }>) {
  const [pdfUrl, setPdfUrl] = useState<string | undefined>(undefined);
  const [reportType, setReportType] = useState<string>(initialReportType);
  const [showInvoices, setShowInvoices] = useState<boolean>(false);
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

  const generateReport = async () => {
    const response = await fetch('/reports/sales', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
        'X-CSRF-Token': csrf_token as string,
      },
      credentials: 'include',
      body: JSON.stringify({
        from: dateRange?.from?.toISOString(),
        to: dateRange?.to?.toISOString(),
        reportType: reportType,
        showInvoices: showInvoices,
      }),
    });
    if (response.ok) {
      // const data = await response.json();
      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      setPdfUrl(url);
    } else {
      alert('Failed to generate report.');
    }
  };
  return (
    <ReportLayout user={auth.user} breadcrumbs={breadcrumbs} trans={t} activeTab="sales" pdfUrl={pdfUrl}>
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
      <ReportLayout.ContentSection>
        <div>
          <div className="flex flex-col space-y-2">
            <Label>Select Report Type</Label>
            <Select onValueChange={setReportType} defaultValue={reportType}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Select report type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="sales_by_customer">Sales by Customer</SelectItem>
                <SelectItem value="sales_by_item">Sales by Item</SelectItem>
                <SelectItem value="sales_by_date">Sales by Date</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="py-4">
            <Label className="hover:bg-accent/50 flex items-start gap-3 rounded-lg border p-3 has-[[aria-checked=true]]:border-blue-600 has-[[aria-checked=true]]:bg-blue-50 dark:has-[[aria-checked=true]]:border-blue-900 dark:has-[[aria-checked=true]]:bg-blue-950">
              <Checkbox
                id="showInvoices"
                checked={showInvoices}
                onCheckedChange={(value) => {
                  const show = value === true;
                  setShowInvoices(show);
                }}
                className="data-[state=checked]:border-blue-600 data-[state=checked]:bg-blue-600 data-[state=checked]:text-white dark:data-[state=checked]:border-blue-700 dark:data-[state=checked]:bg-blue-700"
              />
              <div className="grid gap-1.5 font-normal">
                <p className="text-sm leading-none font-medium">Display invoices</p>
                <p className="text-muted-foreground text-sm">You can show or hide invoices at any time.</p>
              </div>
            </Label>
          </div>
        </div>
      </ReportLayout.ContentSection>
      <ReportLayout.ActionSection>
        <Button onClick={generateReport}>Generate report</Button>
      </ReportLayout.ActionSection>
    </ReportLayout>
  );
}
