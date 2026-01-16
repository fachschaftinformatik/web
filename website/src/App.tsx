import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { Outlet } from 'react-router-dom';
import { useEffect } from 'react';

import { client } from '@lib/api/client.gen';

import {
  ThemeModeProvider
} from '@lib/theme';

import { AuthProvider, useAuth } from '@lib/auth';
import ProtectedRoute from '@lib/routes';

import LoginPage from '@routes/login/page';
import RegistrationPage from '@routes/registration/page';
import DashboardPage from '@routes/user/[id]';
import MediaPage from '@routes/media/page';
import TeamPage from '@routes/team/page';
import ExamsPage from '@routes/exams/page';
import ExamDetailsPage from '@routes/exams/[id]';
import ForumPage from '@routes/forum/page';
import CreateForumPost from '@routes/forum/create';
import EditForumPost from '@routes/forum/[id]/edit';
import ViewForumPost from '@routes/forum/[id]';
import NewsFeedPage from '@routes/homepage/page';
import ContactPage from '@routes/contact/page';
import SettingsPage from '@routes/settings/page';

client.setConfig({
  baseUrl: '/api',
  credentials: 'include',
});

const ScrollToTop = () => {
  const { pathname } = useLocation();

  useEffect(() => {
    window.scrollTo(0, 0);
  }, [pathname]);

  return null;
};

const AuthRedirector: React.FC = () => {
  const { user, isLoading } = useAuth();
  if (isLoading)
    return null;
  if (user)
    return <Navigate to="/" replace />;

  return <Outlet />;
};

function App() {
  return (
    <ThemeModeProvider>
      <AuthProvider>
        <BrowserRouter>
          <ScrollToTop />
          <Routes>

            <Route element={<AuthRedirector />}>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegistrationPage />} />
            </Route>
            <Route element={<ProtectedRoute />}>
              <Route path="/exams" element={<ExamsPage />} />
              <Route path="/exams/:id" element={<ExamDetailsPage />} />
              <Route path="/settings" element={<SettingsPage />} />
              <Route path="/forum/create" element={<CreateForumPost />} />
              <Route path="/forum/:id/edit" element={<EditForumPost />} />
            </Route>

            <Route path="/user/:userId" element={<DashboardPage />} />

            <Route path="/forum" element={<ForumPage />} />
            <Route path="/forum/:id" element={<ViewForumPost />} />
            <Route path="/media" element={<MediaPage />} />
            <Route path="/team" element={<TeamPage />} />
            <Route path="/contact" element={<ContactPage />} />

            <Route path="/" element={<NewsFeedPage />} />
            <Route path="*" element={<Navigate to="/login" replace />} />

          </Routes>
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