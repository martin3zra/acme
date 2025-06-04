export default function AppSimpleLayout({ children }: { children?: React.ReactNode }) {
  return (
    <div className="bg-background flex min-h-svh">
      <div className="w-full">{children}</div>
    </div>
  );
}
