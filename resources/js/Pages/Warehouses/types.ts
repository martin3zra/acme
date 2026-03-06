export interface Warehouse {
  id: number;
  uuid: string;
  code: string;
  name: string;
  address?: string;
  description?: string;
  status: string;
  created_at: string;
  updated_at: string;
}
