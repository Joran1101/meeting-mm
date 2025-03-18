import React, { useState, useEffect } from 'react';
import { Button, CircularProgress } from '@mui/material';
import MicIcon from '@mui/icons-material/Mic';
import StopIcon from '@mui/icons-material/Stop';
import PauseIcon from '@mui/icons-material/Pause';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import { AudioRecorder } from '../services/audioRecorder';

interface RecordingButtonProps {
  onRecordingComplete: (audioBlob: Blob) => void;
  disabled?: boolean;
}

const RecordingButton: React.FC<RecordingButtonProps> = ({
  onRecordingComplete,
  disabled = false,
}) => {
  const [recorder] = useState<AudioRecorder>(new AudioRecorder());
  const [recordingState, setRecordingState] = useState<'inactive' | 'recording' | 'paused'>('inactive');
  const [isProcessing, setIsProcessing] = useState(false);
  const [recordingTime, setRecordingTime] = useState(0);
  const [timerId, setTimerId] = useState<NodeJS.Timeout | null>(null);

  useEffect(() => {
    return () => {
      if (timerId) {
        clearInterval(timerId);
      }
    };
  }, [timerId]);

  const startRecording = async () => {
    try {
      setIsProcessing(true);
      await recorder.startRecording();
      setRecordingState('recording');
      
      // 开始计时
      const timer = setInterval(() => {
        setRecordingTime((prevTime) => prevTime + 1);
      }, 1000);
      setTimerId(timer);
      
      setIsProcessing(false);
    } catch (error) {
      console.error('开始录音失败:', error);
      setIsProcessing(false);
      alert('无法访问麦克风，请检查权限设置。');
    }
  };

  const stopRecording = async () => {
    try {
      setIsProcessing(true);
      
      // 停止计时
      if (timerId) {
        clearInterval(timerId);
        setTimerId(null);
      }
      
      const audioBlob = await recorder.stopRecording();
      setRecordingState('inactive');
      setRecordingTime(0);
      onRecordingComplete(audioBlob);
      setIsProcessing(false);
    } catch (error) {
      console.error('停止录音失败:', error);
      setIsProcessing(false);
    }
  };

  const pauseRecording = () => {
    recorder.pauseRecording();
    setRecordingState('paused');
    
    // 暂停计时
    if (timerId) {
      clearInterval(timerId);
      setTimerId(null);
    }
  };

  const resumeRecording = () => {
    recorder.resumeRecording();
    setRecordingState('recording');
    
    // 恢复计时
    const timer = setInterval(() => {
      setRecordingTime((prevTime) => prevTime + 1);
    }, 1000);
    setTimerId(timer);
  };

  const formatTime = (seconds: number): string => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`;
  };

  const renderButton = () => {
    if (isProcessing) {
      return (
        <Button
          variant="contained"
          color="primary"
          disabled
          startIcon={<CircularProgress size={24} color="inherit" />}
        >
          处理中...
        </Button>
      );
    }

    switch (recordingState) {
      case 'inactive':
        return (
          <Button
            variant="contained"
            color="primary"
            startIcon={<MicIcon />}
            onClick={startRecording}
            disabled={disabled}
          >
            开始录音
          </Button>
        );
      case 'recording':
        return (
          <>
            <Button
              variant="contained"
              color="secondary"
              startIcon={<StopIcon />}
              onClick={stopRecording}
              style={{ marginRight: 8 }}
            >
              停止
            </Button>
            <Button
              variant="outlined"
              color="primary"
              startIcon={<PauseIcon />}
              onClick={pauseRecording}
            >
              暂停
            </Button>
          </>
        );
      case 'paused':
        return (
          <>
            <Button
              variant="contained"
              color="secondary"
              startIcon={<StopIcon />}
              onClick={stopRecording}
              style={{ marginRight: 8 }}
            >
              停止
            </Button>
            <Button
              variant="outlined"
              color="primary"
              startIcon={<PlayArrowIcon />}
              onClick={resumeRecording}
            >
              继续
            </Button>
          </>
        );
      default:
        return null;
    }
  };

  return (
    <div style={{ display: 'flex', alignItems: 'center', flexDirection: 'column' }}>
      {renderButton()}
      {recordingState !== 'inactive' && (
        <div style={{ marginTop: 16, fontSize: 18 }}>
          录音时长: {formatTime(recordingTime)}
        </div>
      )}
    </div>
  );
};

export default RecordingButton; 