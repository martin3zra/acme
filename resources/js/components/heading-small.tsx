export default function HeadingSmall({ title, description, rightPanel }: { title: string; description?: string; rightPanel?: React.ReactNode }) {
  return (
    <header className="flex items-center justify-between">
      <div className="space-y-1">
        <h3 className="mb-0.5 text-base font-medium">{title}</h3>
        {description && <p className="text-muted-foreground text-sm">{description}</p>}
      </div>
      {rightPanel && rightPanel}
    </header>
  );
}
