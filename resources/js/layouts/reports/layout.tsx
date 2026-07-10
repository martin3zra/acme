import ActionButton from '@/Pages/Reports/Shared/action-button';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { type BreadcrumbItem as BreadcrumbItemType, Replacements, SlotProps, User } from '@/types';
import { router } from '@inertiajs/react';
import { Download } from 'lucide-react';
import React, { JSX } from 'react';
import { DateRange } from 'react-day-picker';
import AppLayout from '../app-layout';

export type ReportRequest<T extends Record<string, unknown> = Record<string, unknown>> = {
  endpoint: string;
  reportType: string;
  dateRange: DateRange | undefined;
  [key: string]: unknown;
} & T;

interface Props extends React.ComponentProps<'div'> {
  user: User;
  breadcrumbs?: BreadcrumbItemType[];
  activeTab: string;
  request: ReportRequest;
  children: React.ReactNode;
  trans: (key: string, replacements?: Replacements) => string;
}

interface State {
  pdfUrl: string | undefined;
  processing: boolean;
}

function FilterSection({ children }: SlotProps) {
  return <>{children}</>;
}

function ContentSection({ children }: SlotProps) {
  return <>{children}</>;
}

function ActionSection({ children }: SlotProps) {
  return <>{children}</>;
}

export default class ReportLayout extends React.Component<Props, State> {
  static FilterSection = FilterSection;
  static ContentSection = ContentSection;
  static ActionSection = ActionSection;

  constructor(props: Props) {
    super(props);
    this.state = {
      pdfUrl: undefined,
      processing: false,
    };
  }

  handleDownload = () => {
    const { request } = this.props;
    const { pdfUrl } = this.state;
    if (!pdfUrl) {
      return;
    }
    const formatDate = (d?: Date) => (d ? d.toISOString().split('T')[0] : 'unknown');
    const filename = `${request.reportType}_report_${formatDate(request.dateRange?.from)}_${formatDate(request.dateRange?.to)}.pdf`;
    const link = document.createElement('a');
    link.href = pdfUrl;
    link.download = filename;
    link.click();
  };

  generateReport = async () => {
    try {
      const { request } = this.props;
      const { dateRange, endpoint, csrfToken, ...rest } = request;
      const body = {
        from: dateRange?.from?.toISOString(),
        to: dateRange?.to?.toISOString(),
        ...rest,
      };

      this.setState({ pdfUrl: undefined, processing: true });
      const response = await fetch(`/reports/${endpoint}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json',
          'X-CSRF-Token': csrfToken as string,
        },
        credentials: 'include',
        body: JSON.stringify(body),
      });
      if (!response.ok) {
        throw new Error('Failed to generate report');
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      this.setState({ pdfUrl: url });
    } catch (e) {
      console.error(e);
      alert('Failed to generate report.');
    } finally {
      this.setState({ processing: false });
    }
  };

  handleTabChange = (value: string) => {
    const { activeTab } = this.props;

    if (activeTab === value) {
      return; // 🚀 already on this tab → do nothing
    }

    router.visit(`/reports/${value}`, {
      preserveState: true,
      preserveScroll: true,
    });
  };

  render() {
    const { user, breadcrumbs, children, trans, activeTab } = this.props;
    const { processing, pdfUrl } = this.state;

    const array = React.Children.toArray(children);
    const filterSection = array.find(
      (child): child is React.ReactElement =>
        React.isValidElement(child) && (child.type as React.JSXElementConstructor<SlotProps>) === ReportLayout.FilterSection,
    );
    const contentSection = array.find(
      (child): child is React.ReactElement =>
        React.isValidElement(child) && (child.type as React.JSXElementConstructor<SlotProps>) === ReportLayout.ContentSection,
    );

    return (
      <AppLayout user={user} breadcrumbs={breadcrumbs}>
        <AppLayout.Actions>
          <div className="flex justify-end">
            <Button onClick={this.handleDownload} disabled={!pdfUrl} className="flex cursor-pointer items-center gap-2 disabled:cursor-not-allowed">
              <Download className="h-4 w-4" />
              {trans('reports.action.downloadPDF')}
            </Button>
          </div>
        </AppLayout.Actions>
        <div className="flex h-full space-x-6">
          <div className="flex h-screen flex-col">
            <Tabs
              defaultValue={activeTab}
              onValueChange={this.handleTabChange}
              className="[&_[data-slot=tabs-trigger]:not([data-state=active])]:cursor-pointer"
            >
              <TabsList>
                <TabsTrigger value="sales">{trans('reports.sales')}</TabsTrigger>
                <TabsTrigger value="profit-lost">{trans('reports.profit-lost')}</TabsTrigger>
                <TabsTrigger value="expenses">{trans('reports.expenses')}</TabsTrigger>
                <TabsTrigger value="taxes">{trans('reports.taxes')}</TabsTrigger>
              </TabsList>
              <div className="py-4">{filterSection ? (filterSection as JSX.Element) : null}</div>
              <TabsContent value={activeTab}>{contentSection ? (contentSection as JSX.Element) : null}</TabsContent>
            </Tabs>
            <div className="py-4">
              <ActionButton
                idleTitle={trans('reports.action.idle')}
                processingTitle={trans('reports.action.processing')}
                processing={processing}
                handleOnClick={this.generateReport}
              />
            </div>
          </div>
          <div className="h-full basis-3/4 bg-gray-300">
            {pdfUrl ? (
              <iframe src={pdfUrl} width="100%" height="100%" className="rounded border" title="Report Preview"></iframe>
            ) : (
              <div className="flex h-full flex-col items-center justify-center rounded border bg-gray-50 text-gray-500">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="mb-4 h-12 w-12 text-gray-400"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17v-6h13M5 7h14M5 11h14M5 15h14" />
                </svg>
                <h2 className="text-lg font-semibold">{trans('reports.emptyState.title')}</h2>
                <p className="mt-2 text-sm" dangerouslySetInnerHTML={{ __html: trans('reports.emptyState.description') }} />
              </div>
            )}
          </div>
        </div>
      </AppLayout>
    );
  }
}
