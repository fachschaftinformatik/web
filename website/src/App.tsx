import { StrictMode, Suspense, lazy } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate, useLocation, useParams } from 'react-router-dom';
import { Outlet } from 'react-router-dom';
import { useEffect } from 'react';

// Fonts
import "@fontsource/manrope/400.css";
import "@fontsource/manrope/500.css";
import "@fontsource/manrope/600.css";
import "@fontsource/manrope/700.css";
import "@fontsource/space-grotesk/400.css";
import "@fontsource/space-grotesk/500.css";
import "@fontsource/space-grotesk/600.css";
import "@fontsource/space-grotesk/700.css";
import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";
import "@fontsource/material-icons";
import "@fontsource/material-icons-rounded";

import { client } from '@lib/api/client.gen';
import { getCsrfToken, getCsrfFromCookie, fetchCsrfToken } from '@lib/csrf';

import { ThemeModeProvider } from '@lib/theme';
import { AuthProvider, useAuth, ProtectedRoute } from '@lib/auth';
import { PageLoader } from '@components/layout';

const AuthPage = lazy(() => import('@routes/auth/page'));
const DashboardPage = lazy(() => import('@routes/user/[id]/page'));
const ExamsPage = lazy(() => import('@routes/exams/page'));
const ExamDetailsPage = lazy(() => import('@routes/exams/[id]/page'));
const SettingsPage = lazy(() => import('@routes/settings/page'));
const ForumPage = lazy(() => import('@routes/discussions/page'));
const ViewForumPost = lazy(() => import('@routes/discussions/[id]/page'));
const CreateForumPost = lazy(() => import('@routes/discussions/create/page'));
const EditForumPost = lazy(() => import('@routes/discussions/[id]/edit/page'));
const ImagesPage = lazy(() => import('@routes/images/page'));
const NewsFeedPage = lazy(() => import('@routes/home/page'));
const TeamPage = lazy(() => import('@routes/team/page'));
const ContactPage = lazy(() => import('@routes/contact/page'));

client.setConfig({
  baseUrl: '/api',
  credentials: 'include',
});

// CSRF Interceptor
client.interceptors.request.use(async (request) => {
  const method = request.method.toUpperCase();
  const stateChangingMethods = ['POST', 'PUT', 'DELETE', 'PATCH'];

  if (stateChangingMethods.includes(method)) {
    let token = getCsrfToken() || getCsrfFromCookie();
    
    if (!token && !request.url.includes('/auth/csrf')) {
      token = await fetchCsrfToken();
    }

    if (token) {
      request.headers.set('X-CSRF-Token', token);
    }
  }
  return request;
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
