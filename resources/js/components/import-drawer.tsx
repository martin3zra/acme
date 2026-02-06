import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useHeader } from '@/composables/use-headers';
import { ErrorResponse } from '@/types';
import axios from 'axios';
import { Download } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';
import { Progress } from './ui/progress';

type ImportState = 'idle' | 'previewing' | 'initializing' | 'uploading' | 'processing' | 'done' | 'error';

// Why chunking?
// - Avoid request size limits
// - Allow retries
// - Avoid memory spikes
const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB
const MAX_RETRIES = 3;

type Props = {
  openImportDrawer: boolean;
  setImportDrawer: React.Dispatch<React.SetStateAction<boolean>>;
};
export function ImportDrawer({ openImportDrawer, setImportDrawer }: Props) {
  const headers = useHeader().headers;
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<CsvPreview | null>(null);
  const [encoding, setEncoding] = useState<string | null>(null);
  const [state, setState] = useState<ImportState>('idle');
  const [progress, setProgress] = useState<number>(0);

  const handleFileSelect = (file: File | null) => {
    if (file === null) {
      return;
    }
    setFile(file);
    setState('previewing');
    previewFile(file);
  };

  type CsvPreview = {
    headers: string[];
    rows: string[][];
  };

  const MAX_PREVIEW_ROWS = 8;

  function parseCsvPreview(text: string): CsvPreview {
    const lines = text.split(/\r?\n/).filter(Boolean);

    if (lines.length === 0) {
      return { headers: [], rows: [] };
    }

    const delimiter = lines[0].includes('\t') ? '\t' : ',';

    const headers = lines[0].split(delimiter).map((h) => h.trim());
    const columnCount = headers.length;

    const rows = lines.slice(1, MAX_PREVIEW_ROWS + 1).map((line) => {
      const cells = line.split(delimiter).map((cell) => cell.trim());
      // ✅ Normalize row length
      if (cells.length < columnCount) {
        return [...cells, ...Array(columnCount - cells.length).fill('')];
      }
      if (cells.length > columnCount) {
        return cells.slice(0, columnCount);
      }

      return cells;
    });
    return { headers, rows };
  }

  function previewFile(file: File) {
    const reader = new FileReader();

    reader.onload = (e) => {
      const buffer = e.target?.result as ArrayBuffer;
      if (!buffer) return;

      // Try UTF-8 first
      let text = new TextDecoder('utf-8', { fatal: false }).decode(buffer);
      let detectedEncoding = 'UTF-8';

      // If we see replacement chars, fallback to Latin-1
      if (text.includes('�')) {
        text = new TextDecoder('iso-8859-1').decode(buffer);
        detectedEncoding = 'Latin-1';
      }

      const preview = parseCsvPreview(text);
      setEncoding(detectedEncoding);

      setPreview(preview);
    };

    // Read raw bytes instead of assuming UTF-8
    reader.readAsArrayBuffer(file.slice(0, 100_000));
  }

  const onStartUpload = async () => {
    if (!file) return;
    try {
      setState('initializing');
      // 🔥 create upload session
      const uploadId = await initUpload(file);

      // 🔥 now start sending bytes
      await uploadFileInChunks(file, uploadId);
    } catch (err) {
      setState('error');
    }
  };

  const initUpload = async (file: File): Promise<string> => {
    try {
      const response = await axios.post(
        '/uploads/init',
        {
          filename: file.name,
          size: file.size,
          mime: file.type,
        },
        { ...headers },
      );

      // Axios wraps the backend response in a .data object
      // Your Go backend sends the struct decoded by ParseRequest
      return response.data.upload_id;
    } catch (err: any) {
      // 1. Extract the error response from Axios
      const errorData: ErrorResponse = err.response?.data;

      // 2. Handle the toast (using the 'status' or 'error' field from your foundation.ErrorBag)
      toast.error(errorData?.status || 'Upload initialization failed');

      // 3. Re-throw to prevent the calling function from continuing
      throw new Error(errorData?.status || 'Failed fetching the data');
    }
  };

  const uploadFileInChunks = async (file: File, uploadId: string) => {
    if (!file) return;

    setState('uploading');
    const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

    for (let index = 0; index < totalChunks; index++) {
      const start = index * CHUNK_SIZE;
      const end = Math.min(start + CHUNK_SIZE, file.size);
      const chunk = file.slice(start, end);
      setProgress(Math.round(((index + 1) / totalChunks) * 100));

      await uploadChunk({
        chunk,
        uploadId,
        index,
        totalChunks,
        filename: file.name,
      });
    }

    await finalizedUpload(uploadId, file.name);
  };

  const finalizedUpload = async (uploadId: string, filename: string) => {
    await axios.post(
      '/uploads/complete',
      {
        upload_id: uploadId,
        filename: filename,
      },
      { ...headers },
    );

    // setState('processing');
    setState('done');
  };

  const uploadChunk = async (
    data: { chunk: Blob; uploadId: string; index: number; totalChunks: number; filename: string },
    attempt: number = 1,
  ): Promise<void> => {
    const formData = new FormData();
    formData.append('chunk', data.chunk);
    formData.append('upload_id', data.uploadId);
    formData.append('chunk_index', data.index.toString());
    formData.append('total_chunks', data.totalChunks.toString());
    formData.append('filename', data.filename);

    try {
      await axios.post('/uploads/chunks', formData, { headers: { ...headers.headers, ContentType: 'multipart/form-data' } });
    } catch (err) {
      if (attempt >= MAX_RETRIES) throw err;
      await new Promise((r) => setTimeout(r, 500 * attempt));
      return uploadChunk(data, attempt + 1);
    }
  };

  return (
    <>
      <Sheet open={openImportDrawer} onOpenChange={setImportDrawer}>
        <SheetContent
          side="right"
          onInteractOutside={(e) => e.preventDefault()}
          onPointerDownOutside={(e) => e.preventDefault()}
          onEscapeKeyDown={(e) => e.preventDefault()}
          className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-7xl [&>button]:hidden"
        >
          <SheetHeader>
            <SheetTitle>Import Customers</SheetTitle>
            <SheetDescription>Upload a CSV or TXT file to import your data.</SheetDescription>
          </SheetHeader>

          <Separator className="my-4" />
          <div className="flex flex-col px-6">
            {/* Sample file */}
            <div className="space-y-2">
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

            <Separator className="my-4" />

            {/* Upload */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Upload file</label>

              <input
                type="file"
                accept=".csv,.txt"
                disabled={state !== 'idle' && state !== 'previewing'}
                onChange={(e) => e.target.files && handleFileSelect(e.target?.files[0] ?? null)}
              />

              <p className="text-muted-foreground text-xs">Accepted formats: CSV, TXT</p>
            </div>
            {state}
            {'Alfredo'}
            <p className="text-muted-foreground text-sm">
              {state === 'initializing' && 'Preparing upload…'}
              {state === 'uploading' && `Uploading… ${progress}%`}
              {state === 'processing' && 'Processing CSV…'}
              {state === 'done' && 'Import completed'}
            </p>

            {state === 'uploading' && <Progress value={progress} />}

            {/* Preview */}
            {preview && (
              <>
                <Separator className="my-4" />

                <div className="space-y-2">
                  {/* <pre className="bg-muted max-h-64 overflow-auto rounded p-3 text-xs">{preview}</pre> */}
                  {preview && preview.headers.length > 0 && (
                    <div className="space-y-2">
                      <p className="text-sm font-medium">Preview</p>

                      <div className="max-h-64 overflow-auto rounded border">
                        <table className="w-full border-collapse text-xs">
                          <thead className="bg-muted sticky top-0">
                            <tr>
                              {preview.headers.map((header, i) => (
                                <th key={i} className="border-b px-3 py-2 text-left font-medium">
                                  {header}
                                </th>
                              ))}
                            </tr>
                          </thead>

                          <tbody>
                            {preview.rows.map((row, i) => (
                              <tr key={i} className="odd:bg-background even:bg-muted/40">
                                {row.map((cell, j) => (
                                  <td key={j} className="border-b px-3 py-2 whitespace-nowrap">
                                    {cell || <span className="text-muted-foreground">—</span>}
                                  </td>
                                ))}
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                      <p className="text-xs text-green-600">Header row detected • {preview.headers.length} columns found</p>

                      <p className="text-muted-foreground text-xs">Showing first {preview.rows.length} rows</p>
                      {encoding && (
                        <p className="text-muted-foreground text-xs">
                          Encoding detected: <strong>{encoding}</strong>
                          {encoding !== 'UTF-8' && ' (converted to UTF-8)'}
                        </p>
                      )}
                    </div>
                  )}
                </div>
              </>
            )}

            <div className="flex-1" />
          </div>

          {/* Footer */}
          <SheetFooter className="border-t pt-4">
            <div className="flex w-full items-center justify-between">
              <p className="text-muted-foreground text-xs">Large files may take a few minutes to import.</p>

              <div className="flex gap-2">
                <Button variant="ghost" onClick={() => setImportDrawer(false)}>
                  Cancel
                </Button>

                <Button disabled={!file || state !== 'previewing'} onClick={onStartUpload}>
                  Import file
                </Button>
              </div>
            </div>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </>
  );
}
