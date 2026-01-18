import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate, useLocation, useParams } from 'react-router-dom';
import { Outlet } from 'react-router-dom';
import { useEffect } from 'react';

import { client } from '@lib/api/client.gen';

import { ThemeModeProvider } from '@lib/theme';

import { AuthProvider, useAuth } from '@lib/auth';
import ProtectedRoute from '@lib/routes';

import { lazy, Suspense } from 'react';
import { CircularProgress, Box } from '@mui/material';

const AuthPage = lazy(() => import('@routes/auth/page'));
const DashboardPage = lazy(() => import('@routes/user/[id]/page'));
const ImagesPage = lazy(() => import('@routes/images/page'));
const TeamPage = lazy(() => import('@routes/team/page'));
const ExamsPage = lazy(() => import('@routes/exams/page'));
const ExamDetailsPage = lazy(() => import('@routes/exams/[id]/page'));
const ForumPage = lazy(() => import('@routes/discussions/page'));
const CreateForumPost = lazy(() => import('@routes/discussions/create/page'));
const EditForumPost = lazy(() => import('@routes/discussions/[id]/edit/page'));
const ViewForumPost = lazy(() => import('@routes/discussions/[id]/page'));
const NewsFeedPage = lazy(() => import('@routes/home/page'));
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

const ForumRedirect = () => {
  const { id } = useParams();
  return <Navigate to={`/discussions/${id}`} replace />;
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
                <Route path="/auth" element={<Navigate to="/login" replace />} />
                <Route path="/login" element={<AuthPage />} />
                <Route path="/register" element={<AuthPage />} />
              </Route>
              <Route element={<ProtectedRoute />}>
                <Route path="/exams" element={<ExamsPage />} />
                <Route path="/exams/:id" element={<ExamDetailsPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/discussions/create" element={<CreateForumPost />} />
                <Route path="/discussions/:id/edit" element={<EditForumPost />} />
              </Route>

              <Route path="/user/:userId" element={<DashboardPage />} />

              <Route path="/discussions" element={<ForumPage />} />
              <Route path="/discussions/:id" element={<ViewForumPost />} />

              <Route path="/forum" element={<Navigate to="/discussions" replace />} />
              <Route path="/forum/:id" element={<ForumRedirect />} />

              <Route path="/images" element={<ImagesPage />} />
              <Route path="/images/:eventId" element={<ImagesPage />} />
              <Route path="/images/:eventId/:imageId" element={<ImagesPage />} />

              <Route path="/media" element={<Navigate to="/images" replace />} />

              <Route path="/team" element={<TeamPage />} />
              <Route path="/contact" element={<ContactPage />} />

              <Route path="/" element={<NewsFeedPage />} />
              <Route path="*" element={<Navigate to="/auth" replace />} />
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