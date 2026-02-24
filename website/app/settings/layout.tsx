import { ProtectedGuard } from '../guards';

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  return <ProtectedGuard>{children}</ProtectedGuard>;
}
