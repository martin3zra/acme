import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';
import { PageProps } from '@/types';
import { Link } from '@inertiajs/react';
import { AlertTriangle } from 'lucide-react';

export default function Index({ csrf_token }: PageProps) {
  const t = useTranslation().trans;
  return (
    <div className="bg-muted flex min-h-screen items-center justify-center px-4">
      <Card className="w-full max-w-md rounded-2xl p-6 text-center shadow-xl">
        <CardHeader>
          <div className="mb-4 flex justify-center text-yellow-500">
            <AlertTriangle className="h-12 w-12" />
          </div>
          <CardTitle className="text-xl font-semibold">{t('global.restricted.title')}</CardTitle>
          <CardDescription className="text-muted-foreground mt-2">{t('global.restricted.description')}</CardDescription>
        </CardHeader>
        <CardContent className="mt-4">
          <Link
            href="/logout"
            method="post"
            headers={{ 'X-CSRF-Token': csrf_token }}
            as="button"
            className="cursor-pointer text-start text-sm text-gray-600 underline hover:text-gray-900"
          >
            {t('global.navUser.logout')}
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}
