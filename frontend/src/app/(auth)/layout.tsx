import AuthGuard from "@/components/AuthGuard";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 dark:bg-black px-4">
      <AuthGuard>{children}</AuthGuard>
    </div>
  );
}
