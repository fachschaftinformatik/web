import { ProtectedGuard } from '../../guards';

export default function NewPostLayout({ children }: { children: React.ReactNode }) {
  return <ProtectedGuard>{children}</ProtectedGuard>;
}
