import { Button } from "@/components/ui/button";
import { currencySignature, LineForm } from "@/types";
import { XCircleIcon } from "lucide-react";

type Props = {
  line: LineForm;
  index: number;
  handleRemoveLine: (event: React.MouseEvent<HTMLButtonElement>) => void;
  currency: currencySignature
}

export default function Line({line, index, currency, handleRemoveLine}: Props) {
  return (
    <tr>
      <td className="border px-1 border-gray-300 text-start">{line.name}</td>
      <td className="border px-1 border-gray-300 text-start">{line.description}</td>
      <td className="border px-1 border-gray-300 text-start">{line.unit.name}</td>
      <td className="border px-1 border-gray-300 text-end">{line.quantity}</td>
      <td className="border px-1 border-gray-300 text-end">{currency(line.price || 0)}</td>
      <td className="border px-1 border-gray-300 text-end">{currency(line.amount || 0)}</td>
      <td className="border px-1 border-gray-300 text-end">
        <Button variant={'link'} size={'icon'} className="h-8 w-8 rounded-full p-0" data-index={index} onClick={handleRemoveLine}>
          <XCircleIcon />
        </Button>
      </td>
    </tr>
  )
}