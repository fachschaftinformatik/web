import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import RegistrationPage from './Registration/RegistrationPage';
import LoginPage from './Login/LoginPage';
import ProtectedRoute from './ProtectedRoute';
import { AuthProvider, useAuth } from './AuthContext';
import { Outlet } from 'react-router-dom';

import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import theme from './theme';

import StartseiteRekos from './Reko/StartseiteRekos';
import Klausurrekos from './Reko/Klausurrekos';
import KlausurrekosErweiterung from './Reko/KlausurrekosErweiterung';

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
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <AuthProvider>
        <BrowserRouter>
          <Routes>

            <Route element={<AuthRedirector />}>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegistrationPage />} />
            </Route>

            <Route element={<ProtectedRoute />}>
                <Route path="/rekos" element={<Klausurrekos />} />
<Route path="/rekos/erweitert" element={<KlausurrekosErweiterung />} />
<Route path="/dashboard" element={<StartseiteRekos />} />
            </Route>

          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </ThemeProvider>
  );
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
