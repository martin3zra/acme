import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';

type TemplateEditorProps = {
  id?: string;
  label?: string;
  value: string;
  onChange: (value: string) => void;
  rows?: number;
  disabled?: boolean;
  className?: string;
};

export function TemplateEditor({ id, label = 'Layout JSON', value, onChange, rows = 20, disabled = false, className }: TemplateEditorProps) {
  return (
    <div className="space-y-2">
      {label && <Label htmlFor={id}>{label}</Label>}
      <textarea
        id={id}
        value={value}
        rows={rows}
        disabled={disabled}
        onChange={(event) => onChange(event.target.value)}
        className={cn(
          'border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex w-full rounded-md border px-3 py-2 font-mono text-xs shadow-sm focus-visible:ring-1 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50',
          className,
        )}
      />
    </div>
  );
}
