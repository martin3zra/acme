import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { ErrorResponse } from '@/types';
import axios from 'axios';
import { useEffect, useRef, useState } from 'react';
import { toast } from 'sonner';
import { Progress } from '../ui/progress';
import { Spinner } from '../ui/spinner';
import { DropZonePreview } from './data-preview';
import { DownloadableSampleSection } from './downloadble-sample-section';
import { DropZone } from './drop-zone';
import { ImportErrorTable } from './errors-table';

type ImportState =
  | 'idle'
  | 'previewing'
  | 'initializing'
  | 'uploading'
  | 'processing'
  | 'done'
  | 'error'
  | 'uploaded'
  | 'queued'
  | 'completed'
  | 'failed';

type ImportCompletedPayload = {
  total: number;
  processed: number;
  success: number;
  failed: number;
  warning: number;
};

type ImportPhase = 'reading_file' | 'normalizing_encoding' | 'mapping_columns' | 'importing_rows';

interface ImportProgress {
  processed: number;
  total: number;
}

interface ImportError {
  row: number;
  error: string;
  raw: string;
}

// Why chunking?
// - Avoid request size limits
// - Allow retries
// - Avoid memory spikes
const CHUNK_SIZE = 5 * 1024 * 1024; // 5MB
const MAX_RETRIES = 3;
const MAX_PREVIEW_ROWS = 8;
const IMPORT_TOAST_ID = 'import-progress';

export type CsvPreview = {
  headers: string[];
  rows: string[][];
};

const DELIMITERS = ['auto', ',', ';', '\t', '|'] as const;
type Delimiter = (typeof DELIMITERS)[number];

type Props = {
  source: 'items' | 'customers';
  openImportDrawer: boolean;
  setImportDrawer: React.Dispatch<React.SetStateAction<boolean>>;
};

