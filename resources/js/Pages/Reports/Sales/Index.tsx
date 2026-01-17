import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useTranslation } from '@/hooks/use-translation';
import ReportLayout from '@/layouts/reports/layout';
import { BreadcrumbItem, defaultBreadcrumbs, PageProps } from '@/types';
import { useState } from 'react';
import { DateRange } from 'react-day-picker';
import ActionButton from '../Shared/action-button';
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
  const [pdfUrl, setPdfUrl] = useState<string | undefined>(undefined);
  const [reportType, setReportType] = useState<string>(initialReportType);
  const [processing, setProcessing] = useState<boolean>(false);
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

  const handleDownload = () => {
    if (!pdfUrl) {
      return;
    }
    const formatDate = (d?: Date) => (d ? d.toISOString().split('T')[0] : 'unknown');
    const filename = `sales_report_${reportType}_${formatDate(dateRange?.from)}_${formatDate(dateRange?.to)}.pdf`;
    const link = document.createElement('a');
    link.href = pdfUrl;
    link.download = filename;
    link.click();
  };

  const generateReport = async () => {
    try {
      setProcessing(true);
      setPdfUrl(undefined);
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
      if (!response.ok) {
        throw new Error('Failed to generate report');
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      setPdfUrl(url);
    } catch (e) {
      console.error(e);
      alert('Failed to generate report.');
    } finally {
      setProcessing(false);
    }
  };
  return (
    <ReportLayout user={auth.user} breadcrumbs={breadcrumbs} trans={t} activeTab="sales" pdfUrl={pdfUrl} handleDownloadAction={handleDownload}>
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
              <SelectTrigger className="w-[280px]">
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
            <Label className="hover:bg-accent/50 has-[[aria-checked=true]]:border-primary flex cursor-pointer items-start gap-3 rounded-lg border p-3 has-[:disabled]:cursor-not-allowed has-[[aria-checked=true]]:bg-blue-50 dark:has-[[aria-checked=true]]:border-blue-900 dark:has-[[aria-checked=true]]:bg-blue-950">
              <Checkbox
                id="showInvoices"
                checked={showInvoices}
                disabled={!invoicesAllowed.includes(reportType)}
                onCheckedChange={(value) => {
                  const show = value === true;
                  setShowInvoices(show);
                }}
                className="data-[state=checked]:border-primary data-[state=checked]:bg-primary dark:data-[state=checked]:bg-primary dark:data-[state=checked]:border-primary disabled:cursor-not-allowed data-[state=checked]:text-white"
              />
              <div className="grid gap-1.5 font-normal">
                <p className="text-sm leading-none font-medium">Display invoices</p>
                {!invoicesAllowed.includes(reportType) ? (
                  <p className="text-muted-foreground text-sm italic">Only available for customer/date-based reports.</p>
                ) : (
                  <p className="text-muted-foreground text-sm">You can show or hide invoices at any time.</p>
                )}
              </div>
            </Label>
          </div>
        </div>
      </ReportLayout.ContentSection>
      <ReportLayout.ActionSection>
        <ActionButton processing={processing} handleOnClick={generateReport} />
      </ReportLayout.ActionSection>
    </ReportLayout>
  );
}
