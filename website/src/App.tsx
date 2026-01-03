import { StrictMode, lazy, Suspense } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Outlet } from 'react-router-dom';

import { client } from '@lib/api/client.gen';

import {
  ThemeModeProvider
} from '@lib/theme';

import { AuthProvider, useAuth } from '@lib/auth';
import ProtectedRoute from '@lib/routes';
import { Box, CircularProgress } from '@mui/material';

const LoginPage = lazy(() => import('@routes/login/page'));
const RegistrationPage = lazy(() => import('@routes/registration/page'));
const DashboardPage = lazy(() => import('@routes/dashboard/page'));
const MediaPage = lazy(() => import('@routes/media/page'));
const TeamPage = lazy(() => import('@routes/team/page'));
const NewsPage = lazy(() => import('@routes/news/page'));
const NewsDetail = lazy(() => import('@routes/news/details'));
const NewsCreatePage = lazy(() => import('@routes/news/create'));
const ExamsPage = lazy(() => import('@routes/exams/page'));
const ExamDetailsPage = lazy(() => import('@routes/exams/extensions'));
const ForumPage = lazy(() => import('@routes/forum/page'));
const NewsFeedPage = lazy(() => import('@routes/homepage/page'));

const LoadingScreen = () => (
  <Box display="flex" justifyContent="center" alignItems="center" minHeight="100vh">
    <CircularProgress />
  </Box>
);

client.setConfig({
  baseUrl: '/api',
  credentials: 'include',
});

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
          <Suspense fallback={<LoadingScreen />}>
            <Routes>

              <Route element={<AuthRedirector />}>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegistrationPage />} />
              </Route>

              <Route element={<ProtectedRoute />}>
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/exams" element={<ExamsPage />} />
                <Route path="/rekos/klausuren/modul" element={<ExamDetailsPage />} />
              </Route>

              <Route path="/forum" element={<ForumPage />} />
              <Route path="/media" element={<MediaPage />} />
              <Route path="/team" element={<TeamPage />} />
              <Route path="/news" element={<NewsPage />} />
              <Route path="/news/:id" element={<NewsDetail />} />
              <Route path="/news/create" element={<NewsCreatePage />} />

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