export function ImportDrawer({ source, openImportDrawer, setImportDrawer }: Props) {
  const headers = useHeader().headers;
  const t = useTranslation().trans;
  const [fileText, setFileText] = useState<string>('');
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<CsvPreview | null>(null);
  const [totalRows, setTotalRows] = useState<number>(0);
  const [encoding, setEncoding] = useState<string | null>(null);
  const [state, setState] = useState<ImportState>('idle');
  const [phase, setPhase] = useState<ImportPhase | null>(null);
  const [progress, setProgress] = useState<ImportProgress | null>(null);
  const [errors, setErrors] = useState<ImportError[]>([]);
  const [delimiter, setDelimiter] = useState<Delimiter>(',');
  const [detectedDelimiter, setDetectedDelimiter] = useState<Delimiter>(',');
  const sseRef = useRef<EventSource | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const importDisabled = !file || state === 'initializing' || state === 'uploading' || state === 'processing' || state === 'completed';

  function updateImportToast(event: { type: string; data?: any }) {
    switch (event.type) {
      case 'phase': {
        setState('processing');
        setPhase(event.data);
        break;
      }

      case 'queued':
        toast.loading('Import queued…', {
          id: IMPORT_TOAST_ID,
          duration: Infinity,
        });
        break;

      case 'processing':
        toast.loading(`Processing ${event.data.processed_rows} / ${event.data.total_rows}`, {
          id: IMPORT_TOAST_ID,
          duration: Infinity,
        });
        break;

      case 'progress': {
        setProgress({
          processed: event.data.processed,
          total: event.data.total,
        });
        break;
      }

      case 'completed': {
        handleOnClear();
        const data: ImportCompletedPayload = JSON.parse(event.data);
        const { total, success, failed, processed, warning } = data;
        toast.success(
          warning > 0
            ? `Imported ${success} rows of ${total} with ${warning} warning${warning > 1 ? 's' : ''}`
            : `Imported ${success} of ${total} rows successfully`,
        );

        if (failed > 0) {
          toast.error(`${failed} of ${total} rows failed to import`);
        }
        break;
      }

      case 'failed':
        setState('failed');
        setProgress(null);
        setPhase(null);
        toast.error('Import failed', {
          id: IMPORT_TOAST_ID,
          duration: 8000,
        });
        break;
      default: {
        console.warn('Unknown SSE event', event);
      }
    }
  }

  function connectToSSE(importId: string) {
    // Always close an existing connection first
    if (sseRef.current) {
      sseRef.current.close();
      sseRef.current = null;
    }
    // Safety: ensure we are in processing state
    setState('processing');

    const es = new EventSource(`http://192.168.100.250:8090/sse/imports/${importId}`);
    sseRef.current = es;

    es.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data);

        updateImportToast(event);

        if (event.type === 'completed' || event.type === 'failed') {
          es.close();
          cleanupSSE();
        }
      } catch (err) {
        console.error('Invalid SSE payload', err);
      }
    };

    es.onerror = () => {
      // Network error or backend restart
      console.error('SSE connection error');
      es.close();

      // Optional: fallback to polling here
      setState('error');
      cleanupSSE();
    };

    return () => {
      es.close();
    };
  }

  function cleanupSSE() {
    if (sseRef.current) {
      sseRef.current.close();
      sseRef.current = null;
    }
  }

  useEffect(() => {
    return () => {
      cleanupSSE();
    };
  }, []);

  useEffect(() => {
    if (!fileText) return;

    const preview = parseCsvPreview(fileText, delimiter || undefined);

    setPreview(preview);
    setDetectedDelimiter(preview.detectedDelimiter);
  }, [fileText, delimiter]);

  const handleFileSelect = (file: File | null) => {
    if (file === null) {
      return;
    }
    setFile(file);
    setState('previewing');
    previewFile(file);
  };

  const delimiterLabel = (d: string) => {
    switch (d) {
      case '\t':
        return 'Tab';
      case ';':
        return 'Semicolon (;)';
      case '|':
        return 'Pipe (|)';
      case ',':
        return 'Comma (,)';
      default:
        return d;
    }
  };

  function detectDelimiter(line: string): Delimiter {
    let best: Delimiter = ',';
    let maxCount = 0;

    for (const d of DELIMITERS) {
      const count = line.split(d).length - 1;
      if (count > maxCount) {
        maxCount = count;
        best = d;
      }
    }

    return best;
  }

  function parseCsvPreview(text: string, forcedDelimiter?: string): CsvPreview & { detectedDelimiter: Delimiter } {
    const lines = text.split(/\r?\n/).filter(Boolean);

    if (lines.length === 0) {
      return { headers: [], rows: [], detectedDelimiter: ',' };
    }

    const autoDetected = detectDelimiter(lines[0]);
    const delimiter = forcedDelimiter || autoDetected;

    const headers = lines[0].split(delimiter).map((h) => h.trim());
    const columnCount = headers.length;
    setTotalRows(lines.length - 1);
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
    return { headers, rows, detectedDelimiter: autoDetected };
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

      // const preview = parseCsvPreview(text, delimiter);
      setEncoding(detectedEncoding);

      setFileText(text);
      // setPreview(preview);
    };

    // Read raw bytes instead of assuming UTF-8
    reader.readAsArrayBuffer(file.slice(0, 100_000));
  }

  const startImport = async (uploadId: string) => {
    try {
      const res = await axios.post(
        '/imports',
        {
          upload_id: uploadId,
          type: source,
        },
        { ...headers },
      );

      console.log(res.status);

      return res.data as {
        import_id: string;
        status: 'queued';
      };
    } catch (err: any) {
      // 1. Extract the error response from Axios
      const errorData: ErrorResponse = err.response?.data;

      // 2. Handle the toast (using the 'status' or 'error' field from your foundation.ErrorBag)
      toast.error(errorData?.status || 'Import initialization failed');

      // 3. Re-throw to prevent the calling function from continuing
      throw new Error(errorData?.status || 'Failed fetching the data');
    }
  };

  const onStartUpload = async () => {
    if (!file) return;
    try {
      setState('initializing');
      // 🔥 create upload session
      const uploadId = await initUpload(file);

      setState('uploading');

      // 🔥 now start sending bytes
      await uploadFileInChunks(file, uploadId);

      setState('processing');

      const { import_id } = await startImport(uploadId);
      connectToSSE(import_id);
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
          delimiter,
          encoding,
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

    const totalChunks = Math.ceil(file.size / CHUNK_SIZE);

    for (let index = 0; index < totalChunks; index++) {
      const start = index * CHUNK_SIZE;
      const end = Math.min(start + CHUNK_SIZE, file.size);
      const chunk = file.slice(start, end);

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

  const handleOnClear = () => {
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
    setTotalRows(0);
    setPreview(null);
    setFile(null);
    setFileText('');
    setProgress(null);
    setPhase(null);
    setState('idle');
  };

  return (
    <>
      <Sheet open={openImportDrawer} onOpenChange={setImportDrawer}>
        <SheetContent
          side="right"
          onInteractOutside={(e) => e.preventDefault()}
          onPointerDownOutside={(e) => e.preventDefault()}
          onEscapeKeyDown={(e) => e.preventDefault()}
          className="m-4 flex h-[calc(~'(100%_-_var(--spacing)_*_4)_/_3')] w-full flex-col rounded-md sm:max-w-7xl [&>button]:hidden"
        >
          <SheetHeader>
            <SheetTitle>{t(`${source}.import.title`)}</SheetTitle>
            <SheetDescription>{t(`${source}.import.description`)}</SheetDescription>
          </SheetHeader>

          <Separator className="my-4" />
          <div className="flex flex-col px-6">
            {/* Sample file */}
            <DownloadableSampleSection source={source} t={t} />

            {state !== 'idle' && state !== 'previewing' && (
              <>
                <p className="text-muted-foreground text-sm">{t(`global.import.states.${state}`)}</p>
                <div className="space-y-2 py-6">
                  <p className="text-muted-foreground text-sm">{phase && t(`global.import.states.${phase}`)}</p>
                </div>
              </>
            )}

            {state === 'processing' && progress?.total != null && (
              <div className="space-y-1">
                <Progress value={progress.processed != null ? (progress.processed / progress.total) * 100 : 0} />
                <p className="text-muted-foreground text-xs">
                  {progress.processed ?? 0} of {progress.total} rows
                </p>
              </div>
            )}
            <ImportErrorTable errors={errors} />

            <div className="relative max-h-64">
              <div className="mb-3 max-w-sm">
                <label htmlFor="countries" className="text-heading mb-2.5 block text-sm font-medium">
                  {t('global.import.selectAnDelimiter')}
                </label>
                <select
                  value={delimiter}
                  onChange={(e) => setDelimiter(e.target.value as Delimiter)}
                  id="countries"
                  className="border-default-medium text-heading focus:ring-brand focus:border-brand placeholder:text-body block w-full rounded-md border px-3 py-2.5 text-sm shadow-xs"
                >
                  <option value="auto">{t('global.import.delimiters.autoDetect')}</option>
                  <option value=",">{t('global.import.delimiters.comma')}</option>
                  <option value=";">{t('global.import.delimiters.semicolon')}</option>
                  <option value="\t">{t('global.import.delimiters.tab')}</option>
                  <option value="|">{t('global.import.delimiters.pipe')}</option>
                </select>
              </div>
              <input
                type="file"
                ref={fileInputRef}
                accept=".csv,.txt"
                disabled={state !== 'idle' && state !== 'previewing'}
                onChange={(e) => e.target.files && handleFileSelect(e.target?.files[0] ?? null)}
                className="hidden h-0"
              />

              <div className={preview !== null ? 'opacity-100' : 'pointer-events-none opacity-0'}>
                {delimiter === 'auto' && preview && (
                  <p className="text-muted-foreground text-sm">
                    {t('global.import.detectedDelimiter')} <strong>{delimiterLabel(detectedDelimiter)}</strong>
                    <br />
                    <span className="text-xs">{t('global.import.wrongDelimiter')}</span>
                  </p>
                )}
                <DropZonePreview t={t} totalRows={totalRows} preview={preview} encoding={encoding} handleOnClear={handleOnClear} />
              </div>

              <div className={preview === null ? 'opacity-100' : 'pointer-events-none opacity-0'}>
                <DropZone t={t} handleZoneClick={() => fileInputRef.current?.click()} handleFileSelect={handleFileSelect} />
              </div>
            </div>
          </div>

          {/* Footer */}
          <SheetFooter className="border-t pt-4">
            <div className="flex w-full items-center justify-between">
              <p className="text-muted-foreground text-xs">{t('global.import.footnote')}</p>

              <div className="flex gap-2">
                <Button variant="ghost" onClick={() => setImportDrawer(false)}>
                  {t('global.actions.cancel')}
                </Button>

                <Button disabled={importDisabled} onClick={onStartUpload}>
                  {(state === 'uploading' || state === 'processing') && <Spinner />}
                  {t('global.actions.import')}
                </Button>
              </div>
            </div>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    </>
  );
}
