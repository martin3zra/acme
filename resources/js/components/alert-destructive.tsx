import { AlertCircle } from 'lucide-react';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { useEffect, useState } from 'react';

type AlertDestructiveProps = {
  title?: string;
  description: string;
  duration?: number;
  destroyable?: boolean;
  onDestroy?: () => void;
};

export function AlertDestructive({ title = 'Whoops', description, duration = 3000, destroyable = true, onDestroy }: AlertDestructiveProps) {
  const [isVisible, setIsVisible] = useState(true);

  useEffect(() => {
    if (isVisible && destroyable) {
      const timer = setTimeout(() => {
        setIsVisible(false);
        if (onDestroy) onDestroy();
      }, duration);

      return () => clearTimeout(timer); // Clear timeout if component unmounts or duration changes
    }
  }, [isVisible, duration, onDestroy, destroyable]);

  if (!isVisible) {
    return null;
  }

  return (
    <Alert variant="destructive" className="border-red-400 bg-red-100/50">
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{description}</AlertDescription>
    </Alert>
  );
}
