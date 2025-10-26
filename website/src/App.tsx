import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import RegistrationPage from './Registration/RegistrationPage';
import LoginPage from './Login/LoginPage';

// NEU: Importiere das Theme und den ThemeProvider
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import theme from './theme'; // Importiere dein zentrales Theme

function App() {
  return (
    // Der ThemeProvider umschließt jetzt die gesamte App
    <ThemeProvider theme={theme}>
      <CssBaseline /> {/* Sorgt für konsistentes CSS */}
      <BrowserRouter>
        <Routes>
          <Route path="/register" element={<RegistrationPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="*" element={<Navigate to="/login" />} />
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;