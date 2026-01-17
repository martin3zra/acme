import { Button } from '@/components/ui/button';

interface Props {
  processing: boolean;
  handleOnClick: () => Promise<void>;
}

export default function ActionButton({ processing, handleOnClick }: Props) {
  return (
    <Button onClick={handleOnClick} disabled={processing}>
      {processing ? (
        <span className="flex items-center gap-2">
          <svg className="h-4 w-4 animate-spin text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4l3-3-3-3v4a8 8 0 00-8 8h4z" />
          </svg>
          Generating…
        </span>
      ) : (
        'Generate report'
      )}
    </Button>
  );
}
