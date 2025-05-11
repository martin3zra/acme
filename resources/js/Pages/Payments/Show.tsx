import { StatusBadge } from '@/components/status-badge';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { useNumber } from '@/composables/use-number';
import { cn, isNotEmpty } from '@/lib/utils';
import { Auth, PaymentWithLines } from '@/types';
import { format, formatDate } from 'date-fns/format';
import { CircleCheckIcon, CircleDollarSignIcon, CreditCardIcon, UserPen } from 'lucide-react';

type Props = {
  payment: PaymentWithLines;
  auth: Auth;
};
export default function Show({ payment, auth }: Props) {
  const { currency } = useNumber();
  return (
    <div className="grid grid-cols-12 gap-x-4">
      <div className="col-span-12 pb-6">
        <Separator />
        <div className="flex justify-between py-6 [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-medium">
          <div className="col-span-6 flex items-center gap-x-6 [&>div]:flex [&>div]:gap-x-2">
            <div>
              <Label>Payment</Label>
              <Label>#{payment.header.number}</Label>
            </div>
            {/* <div>
              <Label>NCF</Label>
              <Label>{payment.header.ncf}</Label>
            </div> */}
          </div>
          <div className="col-span-6 flex items-center gap-x-2 [&_[data-slot=label]]:font-normal">
            <Label>Date</Label>
            <Label className="">{formatDate(payment.header.date, 'dd-MM-yyyy')}</Label>
          </div>
        </div>
        <Separator />
      </div>
      <div className="col-span-9 grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4">
        {/* header */}
        <div className="col-span-12 grid grid-cols-12">
          <div className="col-span-6">
            <div className="w-56">
              <Label className="font-bold">Pay to:</Label>
              <div className="pt-2">
                <span className="text-sm font-semibold">{auth.company.name}</span>
                <address className="text-muted-foreground text-sm font-normal">{auth.company.address}</address>
              </div>
              <div className="text-sm font-medium">{payment.header.customer.email}</div>
            </div>
          </div>
          <div className="col-span-6 place-items-end">
            <div className="w-56">
              <Label className="font-bold">Payment from:</Label>
              <div className="pt-2">
                <span className="text-sm font-semibold">{payment.header.customer.name}</span>
                <address className="text-muted-foreground text-sm font-normal">{payment.header.customer.address}</address>
              </div>
              <div className="text-sm font-medium">{payment.header.customer.email}</div>
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
                  Invoice Number
                </th>
                <th scope="col" className="text-start">
                  Date
                </th>
                <th scope="col" data-format="number" className="!min-w-20">
                  Amount
                </th>
                <th scope="col" data-format="number" className="!min-w-20">
                  Balance
                </th>
                <th scope="col" className="text-start">
                  PayDate
                </th>
                <th scope="col" data-format="number" className="!min-w-20">
                  Payment
                </th>
                <th scope="col" className="text-start">
                  Paid Status
                </th>
              </tr>
            </thead>
            <tbody>
              {payment.lines.map((line) => (
                <tr key={line.id}>
                  <td>{line.invoice.number}</td>
                  <td>{formatDate(line.invoice.date, 'dd-MM-yyyy')}</td>
                  <td data-format="number" className="!min-w-20">
                    {currency(line.invoice.amount)}
                  </td>
                  <td data-format="number" className="!min-w-20">
                    {currency(line.invoice.amount_due)}
                  </td>
                  <td>{formatDate(payment.header.date, 'dd-MM-yyyy')}</td>
                  <td data-format="number"> {currency(line.payment)}</td>
                  <td>
                    <StatusBadge type="paid" status={line.invoice.paid_status} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {/* footer */}
        <div className="col-span-12 grid grid-cols-12">
          <div className="col-span-8">
            <div className="max-w-sm rounded-md border p-4">
              <Label className="text-sm/6 font-medium">Notes:</Label>
              <div className="text-muted-foreground text-sm">{isNotEmpty(payment.header.notes) ? payment.header.notes : 'No notes left'}</div>
            </div>
          </div>
          <div className="col-span-4 rounded-md border p-4">
            <Label>Payment summary</Label>
            <div
              className={cn(
                'flex flex-col gap-y-3 py-4',
                '[&_[data-slot=label]]:text-muted-foreground [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-normal',
                '[&_[data-slot=label-value]]:text-lg [&_[data-slot=label-value]]:font-medium',
              )}
            >
              <div className="flex items-center justify-between">
                <Label>Sub-total</Label>
                <Label data-slot="label-value">{currency(payment.header.amount)}</Label>
              </div>
              <div className="flex items-center justify-between">
                <Label>Discount</Label>
                <Label data-slot="label-value">
                  {currency(0)}
                  {/* {payment.header.discount.type === 'percentage' && (
                    <span className="text-muted-foreground text-xs">{currency(payment.header.amount * (payment.header.discount.value / 100))}</span>
                  )}
                  {payment.header.discount.type === 'fixed' ? currency(payment.header.discount.value) : `${payment.header.discount.value}%`} */}
                </Label>
              </div>
              <Separator />
              <div className="flex items-center justify-between">
                <Label>Total</Label>
                <Label data-slot="label-value">{currency(payment.header.amount)}</Label>
              </div>
            </div>
          </div>
        </div>
      </div>
      {/* side panel with summary */}
      <div className="col-span-3 flex flex-col gap-y-3 rounded-lg border p-3">
        {/* <StatusBadge type="paid" variant="alert" prefix="Paid status:" status={payment.header.paid_status} /> */}
        <Label>Details:</Label>
        <Separator />
        <div className="flex items-center gap-x-1 text-sm">
          <UserPen size={14} />
          <span className="font-medium">Created by:</span>
          <span className="text-muted-foreground">Jane Doe</span>
        </div>

        <div className="flex items-center gap-x-1 text-sm">
          <CircleDollarSignIcon size={14} />
          <span className="font-medium">Currency:</span>
          <span className="text-muted-foreground">Domincan Peso</span>
        </div>
        <div className="flex items-center gap-x-1 text-sm">
          <CreditCardIcon size={14} />
          <span className="font-medium">Payment summary</span>
        </div>

        {/* <PaymentSummary paymentData={payment.header.payment} /> */}
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
                <div>payment created</div>
              </div>
              <Label>{format(payment.header.date!, 'P')}</Label>
            </li>
            <li>
              <div>
                <CircleCheckIcon />
                <div>payment sent</div>
              </div>
              <Label>{format(payment.header.date!, 'P')}</Label>
            </li>
            <li>
              <div>
                <CircleCheckIcon data-status="pending" />
                <div>payment paid</div>
              </div>
              <Label>{format(payment.header.date!, 'P')}</Label>
            </li>
          </ol>
        </div>
      </div>
    </div>
  );
}
