import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { AlertTriangle, FileWarning, ShieldOff } from 'lucide-react';
import React from 'react';

type ErrorProps = {
  status: number;
  message?: string;
};

const statusConfig: Record<
  number,
  {
    title: string;
    icon: React.ReactNode;
    description: string;
  }
> = {
  403: {
    title: 'Forbidden',
    icon: <ShieldOff className="h-12 w-12 text-yellow-500" />,
    description: 'You don’t have permission to access this page.',
  },
  404: {
    title: 'Page Not Found',
    icon: <FileWarning className="text-muted-foreground h-12 w-12" />,
    description: "The page you are looking for doesn't exist.",
  },
  500: {
    title: 'Server Error',
    icon: <AlertTriangle className="h-12 w-12 text-red-500" />,
    description: 'Something went wrong on our end. Please try again later.',
  },
};

export default function Index({ status, message }: ErrorProps) {
  const config = statusConfig[status] ?? {
    title: `Error ${status}`,
    icon: <AlertTriangle className="h-12 w-12 text-red-500" />,
    description: message || 'An unexpected error occurred.',
  };
  return (
    <div className="bg-background flex min-h-screen items-center justify-center px-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardContent className="p-8">
          <div className="flex flex-col items-center space-y-4 text-center">
            {config.icon}
            <h1 className="text-2xl font-semibold">{config.title}</h1>
            <p className="text-muted-foreground text-sm">{config.description}</p>
            <Button variant="outline" onClick={() => window.history.back()}>
              Go Back
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
