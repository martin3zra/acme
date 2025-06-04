import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';

export default function Home({ auth, flash }: PageProps) {
  return (
    <AppLayout user={auth.user}>
      <div>
        <div className="grid auto-rows-min gap-4 md:grid-cols-3">
          <div className="bg-muted/50 aspect-video rounded-xl" />
          <div className="bg-muted/50 aspect-video rounded-xl" />
          <div className="bg-muted/50 aspect-video rounded-xl" />
        </div>
        {flash && <span>{flash.success}</span>}
        <h1>Home Page</h1>
      </div>
    </AppLayout>
  );
}
