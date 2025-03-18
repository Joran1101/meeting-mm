import React, { useState } from 'react';
import { Box, CssBaseline } from '@mui/material';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import HomePage from './pages/HomePage';
import SettingsPage from './pages/SettingsPage';
import Navigation from './components/Navigation';
import Footer from './components/Footer';

// 创建主题
const theme = createTheme({
  palette: {
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});

const App: React.FC = () => {
  const [currentPage, setCurrentPage] = useState<'home' | 'settings'>('home');

  const handleNavigate = (page: 'home' | 'settings') => {
    setCurrentPage(page);
  };

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
        <Navigation currentPage={currentPage} onNavigate={handleNavigate} />
        <Box component="main" sx={{ flexGrow: 1, py: 2 }}>
          {currentPage === 'home' && <HomePage />}
          {currentPage === 'settings' && <SettingsPage />}
        </Box>
        <Footer />
      </Box>
    </ThemeProvider>
  );
};

export default App; 