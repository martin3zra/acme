import Heading from '@/components/heading';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';
import { NavItem, PageProps } from '@/types';
import { Link, usePage } from '@inertiajs/react';

const sidebarNavItems: NavItem[] = [
  {
    title: 'Account',
    url: '/settings/:account/profile',
    icon: null,
  },
  {
    title: 'Companies',
    url: '/settings/:account/companies',
    icon: null,
  },
  {
    title: 'Users',
    url: '/settings/:account/users',
    icon: null,
  },
];
export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  const { auth } = usePage<PageProps>().props;
  const currentPath = window.location.pathname;
  return (
    <div className="px-4 py-6">
      <Heading title="Settings" description="Manage your profile and account settings" />

      <div className="flex flex-col space-y-8 md:flex-row md:space-y-0 md:space-x-12">
        <aside className="w-full md:w-1/3 lg:w-1/4 xl:w-1/5">
          <nav className="flex flex-col space-y-1 space-x-0">
            {sidebarNavItems
              .map((item) => {
                return { ...item, url: item.url.replace(/:account/i, auth.account.uuid) };
              })
              .map((item) => (
                <Button
                  key={item.url}
                  size="sm"
                  variant="ghost"
                  asChild
                  className={cn('w-full justify-start', {
                    'bg-muted': currentPath === item.url,
                  })}
                >
                  <Link href={item.url} prefetch>
                    {item.title}
                  </Link>
                </Button>
              ))}
          </nav>
        </aside>

        <Separator className="my-6 md:hidden" />

        <div className="flex-1 md:max-w-4xl">
          <section className="w-full space-y-12">{children}</section>
        </div>
      </div>
    </div>
  );
}
