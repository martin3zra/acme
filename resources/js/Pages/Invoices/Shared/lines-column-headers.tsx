export default function LinesColumnHeaders () {
  return (
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
  )
}