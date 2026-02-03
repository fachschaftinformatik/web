import { Suspense, lazy, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation, useParams, Outlet } from 'react-router-dom';

import "@fontsource/manrope/latin-400.css";
import "@fontsource/manrope/latin-500.css";
import "@fontsource/manrope/latin-600.css";
import "@fontsource/manrope/latin-700.css";
import "@fontsource/space-grotesk/latin-400.css";
import "@fontsource/space-grotesk/latin-700.css";

import { client } from '@lib/api/client.gen';
import { getCsrfFromCookie, fetchCsrfToken } from '@lib/csrf';
import { ThemeModeProvider } from '@lib/theme';
import { AuthProvider, useAuth, ProtectedRoute, AdminRoute } from '@lib/auth';
import Loader from '@components/Loader';
import Layout from '@components/Layout';

const Auth = lazy(() => import('@pages/Auth'));
const Profile = lazy(() => import('@pages/Profile'));
const Archive = lazy(() => import('@pages/Archive'));
const ArchiveDetail = lazy(() => import('@pages/ArchiveDetail'));
const Settings = lazy(() => import('@pages/Settings'));
const Posts = lazy(() => import('@pages/Posts'));
const Post = lazy(() => import('@pages/Post'));
const PostCreate = lazy(() => import('@pages/PostCreate'));
const PostEdit = lazy(() => import('@pages/PostEdit'));
const Events = lazy(() => import('@pages/Events'));
const Home = lazy(() => import('@pages/Home'));
const Members = lazy(() => import('@pages/Members'));
const Contact = lazy(() => import('@pages/Contact'));
const Admin = lazy(() => import('@pages/Admin'));

client.setConfig({
  baseUrl: '/api/v1',
  credentials: 'include',
});

client.interceptors.request.use(async (request) => {
  const method = request.method.toUpperCase();
  const stateChangingMethods = ['POST', 'PUT', 'DELETE', 'PATCH'];

  if (stateChangingMethods.includes(method)) {
    let tokenValue = getCsrfFromCookie();
    if (!tokenValue && !request.url.includes('/auth/csrf')) {
      const data = await fetchCsrfToken();
      tokenValue = data?.csrf || null;
    }
    if (tokenValue) request.headers.set('X-CSRF-Token', tokenValue);
  }
  return request;
});

const ScrollToTop = () => {
  const { pathname } = useLocation();
  useEffect(() => { window.scrollTo(0, 0); }, [pathname]);
  return null;
};

const PostRedirect = () => {
  const { id } = useParams();
  return <Navigate to={`/post/${id}`} replace />;
};

const AuthRedirector: React.FC = () => {
  const { user, isLoading } = useAuth();
  if (isLoading) return null;
  if (user) return <Navigate to="/" replace />;
  return <Outlet />;
};

function App() {
  return (
    <ThemeModeProvider>
      <AuthProvider>
        <BrowserRouter>
          <ScrollToTop />
          <Suspense fallback={<Loader />}>
            <Routes>
              <Route element={<AuthRedirector />}>
                <Route path="/auth" element={<Navigate to="/login" replace />} />
                <Route path="/login" element={<Auth />} />
                <Route path="/register" element={<Auth />} />
              </Route>
              <Route element={<Layout />}>
                <Route element={<ProtectedRoute />}>
                  <Route path="/archive" element={<Archive />} />
                  <Route path="/archive/:moduleId" element={<ArchiveDetail />} />
                  <Route path="/archive/:moduleId/:examId" element={<ArchiveDetail />} />
                  <Route path="/settings" element={<Settings />} />
                  <Route path="/post/new" element={<PostCreate />} />
                  <Route path="/post/:postId/edit" element={<PostEdit />} />
                </Route>
                <Route element={<AdminRoute />}>
                  <Route path="/admin" element={<Admin />} />
                </Route>
                <Route path="/u/:userId" element={<Profile />} />
                <Route path="/posts" element={<Posts />} />
                <Route path="/post/:postId" element={<Post />} />
                <Route path="/events" element={<Events />} />
                <Route path="/events/:eventId" element={<Events />} />
                <Route path="/events/:eventId/:mediaId" element={<Events />} />
                <Route path="/members" element={<Members />} />
                <Route path="/contact" element={<Contact />} />
                <Route path="/" element={<Home />} />
                
                <Route path="/discussions" element={<Navigate to="/posts" replace />} />
                <Route path="/discussions/:id" element={<PostRedirect />} />
                <Route path="/d" element={<Navigate to="/posts" replace />} />
                <Route path="/d/:id" element={<PostRedirect />} />
                <Route path="/forum/:id" element={<PostRedirect />} />
                <Route path="/exams/:id" element={<Navigate to="/archive/:id" replace />} />
              </Route>
              <Route path="*" element={<Navigate to="/auth" replace />} />
            </Routes>
          </Suspense>
        </BrowserRouter>
      </AuthProvider>
    </ThemeModeProvider>
  );
}

export default App;
