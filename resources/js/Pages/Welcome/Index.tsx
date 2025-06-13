import { PageProps } from '@/types';
import { Link } from '@inertiajs/react';

export default function Index({ auth }: PageProps) {
  return (
    <div className="flex h-screen flex-col items-center justify-center">
      <h1 className="text-4xl font-bold">Welcome to the Application!</h1>
      {auth.user ? (
        <>
          <p className="mt-4 text-lg">Hello, {auth.user.name}! You can go to the </p>
          <Link href="/home">Dashboard</Link>
        </>
      ) : (
        <p className="mt-4 text-lg">
          You are not logged in. <Link href="/login">Log in</Link>
        </p>
      )}
    </div>
  );
}
