import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { router } from '@inertiajs/react';
import { Plus, Pencil, Trash2, ChevronRight } from 'lucide-react';
import { useState } from 'react';

interface Attribute {
  id: number;
  uuid: string;
  name: string;
  type: string;
  display_name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export default function Index({ attributes }: PageProps<{ attributes: Attribute[] }>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    type: 'select',
    display_name: '',
    description: '',
  });
  const [editingId, setEditingId] = useState<number | null>(null);

  const handleCreate = () => {
    setFormData({ name: '', type: 'select', display_name: '', description: '' });
    setEditingId(null);
    setOpen(true);
  };

  const handleEdit = (attribute: Attribute) => {
    setFormData({
      name: attribute.name,
      type: attribute.type,
      display_name: attribute.display_name,
      description: attribute.description || '',
    });
    setEditingId(attribute.id);
    setOpen(true);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (editingId) {
      router.put(`/attributes/${editingId}`, formData);
    } else {
      router.post('/attributes', formData);
    }

    setOpen(false);
  };

  const handleDelete = (id: number) => {
    if (confirm('Are you sure?')) {
      router.delete(`/attributes/${id}`);
    }
  };

  return (
    <AppLayout
      title={t('@global.attributes')}
      breadcrumbs={[{ label: t('@global.attributes'), href: '/attributes' }]}
    >
      <div className="space-y-6">
        <div className="flex justify-between items-center">
          <h1 className="text-3xl font-bold">{t('@global.attributes')}</h1>
          <Button onClick={handleCreate}>
            <Plus className="w-4 h-4 mr-2" />
            {t('@global.create')}
          </Button>
        </div>

        {attributes.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500">{t('@global.noDataAvailable')}</p>
          </div>
        ) : (
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('@global.name')}</TableHead>
                  <TableHead>{t('@global.displayName')}</TableHead>
                  <TableHead>{t('@global.type')}</TableHead>
                  <TableHead>{t('@global.description')}</TableHead>
                  <TableHead className="text-right">{t('@global.actions')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {attributes.map((attribute) => (
                  <TableRow key={attribute.id}>
                    <TableCell className="font-medium">{attribute.name}</TableCell>
                    <TableCell>{attribute.display_name}</TableCell>
                    <TableCell>
                      <span className="inline-block px-2 py-1 bg-blue-100 text-blue-800 rounded text-sm">
                        {attribute.type}
                      </span>
                    </TableCell>
                    <TableCell>{attribute.description || '-'}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <a
                          href={`/attributes/${attribute.id}/values`}
                          className="p-2 hover:bg-gray-100 rounded inline-block"
                          title="Manage values"
                        >
                          <ChevronRight className="w-4 h-4" />
                        </a>
                        <button
                          onClick={() => handleEdit(attribute)}
                          className="p-2 hover:bg-gray-100 rounded"
                        >
                          <Pencil className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => handleDelete(attribute.id)}
                          className="p-2 hover:bg-gray-100 rounded text-red-600"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent>
          <SheetHeader>
            <SheetTitle>
              {editingId ? t('@global.edit') : t('@global.create')} {t('@global.attribute')}
            </SheetTitle>
            <SheetDescription>
              {editingId
                ? `${t('@global.update')} attribute settings`
                : `${t('@global.create')} a new attribute`}
            </SheetDescription>
          </SheetHeader>

          <form onSubmit={handleSubmit} className="space-y-6 mt-6">
            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.name')}</label>
              <Input
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="e.g., color, size, length"
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.displayName')}</label>
              <Input
                value={formData.display_name}
                onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
                placeholder="e.g., Color, Shirt Size, Length"
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.type')}</label>
              <select
                value={formData.type}
                onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                className="w-full p-2 border rounded-md"
                required
              >
                <option value="select">Select (Dropdown)</option>
                <option value="text">Text</option>
                <option value="numeric">Numeric</option>
              </select>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">{t('@global.description')}</label>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                className="w-full p-2 border rounded-md"
                placeholder="Optional description"
                rows={3}
              />
            </div>

            <Button type="submit" className="w-full">
              {t('@global.save')}
            </Button>
          </form>
        </SheetContent>
      </Sheet>
    </AppLayout>
  );
}
