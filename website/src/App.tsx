import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Outlet } from 'react-router-dom';

import { client } from '@lib/api/client.gen';

import {
  ThemeModeProvider
} from '@lib/theme';

import { AuthProvider, useAuth } from '@lib/auth';
import ProtectedRoute from '@lib/routes';

import LoginPage from '@routes/login/page';
import RegistrationPage from '@routes/registration/page';
import DashboardPage from '@routes/dashboard/page';
import MediaPage from '@routes/media/page';
import TeamPage from '@routes/team/page';
import NewsPage from '@routes/news/page';
import NewsDetail from '@routes/news/details';
import CreateNewsPage from '@routes/news/create'; 
import ExamsPage from '@routes/exams/page';
import ExamDetailsPage from '@routes/exams/extensions';
import ForumPage from '@routes/forum/page';
import NewsFeedPage from '@routes/homepage/page';

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
          <Routes>

            <Route element={<AuthRedirector />}>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegistrationPage />} />
            </Route>
            <Route element={<ProtectedRoute />}>
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/exams" element={<ExamsPage />} />
                <Route path="/rekos/klausuren/modul" element={<ExamDetailsPage />} />
                <Route path="/news/create" element={<CreateNewsPage />} />
            </Route>

            <Route path="/forum" element={<ForumPage />} />
            <Route path="/media" element={<MediaPage />} />
            <Route path="/team" element={<TeamPage />} />
            <Route path="/news" element={<NewsPage />} />
            <Route path="/news/:id" element={<NewsDetail />} />

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