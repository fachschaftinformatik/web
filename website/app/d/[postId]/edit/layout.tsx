import { ProtectedGuard } from '../../../guards';

export default function EditPostLayout({ children }: { children: React.ReactNode }) {
  return <ProtectedGuard>{children}</ProtectedGuard>;
}
