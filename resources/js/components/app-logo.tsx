import { PageProps } from '@/types';
import { usePage } from '@inertiajs/react';
import AppLogoIcon from './app-logo-icon';

export default function AppLogo() {
  const { auth } = usePage<PageProps>().props;
  return (
    <>
      <div className="bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-md">
        <AppLogoIcon className="size-4 fill-current text-white dark:text-black" />
      </div>
      <div className="grid flex-1 text-left text-sm leading-tight">
        <span className="mb-0.5 truncate leading-none font-semibold">{auth.company.name}</span>
        <span className="truncate text-[11px] leading-none tracking-tight opacity-80">Enterprise</span>
      </div>
    </>
  );
}
