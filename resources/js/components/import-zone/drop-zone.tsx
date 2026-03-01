import { Replacements } from '@/types';
import { HardDriveUpload } from 'lucide-react';

type Props = {
  handleZoneClick: () => void;
  handleFileSelect: (file: File | null) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export function DropZone({ handleZoneClick, handleFileSelect, t }: Props) {
  // 1. Prevent browser from opening the file
  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  // 2. Handle the drop
  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    alert('HAHA');
    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      handleFileSelect(files[0]);
    }
  };

  return (
    <div
      className="flex w-full items-center justify-center"
      onDragOver={handleDragOver}
      onDragEnter={handleDragEnter}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      onClick={handleZoneClick}
    >
      <div className="bg-neutral-secondary-medium border-default-strong hover:bg-neutral-tertiary-medium flex h-64 w-full cursor-pointer flex-col items-center justify-center rounded-md border border-dashed">
        <div className="text-body flex flex-col items-center justify-center pt-5 pb-6">
          <HardDriveUpload className="size-8" />
          <p className="mb-2 text-sm">
            <span className="font-semibold">{t('global.import.upload')}</span> {t('global.import.dragAndDrop')}
          </p>
          <p className="text-xs">CSV or TXT (MAX. 10MB)</p>
        </div>
      </div>
    </div>
  );
}
