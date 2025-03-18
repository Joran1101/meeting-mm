import React, { useState } from 'react';
import {
  Container,
  Typography,
  TextField,
  Button,
  Paper,
  Box,
  CircularProgress,
  Snackbar,
  Alert,
  Tabs,
  Tab,
} from '@mui/material';
import RecordingButton from '../components/RecordingButton';
import AudioPlayer from '../components/AudioPlayer';
import { uploadAudio, analyzeTranscript, syncToNotion as syncToNotionApi } from '../services/api';
import { Meeting } from '../models/Meeting';
import { getErrorMessage } from '../utils/errorHandler';
import ReactMarkdown from 'react-markdown';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

const HomePage: React.FC = () => {
  const [title, setTitle] = useState('');
  const [transcript, setTranscript] = useState('');
  const [meeting, setMeeting] = useState<Meeting | null>(null);
  const [markdownReport, setMarkdownReport] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [tabValue, setTabValue] = useState(0);
  const [syncToNotion, setSyncToNotion] = useState(false);
  const [audioBlob, setAudioBlob] = useState<Blob | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const handleRecordingComplete = async (blob: Blob) => {
    if (!title) {
      setError('请输入会议标题');
      return;
    }

    setAudioBlob(blob);

    try {
      setLoading(true);
      const response = await uploadAudio(blob, title, syncToNotion);
      setMeeting(response.meeting);
      setMarkdownReport(response.markdownReport);
      setTranscript(response.meeting.transcript);
      setTabValue(1); // 切换到转录标签
      setLoading(false);
      setSuccessMessage('音频上传并转录成功！');
    } catch (err) {
      console.error('上传音频失败:', err);
      setError(getErrorMessage(err));
      setLoading(false);
    }
  };

  const handleAnalyzeTranscript = async () => {
    if (!title) {
      setError('请输入会议标题');
      return;
    }

    if (!transcript) {
      setError('请输入会议转录内容');
      return;
    }

    try {
      setLoading(true);
      const response = await analyzeTranscript(title, transcript);
      setMeeting(response.meeting);
      setMarkdownReport(response.markdownReport);
      setTabValue(2); // 切换到分析标签
      setLoading(false);
      setSuccessMessage('转录分析成功！');
    } catch (err) {
      console.error('分析转录失败:', err);
      setError(getErrorMessage(err));
      setLoading(false);
    }
  };

  const handleSyncToNotion = async () => {
    if (!meeting || !markdownReport) {
      setError('没有可同步的会议纪要');
      return;
    }

    try {
      setLoading(true);
      const response = await syncToNotionApi(meeting, markdownReport);
      setMeeting({
        ...meeting,
        notionPageId: response.notionPageId
      });
      setLoading(false);
      setSuccessMessage('已成功同步到Notion！');
    } catch (err) {
      console.error('同步到Notion失败:', err);
      setError(getErrorMessage(err));
      setLoading(false);
    }
  };

  const handleCloseSnackbar = () => {
    setError(null);
    setSuccessMessage(null);
  };

  const handleCopyReport = () => {
    navigator.clipboard.writeText(markdownReport);
    setSuccessMessage('报告已复制到剪贴板');
  };

  return (
    <Container maxWidth="lg">
      <Typography variant="h3" component="h1" gutterBottom align="center" sx={{ mt: 4 }}>
        会议纪要自动生成器
      </Typography>

      <Paper elevation={3} sx={{ p: 3, mb: 4 }}>
        <Box sx={{ mb: 3 }}>
          <TextField
            label="会议标题"
            variant="outlined"
            fullWidth
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            disabled={loading}
            required
          />
        </Box>

        <Tabs value={tabValue} onChange={handleTabChange} aria-label="会议纪要标签">
          <Tab label="录音" />
          <Tab label="转录" />
          <Tab label="分析" />
          <Tab label="报告" />
        </Tabs>

        <TabPanel value={tabValue} index={0}>
          <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', py: 4 }}>
            <Typography variant="h6" gutterBottom>
              录制会议音频
            </Typography>
            <RecordingButton onRecordingComplete={handleRecordingComplete} disabled={loading} />
            
            {audioBlob && (
              <Box sx={{ mt: 4, width: '100%' }}>
                <Typography variant="subtitle1" gutterBottom>
                  预览录音
                </Typography>
                <AudioPlayer audioBlob={audioBlob} />
              </Box>
            )}
            
            <Box sx={{ mt: 2, display: 'flex', alignItems: 'center' }}>
              <Typography variant="body2">同步到Notion:</Typography>
              <Button
                variant={syncToNotion ? "contained" : "outlined"}
                color="primary"
                size="small"
                onClick={() => setSyncToNotion(!syncToNotion)}
                sx={{ ml: 1 }}
              >
                {syncToNotion ? "已启用" : "未启用"}
              </Button>
            </Box>
          </Box>
        </TabPanel>

        <TabPanel value={tabValue} index={1}>
          <Box sx={{ py: 2 }}>
            <Typography variant="h6" gutterBottom>
              会议转录
            </Typography>
            <TextField
              label="转录内容"
              variant="outlined"
              fullWidth
              multiline
              rows={10}
              value={transcript}
              onChange={(e) => setTranscript(e.target.value)}
              disabled={loading}
            />
            <Box sx={{ mt: 2, display: 'flex', justifyContent: 'flex-end' }}>
              <Button
                variant="contained"
                color="primary"
                onClick={handleAnalyzeTranscript}
                disabled={loading || !transcript}
              >
                {loading ? <CircularProgress size={24} /> : '分析转录'}
              </Button>
            </Box>
          </Box>
        </TabPanel>

        <TabPanel value={tabValue} index={2}>
          <Box sx={{ py: 2 }}>
            <Typography variant="h6" gutterBottom>
              分析结果
            </Typography>
            {meeting ? (
              <>
                <Typography variant="subtitle1" gutterBottom>
                  摘要
                </Typography>
                <Paper variant="outlined" sx={{ p: 2, mb: 2 }}>
                  <Typography variant="body1">{meeting.summary}</Typography>
                </Paper>

                <Typography variant="subtitle1" gutterBottom>
                  待办事项
                </Typography>
                <Paper variant="outlined" sx={{ p: 2, mb: 2 }}>
                  {meeting.todoItems.length > 0 ? (
                    meeting.todoItems.map((item, index) => (
                      <Box key={item.id || index} sx={{ mb: 1 }}>
                        <Typography variant="body1">
                          • {item.description} {item.assignee ? `(负责人: ${item.assignee})` : ''}
                        </Typography>
                      </Box>
                    ))
                  ) : (
                    <Typography variant="body2">没有待办事项</Typography>
                  )}
                </Paper>

                <Typography variant="subtitle1" gutterBottom>
                  决策点
                </Typography>
                <Paper variant="outlined" sx={{ p: 2 }}>
                  {meeting.decisions.length > 0 ? (
                    meeting.decisions.map((decision, index) => (
                      <Box key={decision.id || index} sx={{ mb: 1 }}>
                        <Typography variant="body1">
                          • {decision.description}{' '}
                          {decision.madeBy ? `(决策者: ${decision.madeBy})` : ''}
                        </Typography>
                      </Box>
                    ))
                  ) : (
                    <Typography variant="body2">没有决策点</Typography>
                  )}
                </Paper>

                <Box sx={{ mt: 2, display: 'flex', justifyContent: 'flex-end' }}>
                  <Button
                    variant="contained"
                    color="primary"
                    onClick={() => setTabValue(3)}
                  >
                    查看完整报告
                  </Button>
                </Box>
              </>
            ) : (
              <Typography variant="body1">暂无分析结果</Typography>
            )}
          </Box>
        </TabPanel>

        <TabPanel value={tabValue} index={3}>
          <Box sx={{ py: 2 }}>
            <Typography variant="h6" gutterBottom>
              Markdown报告
            </Typography>
            <Paper
              variant="outlined"
              sx={{
                p: 2,
                maxHeight: '500px',
                overflow: 'auto',
                backgroundColor: '#f5f5f5',
              }}
            >
              {markdownReport ? (
                <ReactMarkdown>{markdownReport}</ReactMarkdown>
              ) : (
                <Typography variant="body1">暂无报告</Typography>
              )}
            </Paper>
            <Box sx={{ mt: 2, display: 'flex', justifyContent: 'flex-end' }}>
              <Button
                variant="outlined"
                color="primary"
                onClick={handleCopyReport}
                disabled={!markdownReport}
                sx={{ mr: 1 }}
              >
                复制报告
              </Button>
              {!meeting?.notionPageId && (
                <Button
                  variant="contained"
                  color="primary"
                  onClick={handleSyncToNotion}
                  disabled={loading || !markdownReport}
                >
                  {loading ? <CircularProgress size={24} /> : '同步到Notion'}
                </Button>
              )}
              {meeting?.notionPageId && (
                <Button
                  variant="outlined"
                  color="success"
                  disabled
                >
                  已同步到Notion
                </Button>
              )}
            </Box>
          </Box>
        </TabPanel>
      </Paper>

      <Snackbar open={!!error} autoHideDuration={6000} onClose={handleCloseSnackbar}>
        <Alert onClose={handleCloseSnackbar} severity="error" sx={{ width: '100%' }}>
          {error}
        </Alert>
      </Snackbar>

      <Snackbar open={!!successMessage} autoHideDuration={3000} onClose={handleCloseSnackbar}>
        <Alert onClose={handleCloseSnackbar} severity="success" sx={{ width: '100%' }}>
          {successMessage}
        </Alert>
      </Snackbar>
    </Container>
  );
};

export default HomePage; 