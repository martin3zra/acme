import { Card, CardHeader } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';
import { StatItem } from '@/types';
import { fallbackIcon, iconMap } from '@/types/icons';

interface DashboardSummaryProps {
  stats: StatItem[];
}

export default function DashboardSummary({ stats }: DashboardSummaryProps) {
  const t = useTranslation().trans;
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {stats.map((stat) => {
        const Icon = iconMap[stat.icon as keyof typeof iconMap] || fallbackIcon;
        if (Icon === fallbackIcon) {
          return <div key="unknow-stat" className="bg-muted/50 rounded-xl" />;
        }
        return (
          <Card key={stat.label}>
            <CardHeader className="flex items-start justify-between">
              <div>
                <p className="text-xl font-semibold">{stat.value}</p>
                <p className="text-muted-foreground text-base">{t(stat.label)}</p>
              </div>
              <div className={`rounded-full p-2 ${stat.bg}`}>
                <Icon className="size-5 text-current" />
              </div>
            </CardHeader>
          </Card>
        );
      })}
    </div>
  );
}
