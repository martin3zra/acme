import { DollarSign, FilePlus, FileText, HelpCircle, Users } from 'lucide-react';

export const iconMap = {
  dollar: DollarSign,
  users: Users,
  invoice: FileText,
  estimate: FilePlus,
};

export type IconName = keyof typeof iconMap;

// Fallback icon if name not found
export const fallbackIcon = HelpCircle;
