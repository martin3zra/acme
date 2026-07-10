import React from 'react';
import { TableCell, TableRow } from '../ui/table';

export const SkeletonRow: React.FC<{ columns: number }> = ({ columns }) => (
  <TableRow className="animate-pulse">
    {Array.from({ length: columns }).map((_, i) => (
      <TableCell key={i}>
        <div className="h-4 w-full rounded bg-gray-300" />
      </TableCell>
    ))}
  </TableRow>
);
