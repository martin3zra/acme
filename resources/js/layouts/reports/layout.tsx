import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { type BreadcrumbItem as BreadcrumbItemType, Replacements, SlotProps, User } from '@/types';
import { Link } from '@inertiajs/react';
import React, { JSX } from 'react';
import AppLayout from '../app-layout';

interface Props extends React.ComponentProps<'div'> {
  user: User;
  breadcrumbs?: BreadcrumbItemType[];
  activeTab: string;
  pdfUrl?: string;
  children: React.ReactNode;
  trans: (key: string, replacements?: Replacements) => string;
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
export default class ReportLayout extends React.Component<Props> {
  static FilterSection = FilterSection;
  static ContentSection = ContentSection;
  static ActionSection = ActionSection;
  render() {
    const { user, breadcrumbs, children, pdfUrl, trans, activeTab } = this.props;

    const array = React.Children.toArray(children);
    const filterSection = array.find(
      (child): child is React.ReactElement =>
        React.isValidElement(child) && (child.type as React.JSXElementConstructor<SlotProps>) === ReportLayout.FilterSection,
    );
    const contentSection = array.find(
      (child): child is React.ReactElement =>
        React.isValidElement(child) && (child.type as React.JSXElementConstructor<SlotProps>) === ReportLayout.ContentSection,
    );
    const actionSection = array.find(
      (child): child is React.ReactElement =>
        React.isValidElement(child) && (child.type as React.JSXElementConstructor<SlotProps>) === ReportLayout.ActionSection,
    );

    return (
      <AppLayout user={user} breadcrumbs={breadcrumbs}>
        <div className="flex h-full space-x-6">
          <div className="flex h-screen flex-col">
            <Tabs defaultValue={activeTab} className="[&_[data-slot=tabs-trigger]:not([data-state=active])]:cursor-pointer">
              <TabsList>
                <TabsTrigger value="sales" asChild>
                  <Link href="/reports/sales">{trans('reports.sales')}</Link>
                </TabsTrigger>
                <TabsTrigger value="profit-lost" asChild>
                  <Link href="/reports/profit-lost">{trans('reports.profit-lost')}</Link>
                </TabsTrigger>
                <TabsTrigger value="expenses" asChild>
                  <Link href="/reports/expenses">{trans('reports.expenses')}</Link>
                </TabsTrigger>
                <TabsTrigger value="taxes" asChild>
                  <Link href="/reports/taxes">{trans('reports.taxes')}</Link>
                </TabsTrigger>
              </TabsList>
              <div className="py-4">{filterSection ? (filterSection as JSX.Element) : null}</div>
              <TabsContent value={activeTab}>{contentSection ? (contentSection as JSX.Element) : null}</TabsContent>
            </Tabs>
            <div className="py-4">{actionSection ? (actionSection as JSX.Element) : null}</div>
          </div>
          <div className="h-full basis-3/4 bg-gray-300">
            {pdfUrl && <iframe src={pdfUrl} width="100%" height="100%" className="rounded border" title="Report Preview"></iframe>}
          </div>
        </div>
      </AppLayout>
    );
  }
}
