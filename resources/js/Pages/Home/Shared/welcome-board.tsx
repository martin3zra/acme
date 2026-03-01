import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { useTranslation } from '@/hooks/use-translation';
import { CheckCircle, Circle } from 'lucide-react';

export interface ProgressProps {
  customer_created: boolean;
  products_created: boolean;
  invoices_created: boolean;
  estimates_created: boolean;
}

interface Props {
  progress: ProgressProps;
}

export function WelcomeBoard({ progress }: Props) {
  const t = useTranslation().trans;
  const steps = [
    {
      key: 'customer',
      completed: progress.customer_created,
      link: '/customers?mode=creating',
    },
    {
      key: 'item',
      completed: progress.products_created,
      link: '/items?mode=creating',
    },
    { key: 'invoice', completed: progress.invoices_created, link: '/invoices/create' },
    {
      key: 'estimate',
      completed: progress.estimates_created,
      link: '/estimates',
    },
  ];

  const completedSteps = steps.filter((s) => s.completed).length;
  const totalSteps = steps.length;
  const progressPercent = (completedSteps / totalSteps) * 100;

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('dashboard.welcomeBoard.title')}</CardTitle>
        <CardDescription>{t('dashboard.welcomeBoard.description')}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Progress bar */}
        <div>
          <p className="text-muted-foreground mb-2 text-sm">
            {t('dashboard.welcomeBoard.steps.completed', { completed: completedSteps, total: totalSteps })}
          </p>
          <Progress value={progressPercent} className="h-2" />
        </div>

        {/* Checklist */}
        <div className="space-y-3">
          {steps.map((step) => (
            <a key={step.key} href={step.link} className="hover:bg-muted flex items-center justify-between rounded-md border p-3 transition">
              <div className="flex items-center space-x-3">
                {step.completed ? <CheckCircle className="text-success h-5 w-5" /> : <Circle className="text-muted-foreground h-5 w-5" />}
                <div>
                  <p className="font-medium">{t(`dashboard.welcomeBoard.steps.items.${step.key}.name`)}</p>
                  <p className="text-muted-foreground text-sm">{t(`dashboard.welcomeBoard.steps.items.${step.key}.description`)}</p>
                </div>
              </div>
              <Badge variant={step.completed ? 'success' : 'warning'}>
                {step.completed ? t(`dashboard.welcomeBoard.steps.done`) : t(`dashboard.welcomeBoard.steps.pending`)}
              </Badge>
            </a>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
