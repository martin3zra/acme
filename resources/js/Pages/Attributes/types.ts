export interface Attribute {
  id: number;
  uuid: string;
  name: string;
  type: string;
  display_name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export type AttributeForm = {
  name: string;
  type: string;
  display_name: string;
  description: string;
};
