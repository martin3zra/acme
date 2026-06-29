import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { router } from '@inertiajs/react';

export type TransferStatus = 'requested' | 'in_transit' | 'received' | 'cancelled';

function TransferAction({
  uuid,
  action,
  label,
  title,
  body,
  variant = 'default',
}: {
  uuid: string;
  action: 'dispatch' | 'receive' | 'cancel';
  label: string;
  title: string;
  body: string;
  variant?: 'default' | 'outline' | 'destructive';
}) {
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const run = () => router.put(`/inventories/transfers/${uuid}/${action}`, {}, { ...headers, preserveScroll: true });

  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button size="sm" variant={variant}>
          {label}
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{body}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{t('transfers.confirm.dismiss')}</AlertDialogCancel>
          <AlertDialogAction onClick={run}>{t('transfers.confirm.confirm')}</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

// TransferActions renders the lifecycle buttons valid for the given status.
export function TransferActions({ uuid, status }: { uuid: string; status: TransferStatus }) {
  const t = useTranslation().trans;

  if (status === 'requested') {
    return (
      <div className="flex justify-end gap-2">
        <TransferAction
          uuid={uuid}
          action="dispatch"
          label={t('transfers.actions.dispatch')}
          title={t('transfers.confirm.dispatchTitle')}
          body={t('transfers.confirm.dispatchBody')}
        />
        <TransferAction
          uuid={uuid}
          action="cancel"
          label={t('transfers.actions.cancel')}
          title={t('transfers.confirm.cancelTitle')}
          body={t('transfers.confirm.cancelBody')}
          variant="outline"
        />
      </div>
    );
  }
  if (status === 'in_transit') {
    return (
      <div className="flex justify-end">
        <TransferAction
          uuid={uuid}
          action="receive"
          label={t('transfers.actions.receive')}
          title={t('transfers.confirm.receiveTitle')}
          body={t('transfers.confirm.receiveBody')}
        />
      </div>
    );
  }
  return <span className="text-muted-foreground">—</span>;
}

export const statusVariant: Record<TransferStatus, 'default' | 'secondary' | 'destructive' | 'outline'> = {
  requested: 'secondary',
  in_transit: 'default',
  received: 'outline',
  cancelled: 'destructive',
};
