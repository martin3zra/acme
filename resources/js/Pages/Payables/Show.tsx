import { StatusBadge } from '@/components/status-badge';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { useNumber } from '@/composables/use-number';
import { useTranslation } from '@/hooks/use-translation';
import { cn, isNotEmpty } from '@/lib/utils';
import { Auth, VendorPaymentWithLines } from '@/types';
import { formatDate } from 'date-fns/format';
import { CircleDollarSignIcon, CreditCardIcon } from 'lucide-react';
import PaymentSummary from '../Invoices/Shared/payment-summary';

type Props = {
  vendorPayment: VendorPaymentWithLines;
  auth: Auth;
};

export default function Show({ vendorPayment, auth }: Props) {
  const t = useTranslation().trans;
  const { currency } = useNumber();

  return (
    <div className="grid grid-cols-12 gap-x-4">
      <div className="col-span-12 pb-6">
        <Separator />
        <div className="flex justify-between py-6 [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-medium">
          <div className="col-span-6 flex items-center gap-x-6 [&>div]:flex [&>div]:gap-x-2">
            <div>
              <Label>{t('payables.single.title')}</Label>
              <Label>#{vendorPayment.header.code}</Label>
            </div>
          </div>
          <div className="col-span-6 flex items-center gap-x-2 [&_[data-slot=label]]:font-normal">
            <Label>{t('global.date')}</Label>
            <Label>{formatDate(vendorPayment.header.date, 'dd-MM-yyyy')}</Label>
          </div>
        </div>
        <Separator />
      </div>
      <div className="col-span-9 grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4">
        <div className="col-span-12 grid grid-cols-12">
          <div className="col-span-6">
            <div className="w-56">
              <Label className="font-bold">{t('global.company')}:</Label>
              <div className="pt-2">
                <span className="text-sm font-semibold">{auth.company.name}</span>
                <address className="text-muted-foreground text-sm font-normal">{auth.company.address}</address>
              </div>
            </div>
          </div>
          <div className="col-span-6 place-items-end">
            <div className="w-56">
              <Label className="font-bold">{t('global.vendor')}:</Label>
              <div className="pt-2">
                <span className="text-sm font-semibold">{vendorPayment.header.vendor.name}</span>
                <address className="text-muted-foreground text-sm font-normal">{vendorPayment.header.vendor.address}</address>
              </div>
              <div className="text-sm font-medium">{vendorPayment.header.vendor.email}</div>
            </div>
          </div>
        </div>
        <div className="col-span-12">
          <table
            className={cn(
              'w-full table-auto',
              '[&_th]:border-none [&_th]:border-gray-200 [&_th]:bg-gray-50/25 [&_th]:text-sm [&_th]:font-semibold [&_th]:uppercase',
              '[&_th]:p-2 [&_th]:text-start [&_th[data-format=number]]:text-end',
              '[&_td]:border-y [&_td]:p-2 [&_td]:text-start [&_td[data-format=number]]:w-36 [&_td[data-format=number]]:text-end',
            )}
          >
            <thead>
              <tr className="bg-gray-50/50">
                <th scope="col">{t('payables.single.bill')}</th>
                <th scope="col">{t('global.date')}</th>
                <th scope="col" data-format="number">
                  {t('global.dueDate')}
                </th>
                <th scope="col" data-format="number">
                  {t('global.amount_due')}
                </th>
                <th scope="col" data-format="number">
                  {t('global.payment')}
                </th>
                <th scope="col">{t('global.status')}</th>
              </tr>
            </thead>
            <tbody>
              {vendorPayment.lines.map((line) => (
                <tr key={line.id}>
                  <td>{line.bill_number}</td>
                  <td>{formatDate(line.bill_date, 'dd-MM-yyyy')}</td>
                  <td data-format="number">{formatDate(line.due_date, 'dd-MM-yyyy')}</td>
                  <td data-format="number">{currency(line.amount_due)}</td>
                  <td data-format="number">{currency(line.payment)}</td>
                  <td>
                    <StatusBadge type="paid" status={line.paid_status} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="col-span-12 grid grid-cols-12">
          <div className="col-span-8">
            <div className="max-w-sm rounded-md border p-4">
              <Label className="text-sm/6 font-medium">{t('global.notes')}:</Label>
              <div className="text-muted-foreground text-sm">
                {isNotEmpty(vendorPayment.header.notes) ? vendorPayment.header.notes : t('global.noNotesWasLeft')}
              </div>
            </div>
          </div>
          <div className="col-span-4 rounded-md border p-4">
            <Label>{t('payments.single.totalSummary')}</Label>
            <div
              className={cn(
                'flex flex-col gap-y-3 py-4',
                '[&_[data-slot=label]]:text-muted-foreground [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-normal',
                '[&_[data-slot=label-value]]:text-lg [&_[data-slot=label-value]]:font-medium',
              )}
            >
              <div className="flex items-center justify-between">
                <Label>{t('global.total')}</Label>
                <Label data-slot="label-value">{currency(vendorPayment.header.amount)}</Label>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div className="col-span-3 flex flex-col gap-y-3 rounded-lg border p-3">
        <Label>{t('payables.single.description')}</Label>
        <Separator />
        <div className="flex items-center gap-x-1 text-sm">
          <CircleDollarSignIcon size={14} />
          <span className="font-medium">{t('global.currency')}:</span>
          <span className="text-muted-foreground">Dominican Peso</span>
        </div>
        <div className="flex items-center gap-x-1 text-sm">
          <CreditCardIcon size={14} />
          <span className="font-medium">{t('global.paymentSummary')}</span>
        </div>
        <PaymentSummary paymentData={vendorPayment.header.payment} />
      </div>
    </div>
  );
}
