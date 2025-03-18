import React from 'react';
import { AppBar, Toolbar, Typography, Button, Box } from '@mui/material';
import MicIcon from '@mui/icons-material/Mic';
import SettingsIcon from '@mui/icons-material/Settings';

interface NavigationProps {
  currentPage: 'home' | 'settings';
  onNavigate: (page: 'home' | 'settings') => void;
}

const Navigation: React.FC<NavigationProps> = ({ currentPage, onNavigate }) => {
  return (
    <AppBar position="static">
      <Toolbar>
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          会议纪要自动生成器
        </Typography>
        <Box>
          <Button
            color="inherit"
            startIcon={<MicIcon />}
            onClick={() => onNavigate('home')}
            sx={{
              fontWeight: currentPage === 'home' ? 'bold' : 'normal',
              borderBottom: currentPage === 'home' ? '2px solid white' : 'none',
            }}
          >
            录音
          </Button>
          <Button
            color="inherit"
            startIcon={<SettingsIcon />}
            onClick={() => onNavigate('settings')}
            sx={{
              fontWeight: currentPage === 'settings' ? 'bold' : 'normal',
              borderBottom: currentPage === 'settings' ? '2px solid white' : 'none',
            }}
          >
            设置
          </Button>
        </Box>
      </Toolbar>
    </AppBar>
  );
};

export default Navigation; 