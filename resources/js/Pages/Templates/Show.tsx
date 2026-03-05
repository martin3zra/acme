import InputError from '@/components/input-error';
import { TemplateEditor } from '@/components/templates/template-editor';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useHeader } from '@/composables/use-headers';
import AppLayout from '@/layouts/app-layout';
import { BreadcrumbItem, PageProps } from '@/types';
import { Link, router, useForm } from '@inertiajs/react';
import { FormEvent, useEffect, useMemo, useState } from 'react';

type TemplateDetails = {
  id: number;
  uuid: string;
  name: string;
  description: string;
  status: string;
};

type TemplateVersion = {
  id: number;
  uuid: string;
  template_id: number;
  version_number: number;
  layout_json: unknown;
  status: string;
  notes: string;
  created_at?: string;
};

const defaultLayoutObject = {
  page_size: 'A4',
  page_format: 'P',
  margins: { top: 15, right: 15, bottom: 15, left: 15 },
  elements: [
    {
      type: 'text',
      x: 15,
      y: 20,
      width: 120,
      height: 8,
      content: '{{business.name}}',
      font_size: 16,
      bold: true,
    },
  ],
};

const toPrettyJSON = (value: unknown): string => {
  if (typeof value === 'string') {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }

  try {
    return JSON.stringify(value ?? defaultLayoutObject, null, 2);
  } catch {
    return JSON.stringify(defaultLayoutObject, null, 2);
  }
};

export default function Show({ auth, csrf_token, template, versions }: PageProps<{ template: TemplateDetails; versions: TemplateVersion[] }>) {
  const latestVersion = versions[0];
  const initialLayout = useMemo(() => toPrettyJSON(latestVersion?.layout_json), [latestVersion?.layout_json]);
  const { headers } = useHeader();

  const [previewUrl, setPreviewUrl] = useState<string>('');
  const [previewLoading, setPreviewLoading] = useState<boolean>(false);
  const [previewError, setPreviewError] = useState<string>('');

  const { data, setData, post, processing, errors } = useForm({
    name: template.name ?? '',
    description: template.description ?? '',
    layout_json: initialLayout,
  });

  const breadcrumbs: BreadcrumbItem[] = [
    {
      title: 'Templates',
      href: '/admin/templates',
    },
    {
      title: template.name,
      href: `/admin/templates/${template.id}`,
    },
  ];

  useEffect(() => {
    setData({
      name: template.name ?? '',
      description: template.description ?? '',
      layout_json: initialLayout,
    });
  }, [initialLayout, setData, template.description, template.name]);

  useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  const onSaveDraft = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    post(`/admin/templates/${template.id}/update`, {
      preserveScroll: true,
    });
  };

  const onPublish = () => {
    router.post(
      `/admin/templates/${template.id}/publish`,
      {
        template_id: template.id,
        notes: '',
      },
      { preserveScroll: true },
    );
  };

  const onPreview = async () => {
    setPreviewLoading(true);
    setPreviewError('');

    try {
      let previewLayoutJSON = data.layout_json;
      let previewData: unknown = {};

      try {
        const parsedLayout = JSON.parse(data.layout_json);
        if (parsedLayout && typeof parsedLayout === 'object' && 'template' in parsedLayout) {
          previewLayoutJSON = JSON.stringify((parsedLayout as { template: unknown }).template ?? {});
          const parsedData = (parsedLayout as { data?: unknown }).data;
          previewData = parsedData ?? {};
        }
      } catch {
      }

      const response = await fetch(`/admin/templates/${template.id}/preview`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/pdf, application/json',
          ...headers.headers,
        },
        body: JSON.stringify({
          template_id: template.id,
          data: previewData,
          layout_json: previewLayoutJSON,
        }),
      });

      const contentType = response.headers.get('content-type') ?? '';

      if (!response.ok || !contentType.includes('application/pdf')) {
        let backendMessage = '';
        if (contentType.includes('application/json')) {
          try {
            const payload = (await response.json()) as { error?: string; status?: string };
            backendMessage = payload.error ?? payload.status ?? '';
          } catch {
          }
        }

        if (contentType.includes('text/html')) {
          throw new Error(backendMessage || 'Preview returned HTML instead of PDF (likely validation/auth redirect).');
        }
        throw new Error(backendMessage || 'Failed to render preview');
      }

      const blob = await response.blob();
      const nextUrl = URL.createObjectURL(blob);

      setPreviewUrl((current) => {
        if (current) {
          URL.revokeObjectURL(current);
        }
        return nextUrl;
      });
    } catch (error) {
      setPreviewError(error instanceof Error ? error.message : 'Failed to render preview');
    } finally {
      setPreviewLoading(false);
    }
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold">Template Editor</h2>
          <Button asChild variant="outline">
            <Link href="/admin/templates">Back to Templates</Link>
          </Button>
        </div>

        <form onSubmit={onSaveDraft} className="grid gap-6 xl:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Template JSON</CardTitle>
              <CardDescription>Edit metadata and layout JSON for this template draft.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Input value={data.name} onChange={(event) => setData('name', event.target.value)} placeholder="Template name" />
                <InputError message={errors.name} />
              </div>

              <div className="space-y-2">
                <Input
                  value={data.description}
                  onChange={(event) => setData('description', event.target.value)}
                  placeholder="Description (optional)"
                />
                <InputError message={errors.description} />
              </div>

              <TemplateEditor id="layout-json" value={data.layout_json} onChange={(value) => setData('layout_json', value)} rows={24} />
              <InputError message={errors.layout_json} />

              <div className="flex flex-wrap gap-2">
                <Button type="submit" disabled={processing}>
                  {processing ? 'Saving...' : 'Save Draft'}
                </Button>
                <Button type="button" variant="outline" onClick={onPublish}>
                  Publish
                </Button>
                <Button type="button" variant="secondary" onClick={onPreview} disabled={previewLoading}>
                  {previewLoading ? 'Rendering...' : 'Preview PDF'}
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Preview</CardTitle>
              <CardDescription>Generates a PDF from the current published rendering pipeline.</CardDescription>
            </CardHeader>
            <CardContent>
              {previewError && <p className="text-sm text-red-600">{previewError}</p>}
              {!previewError && !previewUrl && <p className="text-muted-foreground text-sm">Click "Preview PDF" to render the document.</p>}
              {previewUrl && (
                <iframe
                  title="Template preview"
                  src={previewUrl}
                  className="h-190 w-full rounded-md border"
                />
              )}
            </CardContent>
          </Card>
        </form>

        <Card>
          <CardHeader>
            <CardTitle>Version History</CardTitle>
            <CardDescription>Latest versions first.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Version</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Notes</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {versions.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={3} className="text-muted-foreground h-20 text-center">
                        No versions found.
                      </TableCell>
                    </TableRow>
                  ) : (
                    versions.map((version) => (
                      <TableRow key={version.id}>
                        <TableCell>v{version.version_number}</TableCell>
                        <TableCell className="capitalize">{version.status}</TableCell>
                        <TableCell>{version.notes || '-'}</TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}
