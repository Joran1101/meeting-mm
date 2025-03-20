import React, { useState, useRef, useEffect } from 'react';
import { Button, Slider, Box, Typography } from '@mui/material';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import PauseIcon from '@mui/icons-material/Pause';
import StopIcon from '@mui/icons-material/Stop';

interface AudioPlayerProps {
  audioBlob: Blob;
}

const AudioPlayer: React.FC<AudioPlayerProps> = ({ audioBlob }) => {
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const audioUrl = useRef<string | null>(null);

  useEffect(() => {
    // 创建音频URL
    if (audioBlob) {
      setIsLoading(true);
      if (audioUrl.current) {
        URL.revokeObjectURL(audioUrl.current);
      }
      audioUrl.current = URL.createObjectURL(audioBlob);
      
      // 创建音频元素
      const audio = new Audio(audioUrl.current);
      audioRef.current = audio;
      
      // 加载音频元数据
      audio.addEventListener('loadedmetadata', () => {
        // 检查持续时间是否为有效数值
        if (audio.duration && !isNaN(audio.duration) && isFinite(audio.duration)) {
          setDuration(audio.duration);
        } else {
          // 备用方案：使用音频文件大小估算时长
          const audioBitRate = 128000; // 假设比特率为128kbps
          const estimatedDuration = (audioBlob.size * 8) / audioBitRate;
          setDuration(estimatedDuration);
          console.log('无法获取准确时长，使用估算值:', estimatedDuration);
        }
        setIsLoading(false);
      });
      
      // 更新当前时间
      audio.addEventListener('timeupdate', () => {
        setCurrentTime(audio.currentTime);
      });
      
      // 播放结束
      audio.addEventListener('ended', () => {
        setIsPlaying(false);
        setCurrentTime(0);
      });

      // 加载失败时设置默认时长
      audio.addEventListener('error', () => {
        console.error('音频加载失败');
        setDuration(60); // 设置默认时长为60秒
        setIsLoading(false);
      });

      // 确保5秒后无论如何都结束加载状态
      const loadingTimeout = setTimeout(() => {
        if (isLoading) {
          console.log('加载超时，使用默认时长');
          setDuration(60);
          setIsLoading(false);
        }
      }, 5000);

      return () => {
        clearTimeout(loadingTimeout);
      };
    }
    
    // 组件卸载时清理
    return () => {
      if (audioRef.current) {
        audioRef.current.pause();
        audioRef.current = null;
      }
      
      if (audioUrl.current) {
        URL.revokeObjectURL(audioUrl.current);
        audioUrl.current = null;
      }
    };
  }, [audioBlob, isLoading]);

  const handlePlayPause = () => {
    if (!audioRef.current) return;
    
    if (isPlaying) {
      audioRef.current.pause();
    } else {
      audioRef.current.play();
    }
    
    setIsPlaying(!isPlaying);
  };

  const handleStop = () => {
    if (!audioRef.current) return;
    
    audioRef.current.pause();
    audioRef.current.currentTime = 0;
    setCurrentTime(0);
    setIsPlaying(false);
  };

  const handleSliderChange = (_event: Event, newValue: number | number[]) => {
    if (!audioRef.current) return;
    
    const newTime = newValue as number;
    audioRef.current.currentTime = newTime;
    setCurrentTime(newTime);
  };

  const formatTime = (seconds: number): string => {
    if (!seconds || isNaN(seconds) || !isFinite(seconds)) {
      return "00:00";
    }
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`;
  };

  return (
    <Box sx={{ width: '100%', maxWidth: 500, mx: 'auto', mt: 2 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
        <Button
          variant="contained"
          color={isPlaying ? 'secondary' : 'primary'}
          onClick={handlePlayPause}
          startIcon={isPlaying ? <PauseIcon /> : <PlayArrowIcon />}
          sx={{ mr: 1 }}
          disabled={isLoading}
        >
          {isPlaying ? '暂停' : '播放'}
        </Button>
        <Button
          variant="outlined"
          color="primary"
          onClick={handleStop}
          startIcon={<StopIcon />}
          disabled={(!isPlaying && currentTime === 0) || isLoading}
        >
          停止
        </Button>
      </Box>
      
      <Box sx={{ display: 'flex', alignItems: 'center' }}>
        <Typography variant="body2" sx={{ mr: 1, minWidth: 40 }}>
          {formatTime(currentTime)}
        </Typography>
        <Slider
          value={currentTime}
          max={duration || 100}
          onChange={handleSliderChange}
          aria-labelledby="audio-slider"
          sx={{ mx: 2 }}
          disabled={isLoading}
        />
        <Typography variant="body2" sx={{ ml: 1, minWidth: 40 }}>
          {isLoading ? "加载中..." : formatTime(duration)}
        </Typography>
      </Box>
    </Box>
  );
};

export default AudioPlayer; 