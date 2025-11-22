import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Outlet } from 'react-router-dom';

import { client } from '@lib/api/client.gen';

import {
  ThemeModeProvider,
  ThemeModeToggle
} from '@lib/theme';

import { AuthProvider, useAuth } from '@lib/auth';
import ProtectedRoute from '@lib/routes';

import LoginPage from '@routes/login/page';
import RegistrationPage from '@routes/registration/page';
import DashboardPage from '@routes/dashboard/page';
import MediaPage from '@routes/media/page';
import TeamPage from '@routes/team/page';
import NewsPage from '@routes/news/page';
import ExamsPage from '@routes/exams/page';
import ExamsListPage from '@routes/exams/list/page';
import ForumPage from '@routes/forum/page';

// Configure global client settings
client.setConfig({
  baseUrl: '/api', 
  credentials: 'include', 
});

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
                <Route path="/exams" element={<ExamsPage />} />
                <Route path="/exams/list" element={<ExamsListPage />} />

                <Route path="/" element={<Navigate to="/dashboard" replace />} />
            </Route>

            <Route path="/forum" element={<ForumPage />} />
            <Route path="/media" element={<MediaPage />} />
            <Route path="/team" element={<TeamPage />} />
            <Route path="/news" element={<NewsPage />} />

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
