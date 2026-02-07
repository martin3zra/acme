import { Download } from 'lucide-react';
import { Button } from '../ui/button';

export function DownloadableSampleSection() {
  return (
    <div className="mb-6 space-y-2">
      <p className="text-muted-foreground text-sm">Use our sample file to ensure the correct format.</p>

      <div className="flex space-x-3">
        <Button variant="outline" size="sm">
          <Download className="mr-2 h-4 w-4" />
          Download sample CSV
        </Button>
        <Button variant="outline" size="sm">
          <Download className="mr-2 h-4 w-4" />
          Download sample TXT
        </Button>
      </div>
    </div>
  );
}
