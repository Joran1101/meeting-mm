import React from 'react';
import { Box, Typography, Link } from '@mui/material';

const Footer: React.FC = () => {
  return (
    <Box
      component="footer"
      sx={{
        py: 3,
        px: 2,
        mt: 'auto',
        backgroundColor: (theme) => theme.palette.grey[100],
      }}
    >
      <Typography variant="body2" color="text.secondary" align="center">
        {'© '}
        {new Date().getFullYear()}
        {' 会议纪要自动生成器 | '}
        <Link color="inherit" href="https://github.com/yourusername/meeting-mm" target="_blank">
          GitHub
        </Link>
      </Typography>
    </Box>
  );
};

export default Footer; 