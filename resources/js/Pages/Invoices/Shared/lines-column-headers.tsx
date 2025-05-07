export default function LinesColumnHeaders() {
  return (
    <tr>
      <th scope="col" className="w-60 border border-gray-300 px-1 text-start">
        Reference
      </th>
      <th scope="col" className="w-auto border border-gray-300 px-1 text-start">
        Description
      </th>
      <th scope="col" className="w-36 border border-gray-300 px-1 text-start">
        Unit
      </th>
      <th scope="col" className="w-36 border border-gray-300 px-1 text-end">
        Qty
      </th>
      <th scope="col" className="w-36 border border-gray-300 px-1 text-end">
        Price
      </th>
      <th scope="col" className="w-36 border border-gray-300 px-1 text-end">
        Amount
      </th>
      <th scope="col" className="w-6 gap-2 border border-gray-300 px-5 text-end whitespace-nowrap"></th>
    </tr>
  );
}
