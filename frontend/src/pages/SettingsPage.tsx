import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  TextField,
  Button,
  Paper,
  Box,
  Snackbar,
  Alert,
  Switch,
  FormControlLabel,
  Divider,
} from '@mui/material';

interface Settings {
  deepseekApiKey: string;
  notionApiKey: string;
  notionDatabaseId: string;
  useLocalWhisper: boolean;
  whisperModelPath: string;
}

const SettingsPage: React.FC = () => {
  const [settings, setSettings] = useState<Settings>({
    deepseekApiKey: '',
    notionApiKey: '',
    notionDatabaseId: '',
    useLocalWhisper: true,
    whisperModelPath: '../whisper/models/ggml-base.bin',
  });
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // 从localStorage加载设置
  useEffect(() => {
    const savedSettings = localStorage.getItem('meetingMmSettings');
    if (savedSettings) {
      try {
        const parsedSettings = JSON.parse(savedSettings);
        setSettings(parsedSettings);
      } catch (err) {
        console.error('解析设置失败:', err);
      }
    }
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setSettings({
      ...settings,
      [name]: type === 'checkbox' ? checked : value,
    });
  };

  const handleSave = () => {
    try {
      // 保存到localStorage
      localStorage.setItem('meetingMmSettings', JSON.stringify(settings));
      setSuccess('设置已保存');
    } catch (err) {
      console.error('保存设置失败:', err);
      setError('保存设置失败');
    }
  };

  const handleReset = () => {
    const defaultSettings: Settings = {
      deepseekApiKey: '',
      notionApiKey: '',
      notionDatabaseId: '',
      useLocalWhisper: true,
      whisperModelPath: '../whisper/models/ggml-base.bin',
    };
    setSettings(defaultSettings);
    localStorage.removeItem('meetingMmSettings');
    setSuccess('设置已重置');
  };

  const handleCloseSnackbar = () => {
    setError(null);
    setSuccess(null);
  };

  return (
    <Container maxWidth="md">
      <Typography variant="h4" component="h1" gutterBottom align="center" sx={{ mt: 4 }}>
        设置
      </Typography>

      <Paper elevation={3} sx={{ p: 4, mb: 4 }}>
        <Box sx={{ mb: 4 }}>
          <Typography variant="h6" gutterBottom>
            DeepSeek API设置
          </Typography>
          <TextField
            label="DeepSeek API密钥"
            name="deepseekApiKey"
            value={settings.deepseekApiKey}
            onChange={handleChange}
            fullWidth
            margin="normal"
            type="password"
            helperText="用于分析会议内容，提取待办事项和决策点"
          />
        </Box>

        <Divider sx={{ my: 3 }} />

        <Box sx={{ mb: 4 }}>
          <Typography variant="h6" gutterBottom>
            Notion集成
          </Typography>
          <TextField
            label="Notion API密钥"
            name="notionApiKey"
            value={settings.notionApiKey}
            onChange={handleChange}
            fullWidth
            margin="normal"
            type="password"
            helperText="用于将会议纪要同步到Notion"
          />
          <TextField
            label="Notion数据库ID"
            name="notionDatabaseId"
            value={settings.notionDatabaseId}
            onChange={handleChange}
            fullWidth
            margin="normal"
            helperText="会议纪要将被添加到此数据库中"
          />
        </Box>

        <Divider sx={{ my: 3 }} />

        <Box sx={{ mb: 4 }}>
          <Typography variant="h6" gutterBottom>
            Whisper设置
          </Typography>
          <FormControlLabel
            control={
              <Switch
                checked={settings.useLocalWhisper}
                onChange={handleChange}
                name="useLocalWhisper"
              />
            }
            label="使用本地Whisper模型（推荐，保护隐私）"
          />
          {settings.useLocalWhisper && (
            <TextField
              label="Whisper模型路径"
              name="whisperModelPath"
              value={settings.whisperModelPath}
              onChange={handleChange}
              fullWidth
              margin="normal"
              helperText="本地Whisper模型的路径"
            />
          )}
        </Box>

        <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
          <Button
            variant="outlined"
            color="secondary"
            onClick={handleReset}
            sx={{ mr: 2 }}
          >
            重置
          </Button>
          <Button
            variant="contained"
            color="primary"
            onClick={handleSave}
          >
            保存
          </Button>
        </Box>
      </Paper>

      <Snackbar open={!!error} autoHideDuration={6000} onClose={handleCloseSnackbar}>
        <Alert onClose={handleCloseSnackbar} severity="error" sx={{ width: '100%' }}>
          {error}
        </Alert>
      </Snackbar>

      <Snackbar open={!!success} autoHideDuration={3000} onClose={handleCloseSnackbar}>
        <Alert onClose={handleCloseSnackbar} severity="success" sx={{ width: '100%' }}>
          {success}
        </Alert>
      </Snackbar>
    </Container>
  );
};

export default SettingsPage; 