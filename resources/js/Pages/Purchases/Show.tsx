import { StatusBadge } from '@/components/status-badge';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { useNumber } from '@/composables/use-number';
import { useTranslation } from '@/hooks/use-translation';
import { cn, isNotEmpty } from '@/lib/utils';
import type { Auth, PurchaseTransactionKind, PurchaseWithLines } from '@/types';
import { Link } from '@inertiajs/react';
import { format } from 'date-fns';
import { FileText } from 'lucide-react';

type Props = {
  kind: PurchaseTransactionKind;
  purchase: PurchaseWithLines;
  auth: Auth;
};

export default function Show({ kind, purchase, auth }: Props) {
  const t = useTranslation().trans;
  const { currency } = useNumber();

  return (
    <div className="grid grid-cols-12 gap-x-4">
      <div className="col-span-12 pb-6">
        <Separator />
        <div className="flex justify-between py-6 [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-medium">
          <div className="col-span-6 flex items-center gap-x-6 [&>div]:flex [&>div]:gap-x-2">
            <div>
              <Label>{t('global.purchase')}</Label>
              <Label>#{purchase.header.number}</Label>
            </div>
          </div>
          <div className="col-span-6 flex items-center gap-x-2 [&_[data-slot=label]]:font-normal">
            <Label>{t('global.date')}</Label>
            <Label>{format(purchase.header.date, 'dd-MM-yyyy')}</Label>
            {purchase.header.due_on && (
              <>
                <Separator orientation="vertical" />
                <Label>{t('global.dueDateAbrev')}</Label>
                <Label>{format(purchase.header.due_on, 'dd-MM-yyyy')}</Label>
              </>
            )}
            {kind === 'purchase_order' &&
              purchase.header.linked_receipts?.map((receipt) => (
                <Link
                  key={receipt.id}
                  href={`/purchases/receipts?id=${receipt.uuid}`}
                  className="flex items-center space-x-2 rounded bg-red-200 px-2 py-1 text-sm text-red-500"
                >
                  <FileText className="mr-2 size-4" />
                  {t('purchases.orders.linkedReceipt', { receipt: receipt.number })}
                </Link>
              ))}
          </div>
        </div>
        <Separator />
      </div>

      <div className="col-span-8 grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4">
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
                <Link href={`/vendors?id=${purchase.header.vendor.uuid}`} className="text-sm font-semibold">
                  {purchase.header.vendor.name}
                </Link>
                <address className="text-muted-foreground text-sm font-normal">{purchase.header.vendor.address}</address>
              </div>
              <div className="text-sm font-medium">{purchase.header.vendor.email}</div>
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
                <th scope="col" className="text-start">
                  ID
                </th>
                <th scope="col" className="text-start">
                  {t('global.name')}
                </th>
                <th scope="col" className="w-20 text-start">
                  {t('global.unit')}
                </th>
                <th scope="col" data-format="number" className="!w-20">
                  {t('global.qty')}
                </th>
                <th scope="col" data-format="number">
                  {t('global.price')}
                </th>
                <th scope="col" data-format="number">
                  {t('global.tax')}
                </th>
                <th scope="col" data-format="number">
                  {t('global.amount')}
                </th>
              </tr>
            </thead>
            <tbody>
              {purchase.lines.map((line) => (
                <tr key={line.id}>
                  <td>{line.id}</td>
                  <td>{line.name}</td>
                  <td>{line.unit.name}</td>
                  <td data-format="number" className="!w-16">
                    {line.qty}
                  </td>
                  <td data-format="number">{currency(line.price)}</td>
                  <td data-format="number">{currency(line.tax.amount)}</td>
                  <td data-format="number">{currency(line.amount)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="col-span-12 grid grid-cols-12">
          <div className="col-span-7">
            <div className="max-w-sm rounded-md border p-4">
              <Label className="text-sm/6 font-medium">{t('global.notes')}:</Label>
              <div className="text-muted-foreground text-sm">
                {isNotEmpty(purchase.header.notes) ? purchase.header.notes : t('global.noNotesWasLeft')}
              </div>
            </div>
          </div>
          <div className="col-span-5 rounded-md border p-4">
            <Label>{t('global.total')}</Label>
            <div
              className={cn(
                'flex flex-col gap-y-3 py-4',
                '[&_[data-slot=label]]:text-muted-foreground [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-normal',
                '[&_[data-slot=label-value]]:text-lg [&_[data-slot=label-value]]:font-medium',
              )}
            >
              <div className="flex items-center justify-between">
                <Label>{t('global.subTotal')}</Label>
                <Label data-slot="label-value">{currency(purchase.header.amount)}</Label>
              </div>
              <div className="flex items-center justify-between">
                <Label>{t('global.discount')}</Label>
                <Label data-slot="label-value">
                  {purchase.header.discount.type === 'fixed' ? currency(purchase.header.discount.value) : `${purchase.header.discount.value}%`}
                </Label>
              </div>
              <div className="flex items-center justify-between">
                <Label>{t('global.tax')}</Label>
                <Label data-slot="label-value">{currency(purchase.header.tax)}</Label>
              </div>
              <Separator />
              <div className="flex items-center justify-between">
                <Label>{t('global.total')}</Label>
                <Label data-slot="label-value">{currency(purchase.header.total)}</Label>
              </div>
              <div className="flex items-center justify-between">
                <Label>{t('global.balance')}</Label>
                <Label data-slot="label-value">{currency(purchase.header.amount_due)}</Label>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="col-span-4 flex flex-col gap-y-3 rounded-lg border p-3">
        <StatusBadge type="paid" variant="alert" prefix={`${t('global.payment')}:`} status={purchase.header.payment_status} />
        <Separator />
        <div className="flex items-center justify-between">
          <Label className="text-lg">{currency(purchase.header.total)}</Label>
        </div>
        <Separator />
        <div className="text-muted-foreground text-sm">{t('purchases.preview.description')}</div>
        <div className="text-muted-foreground text-xs">{kind}</div>
      </div>
    </div>
  );
}
