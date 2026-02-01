import { Suspense, lazy } from 'react'
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
import "@fontsource/material-icons-round";

import { client } from '@lib/api/client.gen';
import { getCsrfFromCookie, fetchCsrfToken } from '@lib/csrf';

import { ThemeModeProvider } from '@lib/theme';
import { AuthProvider, useAuth, ProtectedRoute, AdminRoute } from '@lib/auth';
import { PageLoader } from '@components/layout';

const AuthPage = lazy(() => import('@routes/auth/page'));
const ForgotPage = lazy(() => import('@routes/auth/forgot/page'));
const ResetPage = lazy(() => import('@routes/auth/reset/page'));
const UserProfilePage = lazy(() => import('@routes/u/[id]/page'));
const ArchivePage = lazy(() => import('@routes/archive/page'));
const ArchiveDetailsPage = lazy(() => import('@routes/archive/[id]/page'));
const SettingsPage = lazy(() => import('@routes/settings/page'));
const DiscussionsPage = lazy(() => import('@routes/d/page'));
const ViewDiscussionPost = lazy(() => import('@routes/d/[id]/page'));
const CreateDiscussionPost = lazy(() => import('@routes/d/create/page'));
const EditDiscussionPost = lazy(() => import('@routes/d/[id]/edit/page'));
const EventsPage = lazy(() => import('@routes/events/page'));
const NewsFeedPage = lazy(() => import('@routes/home/page'));
const TeamPage = lazy(() => import('@routes/team/page'));
const ContactPage = lazy(() => import('@routes/contact/page'));
const AdminDashboardPage = lazy(() => import('@routes/admin/page'));

client.setConfig({
  baseUrl: '/api/v1',
  credentials: 'include',
});

// CSRF Interceptor
client.interceptors.request.use(async (request) => {
  const method = request.method.toUpperCase();
  const stateChangingMethods = ['POST', 'PUT', 'DELETE', 'PATCH'];

  if (stateChangingMethods.includes(method)) {
    // Prefer cookie as it is the source of truth for the server
    let tokenValue = getCsrfFromCookie();
    
    if (!tokenValue && !request.url.includes('/auth/csrf')) {
      const data = await fetchCsrfToken();
      tokenValue = data?.csrf || null;
    }

    if (tokenValue) {
      request.headers.set('X-CSRF-Token', tokenValue);
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

const DiscussionRedirect = () => {
  const { id } = useParams();
  return <Navigate to={`/d/${id}`} replace />;
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
                <Route path="/auth/forgot" element={<ForgotPage />} />
                <Route path="/auth/reset" element={<ResetPage />} />
              </Route>
              <Route element={<ProtectedRoute />}>
                <Route path="/archive" element={<ArchivePage />} />
                <Route path="/archive/:moduleId" element={<ArchiveDetailsPage />} />
                <Route path="/archive/:moduleId/:examId" element={<ArchiveDetailsPage />} />
                <Route path="/settings" element={<SettingsPage />} />
                 <Route path="/d/new" element={<CreateDiscussionPost />} />
                 <Route path="/d/:postId/edit" element={<EditDiscussionPost />} />
               </Route>

               <Route element={<AdminRoute />}>
                 <Route path="/admin" element={<AdminDashboardPage />} />
               </Route>

               <Route path="/u/:userId" element={<UserProfilePage />} />

              <Route path="/discussions" element={<DiscussionsPage />} />
              <Route path="/d/:postId" element={<ViewDiscussionPost />} />

              {/* Legacy Redirects */}
              <Route path="/d" element={<Navigate to="/discussions" replace />} />
              <Route path="/discussions/:id" element={<DiscussionRedirect />} />
              <Route path="/forum" element={<Navigate to="/discussions" replace />} />
              <Route path="/forum/:id" element={<DiscussionRedirect />} />
              <Route path="/exams" element={<Navigate to="/archive" replace />} />
              <Route path="/exams/:id" element={<Navigate to="/archive/:id" replace />} />

              <Route path="/events" element={<EventsPage />} />
              <Route path="/events/:eventId" element={<EventsPage />} />
              <Route path="/events/:eventId/:mediaId" element={<EventsPage />} />

              <Route path="/images" element={<Navigate to="/events" replace />} />
              <Route path="/media" element={<Navigate to="/events" replace />} />

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

export default App;
