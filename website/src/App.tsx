import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import RegistrationPage from './Registration/RegistrationPage';
import LoginPage from './Login/LoginPage';
import ProtectedRoute from './ProtectedRoute';
import DashboardPage from './Dashboard/DashboardPage';
import ForumPostsPage from './Forum/ForumPostsPage';
import { AuthProvider, useAuth } from './AuthContext';
import { Outlet } from 'react-router-dom';

import { ThemeModeProvider } from './ThemeModeContext';
import ThemeModeToggle from './ThemeModeToggle';

const AuthRedirector: React.FC = () => {
    const { user, isLoading } = useAuth();
    if (isLoading)
        return null;
    if (user)
        return <Navigate to="/dashboard" replace />;

    return <Outlet />;
};

function App() {
  return (
    <ThemeModeProvider>
      <AuthProvider>
        <BrowserRouter>
          <Routes>

            <Route element={<AuthRedirector />}>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegistrationPage />} />
            </Route>

            <Route element={<ProtectedRoute />}>
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/forum" element={<ForumPostsPage />} />
                <Route path="/" element={<Navigate to="/dashboard" replace />} />
            </Route>

            <Route path="*" element={<Navigate to="/login" replace />} />

          </Routes>
          <ThemeModeToggle />
        </BrowserRouter>
      </AuthProvider>
    </ThemeModeProvider>
  );
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
