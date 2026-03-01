import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';

interface Props {
  processing: boolean;
  idleTitle?: string;
  processingTitle?: string;
  handleOnClick: () => Promise<void>;
}

export default function ActionButton({ processing, idleTitle, processingTitle, handleOnClick }: Props) {
  return (
    <Button onClick={handleOnClick} disabled={processing}>
      {processing ? (
        <span className="flex items-center gap-2">
          <Spinner />
          {processingTitle || 'Generating…'}
        </span>
      ) : (
        idleTitle || 'Generate report'
      )}
    </Button>
  );
}
