import { cn } from '@/lib/utils';

type ChipValue = string | number;

export type ChipOption<T extends ChipValue = string> = {
  label: string;
  value: T;
};

type ChipSelectorProps<T extends ChipValue = string> = {
  options: ChipOption<T>[];
  selectedValues: T[];
  onToggle: (value: T) => void;
  className?: string;
  chipClassName?: string;
};

export default function ChipSelector<T extends ChipValue = string>({
  options,
  selectedValues,
  onToggle,
  className,
  chipClassName,
}: ChipSelectorProps<T>) {
  return (
    <div className={cn('flex flex-wrap gap-2', className)}>
      {options.map((option) => {
        const isSelected = selectedValues.includes(option.value);

        return (
          <button
            key={String(option.value)}
            type="button"
            aria-pressed={isSelected}
            onClick={() => onToggle(option.value)}
            className={cn(
              'rounded-full border px-3 py-1.5 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
              isSelected
                ? 'border-primary bg-primary text-primary-foreground hover:bg-primary/90'
                : 'border-border bg-background text-muted-foreground hover:bg-accent hover:text-accent-foreground',
              chipClassName,
            )}
          >
            {option.label}
          </button>
        );
      })}
    </div>
  );
}
