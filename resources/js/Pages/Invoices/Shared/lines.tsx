import { AlertDestructive } from "@/components/alert-destructive"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useNumber } from "@/composables/use-number"
import { Item, LineForm } from "@/types"
import { XCircleIcon } from "lucide-react"

type LinesProps = {
  lines: LineForm[]
  lineError?: string
  currentItem: Item | undefined
  handleRemoveLine: (event: React.MouseEvent<HTMLButtonElement>) => void
  handleKeyDown: (event: React.KeyboardEvent<HTMLInputElement>) => void
  amount: number
  setAmount: React.Dispatch<React.SetStateAction<number>>
  referenceInputRef: React.RefObject<HTMLInputElement | null>
  qtyInputRef: React.RefObject<HTMLInputElement | null>
}
export const Lines = ({
  lineError,
  referenceInputRef,
  qtyInputRef,
  lines,
  currentItem,
  handleRemoveLine,
  handleKeyDown,
  amount,
  setAmount
}: LinesProps) => {
  const currency = useNumber().currency;
  const computedItemAmount = (qty: number) => {
    setAmount(qty * (currentItem?.price || 0));
  };

  return (
    <table className="w-full table-auto">
      <thead>
        <tr>
          <th scope="col" className="w-60 pe-1 border border-gray-300">
            <Input
              name="reference"
              ref={referenceInputRef}
              data-reset={false}
              placeholder="Item reference"
              onKeyDown={handleKeyDown}
              className="border-none focus-visible:border-none focus-visible:ring-[2px] rounded-none"
              tabIndex={0}
            />
          </th>
          <th scope="col" className="w-auto px-1 border bg-gray-50 border-gray-300">
            <Label>{currentItem?.description}</Label>
          </th>
          <th scope="col" className="w-36 px-1 border bg-gray-50 border-gray-300">
            <Label>{currentItem?.unit.name}</Label>
          </th>
          <th scope="col" className="w-36 border border-gray-300">
            <Input
              type="number"
              min={1}
              name="quantity"
              className="text-end border-none focus-visible:border-none focus-visible:ring-[2px] rounded-none"
              tabIndex={1}
              ref={qtyInputRef}
              onFocus={(e) => computedItemAmount(e.currentTarget.valueAsNumber)}
              onChange={(e) => computedItemAmount(e.target.valueAsNumber)}
              onKeyDown={handleKeyDown}
            />
          </th>
          <th scope="col" className="w-36 px-1 border border-gray-300 bg-gray-50 text-end">
            <Label className="block">{currency(currentItem?.price || 0)}</Label>
          </th>
          <th scope="col" className="w-36 px-1 border border-gray-300 text-end">
            {amount > 0 ? currency(amount) : ''}
          </th>
          <th scope="col" className="w-6 border border-gray-300 text-end" />
        </tr>
        <tr>
          <th scope="col" className="w-60 px-1 border border-gray-300 text-start">
            Reference
          </th>
          <th scope="col" className="w-auto px-1 border border-gray-300 text-start">
            Description
          </th>
          <th scope="col" className="w-36 px-1 border border-gray-300 text-start">
            Unit
          </th>
          <th scope="col" className="w-36 px-1 border border-gray-300 text-end">
            Quantity
          </th>
          <th scope="col" className="w-36 px-1 border border-gray-300 text-end">
            Price
          </th>
          <th scope="col" className="w-36 px-1 border border-gray-300 text-end">
            Amount
          </th>
          <th scope="col" className="w-6 gap-2 border border-gray-300 px-5 text-end whitespace-nowrap"></th>
        </tr>
      </thead>
      <tbody>
        {lines && lines.map((line, index) => (
          <tr key={index}>
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
        ))}
      </tbody>
      {lineError &&
        <tfoot>
          <tr>
            <td colSpan={7}>
            <div className="py-3">
              <AlertDestructive description={lineError} destroyable={false} />
            </div>
            </td>
          </tr>
        </tfoot>
      }
    </table>
  )
}