export class AudioRecorder {
  private mediaRecorder: MediaRecorder | null = null;
  private audioChunks: Blob[] = [];
  private stream: MediaStream | null = null;
  private isRecording = false;

  // 开始录音
  async startRecording(): Promise<void> {
    if (this.isRecording) {
      return;
    }

    try {
      this.stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      this.mediaRecorder = new MediaRecorder(this.stream);
      this.audioChunks = [];

      this.mediaRecorder.addEventListener('dataavailable', (event) => {
        if (event.data.size > 0) {
          this.audioChunks.push(event.data);
        }
      });

      this.mediaRecorder.start();
      this.isRecording = true;
    } catch (error) {
      console.error('获取麦克风权限失败:', error);
      throw new Error('无法访问麦克风');
    }
  }

  // 停止录音
  stopRecording(): Promise<Blob> {
    return new Promise((resolve, reject) => {
      if (!this.mediaRecorder || !this.isRecording) {
        reject(new Error('没有正在进行的录音'));
        return;
      }

      this.mediaRecorder.addEventListener('stop', () => {
        const audioBlob = new Blob(this.audioChunks, { type: 'audio/wav' });
        this.releaseResources();
        resolve(audioBlob);
      });

      this.mediaRecorder.stop();
      this.isRecording = false;
    });
  }

  // 暂停录音
  pauseRecording(): void {
    if (this.mediaRecorder && this.isRecording && this.mediaRecorder.state === 'recording') {
      this.mediaRecorder.pause();
    }
  }

  // 恢复录音
  resumeRecording(): void {
    if (this.mediaRecorder && this.isRecording && this.mediaRecorder.state === 'paused') {
      this.mediaRecorder.resume();
    }
  }

  // 获取当前录音状态
  getRecordingState(): 'inactive' | 'recording' | 'paused' {
    return this.mediaRecorder ? this.mediaRecorder.state : 'inactive';
  }

  // 释放资源
  private releaseResources(): void {
    if (this.stream) {
      this.stream.getTracks().forEach((track) => track.stop());
      this.stream = null;
    }
    this.mediaRecorder = null;
    this.audioChunks = [];
    this.isRecording = false;
  }

  // 获取音频数据
  getAudioBlob(): Blob | null {
    if (this.audioChunks.length === 0) {
      return null;
    }
    return new Blob(this.audioChunks, { type: 'audio/wav' });
  }

  // 检查浏览器是否支持录音
  static isSupported(): boolean {
    return !!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia);
  }
} 