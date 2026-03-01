import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { CheckCircle, FilePlus } from 'lucide-react';
import { Button } from './ui/button';

interface EmptyStateProps {
  title: string;
  description: string;
  variant?: 'positive' | 'action';
  actionLabel?: string;
  onAction?: () => void;
}
export function EmptyState({ title, description, variant = 'action', actionLabel, onAction }: EmptyStateProps) {
  return (
    <Empty className="border border-dashed">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          {variant === 'positive' ? <CheckCircle className="h-10 w-10 text-green-500" /> : <FilePlus className="h-10 w-10 text-gray-400" />}
        </EmptyMedia>
        <EmptyTitle>{title} </EmptyTitle>
        <EmptyDescription>{description}</EmptyDescription>
      </EmptyHeader>
      {variant === 'action' && actionLabel && onAction && (
        <EmptyContent>
          <Button variant="outline" size="sm" onClick={onAction}>
            {actionLabel}
          </Button>
        </EmptyContent>
      )}
    </Empty>
  );
}
