import axios from 'axios';
import { Meeting, MeetingResponse, NotionSyncResponse, TranscriptResponse } from '../models/Meeting';

const API_URL = '/api';

// 创建axios实例
const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 上传音频文件
export const uploadAudio = async (
  audioBlob: Blob,
  title: string,
  syncToNotion: boolean = false
): Promise<MeetingResponse> => {
  const formData = new FormData();
  formData.append('title', title);
  formData.append('syncToNotion', syncToNotion.toString());
  formData.append('audio', audioBlob, 'recording.wav');

  const response = await api.post<MeetingResponse>('/audio/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });

  return response.data;
};

// 发送音频流
export const streamAudio = async (
  audioBlob: Blob,
  sampleRate: number = 16000
): Promise<TranscriptResponse> => {
  const response = await api.post<TranscriptResponse>(
    `/audio/stream?sampleRate=${sampleRate}`,
    audioBlob,
    {
      headers: {
        'Content-Type': 'application/octet-stream',
      },
    }
  );

  return response.data;
};

// 分析转录文本
export const analyzeTranscript = async (
  title: string,
  transcript: string
): Promise<MeetingResponse> => {
  const response = await api.post<MeetingResponse>('/meetings/analyze', {
    title,
    transcript,
  });

  return response.data;
};

// 同步到Notion
export const syncToNotion = async (
  meeting: Meeting,
  markdownReport: string
): Promise<NotionSyncResponse> => {
  const response = await api.post<NotionSyncResponse>('/meetings/sync-notion', {
    meeting,
    markdownReport,
  });

  return response.data;
};

// 健康检查
export const checkHealth = async (): Promise<{ status: string }> => {
  const response = await api.get<{ status: string }>('/health');
  return response.data;
}; 