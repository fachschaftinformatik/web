import { ProtectedGuard } from '../guards';

export default function ArchiveLayout({ children }: { children: React.ReactNode }) {
  return <ProtectedGuard>{children}</ProtectedGuard>;
}
