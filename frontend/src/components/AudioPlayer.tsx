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
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const audioUrl = useRef<string | null>(null);

  useEffect(() => {
    // 创建音频URL
    if (audioBlob) {
      if (audioUrl.current) {
        URL.revokeObjectURL(audioUrl.current);
      }
      audioUrl.current = URL.createObjectURL(audioBlob);
      
      // 创建音频元素
      const audio = new Audio(audioUrl.current);
      audioRef.current = audio;
      
      // 加载音频元数据
      audio.addEventListener('loadedmetadata', () => {
        setDuration(audio.duration);
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
  }, [audioBlob]);

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
        >
          {isPlaying ? '暂停' : '播放'}
        </Button>
        <Button
          variant="outlined"
          color="primary"
          onClick={handleStop}
          startIcon={<StopIcon />}
          disabled={!isPlaying && currentTime === 0}
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
        />
        <Typography variant="body2" sx={{ ml: 1, minWidth: 40 }}>
          {formatTime(duration)}
        </Typography>
      </Box>
    </Box>
  );
};

export default AudioPlayer; 