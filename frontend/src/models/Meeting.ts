export interface Meeting {
  id: string;
  title: string;
  date: string;
  participants: string[];
  transcript: string;
  summary: string;
  todoItems: TodoItem[];
  decisions: Decision[];
  createdAt: string;
  updatedAt: string;
  notionPageId?: string;
}

export interface TodoItem {
  id: string;
  description: string;
  assignee: string;
  dueDate?: string;
  status: 'pending' | 'in_progress' | 'completed';
}

export interface Decision {
  id: string;
  description: string;
  madeBy?: string;
}

export interface TranscriptSegment {
  id: string;
  meetingId: string;
  startTime: number;
  endTime: number;
  speaker?: string;
  text: string;
  timestamp: string;
}

export interface MeetingResponse {
  meeting: Meeting;
  markdownReport: string;
}

export interface TranscriptResponse {
  transcript: string;
}

export interface NotionSyncResponse {
  notionPageId: string;
}

export interface ApiError {
  error: string;
  message?: string;
} 