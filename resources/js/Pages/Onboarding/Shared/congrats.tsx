import { useTranslation } from '@/hooks/use-translation';
import { Link } from '@inertiajs/react';
import ConfettiBall from './confetti-ball';

export default function Congrats() {
  const t = useTranslation().trans;
  return (
    <div className="flex flex-col items-center justify-center">
      <ConfettiBall />
      <div className="flex max-w-md flex-col items-center justify-center gap-2 p-10">
        <h1 className="text-lg font-medium">{t('onboarding.congrats.title')}</h1>
        <p className="text-base font-normal">{t('onboarding.congrats.description')}</p>

        <Link
          href="/home"
          className="focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive bg-primary text-primary-foreground hover:bg-primary/90 mt-10 inline-flex h-9 shrink-0 items-center justify-center gap-2 rounded-md px-4 py-4 text-sm font-medium whitespace-nowrap shadow-xs transition-all outline-none focus-visible:ring-[3px] disabled:pointer-events-none disabled:opacity-50 has-[>svg]:px-3 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4"
        >
          {t('onboarding.congrats.action')}
        </Link>
      </div>
    </div>
  );
}
