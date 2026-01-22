import { StatusBadge } from '@/components/status-badge';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { useNumber } from '@/composables/use-number';
import { useTranslation } from '@/hooks/use-translation';
import { capitalize, cn, isNotEmpty } from '@/lib/utils';
import { Auth, InvoiceWithLines, PaidStatuses, TransactionKind } from '@/types';
import { format } from 'date-fns';
import { Calendar1, CircleCheckIcon, CircleDollarSignIcon, CreditCardIcon, UserPen } from 'lucide-react';
import PaymentSummary from './Shared/payment-summary';

type Props = {
  kind: TransactionKind;
  invoice: InvoiceWithLines;
  auth: Auth;
};

export default function Show({ kind, invoice, auth }: Props) {
  const isInvoice = kind === 'invoice';
  const t = useTranslation().trans;
  const { currency } = useNumber();

  return (
    <div className="grid grid-cols-12 gap-x-4">
      <div className="col-span-12 pb-6">
        <Separator />
        <div className="flex justify-between py-6 [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-medium">
          <div className="col-span-6 flex items-center gap-x-6 [&>div]:flex [&>div]:gap-x-2">
            <div>
              <Label>{t(`${kind}s.single.title`)}</Label>
              <Label>#{invoice.header.number}</Label>
            </div>
            {isInvoice && (
              <div>
                <Label>NCF</Label>
                <Label>{invoice.header.ncf}</Label>
              </div>
            )}
          </div>
          <div className="col-span-6 flex items-center gap-x-2 [&_[data-slot=label]]:font-normal">
            <Label>{t('global.date')}</Label>
            <Label className="">{format(invoice.header.date, 'dd-MM-yyyy')}</Label>
          </div>
        </div>
        <Separator />
      </div>
      <div className="col-span-8 grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4">
        {/* header */}
        <div className="col-span-12 grid grid-cols-12">
          <div className="col-span-6">
            <div className="w-56">
              <Label className="font-bold">{t(`${kind}s.single.payTo`)}:</Label>
              <div className="pt-2">
                <span className="text-sm font-semibold">{auth.company.name}</span>
                <address className="text-muted-foreground text-sm font-normal">{auth.company.address}</address>
              </div>
              <div className="text-sm font-medium">{invoice.header.customer.email}</div>
            </div>
          </div>
          <div className="col-span-6 place-items-end">
            <div className="w-56">
              <Label className="font-bold">{t(`${kind}s.single.${kind}To`)}:</Label>
              <div className="pt-2">
                <span className="text-sm font-semibold">{invoice.header.customer.name}</span>
                <address className="text-muted-foreground text-sm font-normal">{invoice.header.customer.address}</address>
              </div>
              <div className="text-sm font-medium">{invoice.header.customer.email}</div>
            </div>
          </div>
        </div>
        {/* lines */}
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
              {invoice.lines.map((line) => (
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
        {/* footer */}
        <div className="col-span-12 grid grid-cols-12">
          <div className="col-span-7">
            <div className="max-w-sm rounded-md border p-4">
              <Label className="text-sm/6 font-medium">{t('global.notes')}:</Label>
              <div className="text-muted-foreground text-sm">
                {isNotEmpty(invoice.header.notes) ? invoice.header.notes : t('global.noNotesWasLeft')}
              </div>
            </div>
          </div>
          <div className="col-span-5 rounded-md border p-4">
            <Label>{t(`${kind}s.single.totalSummary`)}</Label>
            <div
              className={cn(
                'flex flex-col gap-y-3 py-4',
                '[&_[data-slot=label]]:text-muted-foreground [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-normal',
                '[&_[data-slot=label-value]]:text-lg [&_[data-slot=label-value]]:font-medium',
              )}
            >
              <div className="flex items-center justify-between">
                <Label>{t('global.subTotal')}</Label>
                <Label data-slot="label-value">{currency(invoice.header.amount)}</Label>
              </div>
              <div className="flex items-center justify-between">
                <Label>{t('global.discount')}</Label>
                <Label data-slot="label-value">
                  {invoice.header.discount.type === 'percentage' && (
                    <span className="text-muted-foreground text-xs">{currency(invoice.header.amount * (invoice.header.discount.value / 100))}</span>
                  )}
                  {invoice.header.discount.type === 'fixed' ? currency(invoice.header.discount.value) : `${invoice.header.discount.value}%`}
                </Label>
              </div>
              <div className="flex items-center justify-between">
                <Label>{t('global.tax')}</Label>
                <Label data-slot="label-value">{currency(invoice.header.tax)}</Label>
              </div>
              <Separator />
              <div className="flex items-center justify-between">
                <Label>{t('global.total')}</Label>
                <Label data-slot="label-value">{currency(invoice.header.total)}</Label>
              </div>
              {invoice.header.due_on && (
                <>
                  <div className="flex items-center justify-between">
                    <Label>{t('global.totalPaymentApplied')}</Label>
                    <Label data-slot="label-value">{currency(invoice.header.total - invoice.header.amount_due)}</Label>
                  </div>
                  <div className="flex items-center justify-between">
                    <Label>{t('global.balance')}</Label>
                    <Label data-slot="label-value">{currency(invoice.header.amount_due)}</Label>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
      {/* side panel with summary */}
      <div className="col-span-4 flex flex-col gap-y-3 rounded-lg border p-3">
        {isInvoice && <StatusBadge type="paid" variant="alert" prefix={`${t('invoices.paidStatus')}:`} status={invoice.header.paid_status} />}
        <Label>{t(`${kind}s.single.description`)}</Label>
        <Separator />
        <div className="flex items-center justify-between">
          <Label className="text-lg">{currency(invoice.header.total)}</Label>
          {isInvoice && (
            <Select name="paid_status" defaultValue={'0'} value={invoice.header.paid_status} required disabled={invoice.header.status === 'void'}>
              <SelectTrigger className="w-46">
                <SelectValue placeholder="Paid status" />
              </SelectTrigger>
              <SelectContent className="w-46">
                {PaidStatuses.map((status) => (
                  <SelectItem key={status} value={status}>
                    {t(`global.paidStatuses.${status}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </div>
        <Separator />
        <div className="flex items-center gap-x-1 text-sm">
          <UserPen size={14} />
          <span className="font-medium">{t('global.createdBy')}:</span>
          <span className="text-muted-foreground">Jane Doe</span>
        </div>
        {invoice.header.due_on && (
          <div className="flex items-center gap-x-1 text-sm">
            <Calendar1 size={14} />
            <span className="font-medium">{t('global.dueDate')}:</span>
            <span className="text-muted-foreground">{format(invoice.header.due_on, 'PPP')}</span>
          </div>
        )}
        <div className="flex items-center gap-x-1 text-sm">
          <CircleDollarSignIcon size={14} />
          <span className="font-medium">{t('global.currency')}:</span>
          <span className="text-muted-foreground">Domincan Peso</span>
        </div>
        {isInvoice && (
          <div className="flex items-center gap-x-1 text-sm">
            <CreditCardIcon size={14} />
            <span className="font-medium">{t('global.paymentSummary')}</span>
          </div>
        )}
        {invoice.header.due_on !== null && <span className="text-muted-foreground -m-1.5 block px-1.5 text-sm">{t('global.noAvailable.yet')}</span>}
        <PaymentSummary paymentData={invoice.header.payment} />
        <Separator />
        <div className="px-1 py-4">
          <ol
            className={cn(
              'relative flex flex-col justify-between gap-y-2 border-s border-gray-200',
              '[&_li]:ms-4 [&_li]:flex [&_li]:items-center [&_li]:justify-between [&_li]:text-sm',
              '[&_li>div]:flex [&_li>div]:items-center [&_li>div]:gap-x-1 [&_svg]:absolute [&_svg]:-start-2.5 [&_svg]:size-5',
              '[&_svg]:fill-green-400 [&_svg]:stroke-white',
              '[&_[data-status=pending]]:text-muted-foreground [&_[data-status=pending]]:fill-gray-200',
              '[&_[data-slot=label]]:text-muted-foreground [&_[data-slot=label]]:text-sm/1 [&_[data-slot=label]]:font-normal',
            )}
          >
            <li>
              <div>
                <CircleCheckIcon />
                <div>{capitalize(kind)} created</div>
              </div>
              <Label>{format(invoice.header.date, 'P')}</Label>
            </li>
            <li>
              <div>
                <CircleCheckIcon />
                <div>{capitalize(kind)} sent</div>
              </div>
              <Label>{format(invoice.header.date, 'P')}</Label>
            </li>
            {isInvoice && (
              <li>
                <div>
                  <CircleCheckIcon data-status="pending" />
                  <div>Invoice paid</div>
                </div>
                <Label>{format(invoice.header.date, 'P')}</Label>
              </li>
            )}
          </ol>
        </div>
      </div>
    </div>
  );
}
