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

import { lazy, Suspense } from 'react';
import { CircularProgress, Box } from '@mui/material';

const LoginPage = lazy(() => import('@routes/login/page'));
const RegistrationPage = lazy(() => import('@routes/registration/page'));
const DashboardPage = lazy(() => import('@routes/user/[id]'));
const MediaPage = lazy(() => import('@routes/media/page'));
const TeamPage = lazy(() => import('@routes/team/page'));
const ExamsPage = lazy(() => import('@routes/exams/page'));
const ExamDetailsPage = lazy(() => import('@routes/exams/[id]'));
const ForumPage = lazy(() => import('@routes/forum/page'));
const CreateForumPost = lazy(() => import('@routes/forum/create'));
const EditForumPost = lazy(() => import('@routes/forum/[id]/edit'));
const ViewForumPost = lazy(() => import('@routes/forum/[id]'));
const NewsFeedPage = lazy(() => import('@routes/homepage/page'));
const ContactPage = lazy(() => import('@routes/contact/page'));
const SettingsPage = lazy(() => import('@routes/settings/page'));

const PageLoader = () => (
  <Box display="flex" justifyContent="center" alignItems="center" minHeight="60vh">
    <CircularProgress />
  </Box>
);

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
          <Suspense fallback={<PageLoader />}>
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
          </Suspense>
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