import { type BreadcrumbItem as BreadcrumbItemType, User } from '@/types';
import React, { JSX } from 'react';
import { AppSidebar } from '@/components/app-sidebar';
import { Breadcrumbs } from '@/components/breadcrumbs';
import { Separator } from '@/components/ui/separator';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';
import { Toaster } from '@/components/ui/sonner';

function Actions({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

interface Props extends React.ComponentProps<'div'> {
  user: User;
  breadcrumbs?: BreadcrumbItemType[];
  children: React.ReactNode;
}

export default class AuthenticatedLayout extends React.Component<Props> {
  static Actions = Actions;
  render() {
    const { user, breadcrumbs = [], children } = this.props;
    const actions = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Actions);
    const content = (React.Children.toArray(children) as React.ReactNode).filter((children: React.ReactNode) => children.type !== Actions);

    return (
      <SidebarProvider>
        <AppSidebar user={user} />
        <SidebarInset>
          <header className="flex justify-between h-16 shrink-0 items-center gap-2 border-b px-4">
            <div className='flex grow'>
              <SidebarTrigger className="-ml-1" />
              <Separator orientation="vertical" className="mr-2 h-4" />
              <Breadcrumbs breadcrumbs={breadcrumbs} />
            </div>
            <div className='min-w-md'>
              {actions ? (actions as JSX.Element) : null}
            </div>
          </header>
          <div className="flex flex-1 flex-col gap-4 p-4">
            <div className="min-h-[100vh] flex-1 rounded-xl md:min-h-min">{content}</div>
          </div>
          <Toaster position="top-right" richColors />
        </SidebarInset>
      </SidebarProvider>
    );
  }
}
