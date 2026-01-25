import { Button } from '@/components/ui/button';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { useInitials } from '@/hooks/use-initials';
import { useTranslation } from '@/hooks/use-translation';
import { Company } from '@/types';
import { Avatar, AvatarFallback } from '@radix-ui/react-avatar';
import { BadgeCheck } from 'lucide-react';

type Props = {
  company: Company;
};
export default function Overview({ company }: Props) {
  const t = useTranslation().trans;
  const getInitials = useInitials();
  return (
    <div className="flex max-w-4xl flex-col gap-y-6 py-6">
      <div className="flex items-end gap-6">
        <div className="flex size-22 items-center">
          <Avatar className="bg-muted flex h-22 w-22 items-center justify-center rounded-full">
            <AvatarFallback className="rounded-lg text-4xl">{getInitials(company.name)}</AvatarFallback>
          </Avatar>
        </div>
        <div className="mb-2">
          <div className="flex items-end gap-2">
            <h1 className="text-2xl">{company.name}</h1>
            <BadgeCheck size={16} className="mb-1" />
          </div>
          <h4 className="text-foreground text-sm">
            <span>{t('companies.single.rnc_short')}: </span>
            {company.identifier}
          </h4>
        </div>
      </div>
      <div>
        <h4>Address</h4>
        <p>
          {company.address}, {company.city}
        </p>
      </div>
      <Card className="bg-red-100 text-red-700">
        <CardHeader>
          <CardTitle>Danger Zone</CardTitle>
        </CardHeader>
        <CardContent>
          <p>
            If the company is disabled, everything related to employees, payroll, income, discounts, or any other action will no longer be accepted.
          </p>
        </CardContent>
        <CardFooter>
          <Button variant={'destructive'}>Mark as disabled</Button>
        </CardFooter>
      </Card>
    </div>
  );
}
