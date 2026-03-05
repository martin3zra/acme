import HeadingSmall from '@/components/heading-small';
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
import { FormEvent } from 'react';

type TemplateSummary = {
  id: number;
  uuid: string;
  name: string;
  description: string;
  status: string;
  updated_at?: string;
};

const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Templates',
    href: '/admin/templates',
  },
];

const defaultLayout = JSON.stringify(
  {
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
  },
  null,
  2,
);

export default function Index({ auth, templates, csrf_token }: PageProps<{ templates: TemplateSummary[] | null }>) {
  const safeTemplates = templates ?? [];
  const { headers } = useHeader();

  const { data, setData, post, processing, errors, reset } = useForm({
    name: '',
    description: '',
    layout_json: defaultLayout,
  });

  const onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    // Use Inertia router.post() which properly handles redirects
    router.post(
      '/admin/templates',
      {
        name: data.name,
        description: data.description,
        layout_json: data.layout_json,
      },
      {
        ...headers,
        preserveScroll: true,
      },
    );
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        <HeadingSmall title="PDF Templates" description="Create and manage JSON-based PDF templates." />

        <Card>
          <CardHeader>
            <CardTitle>Create Template</CardTitle>
            <CardDescription>Use a name, optional description, and a JSON layout definition.</CardDescription>
          </CardHeader>
          <CardContent>
            <form className="space-y-4" onSubmit={onSubmit}>
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Input placeholder="Template name" value={data.name} onChange={(event) => setData('name', event.target.value)} />
                  <InputError message={errors.name} />
                </div>
                <div className="space-y-2">
                  <Input
                    placeholder="Description (optional)"
                    value={data.description}
                    onChange={(event) => setData('description', event.target.value)}
                  />
                  <InputError message={errors.description} />
                </div>
              </div>

              <TemplateEditor id="create-layout-json" value={data.layout_json} onChange={(value) => setData('layout_json', value)} rows={14} />
              <InputError message={errors.layout_json} />

              <div className="flex justify-end">
                <Button type="submit" disabled={processing}>
                  {processing ? 'Creating...' : 'Create Template'}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Templates</CardTitle>
            <CardDescription>Open a template to edit JSON, publish, and preview PDF.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead className="w-30 text-right">Action</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {safeTemplates.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={4} className="text-muted-foreground h-24 text-center">
                        No templates yet.
                      </TableCell>
                    </TableRow>
                  ) : (
                    safeTemplates.map((template) => (
                      <TableRow key={template.id}>
                        <TableCell className="font-medium">{template.name}</TableCell>
                        <TableCell className="capitalize">{template.status}</TableCell>
                        <TableCell>{template.description || '-'}</TableCell>
                        <TableCell className="text-right">
                          <Button asChild size="sm" variant="outline">
                            <Link href={`/admin/templates/${template.id}`}>Open</Link>
                          </Button>
                        </TableCell>
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
