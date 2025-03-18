import axios from 'axios';
import { getErrorMessage } from '../utils/errorHandler';

/**
 * 测试API连接
 * @returns 测试结果
 */
export const testApiConnection = async (): Promise<{ success: boolean; message: string }> => {
  try {
    const response = await axios.get('/api/health');
    return {
      success: true,
      message: `API连接成功，服务器时间：${response.data.time}`,
    };
  } catch (error) {
    return {
      success: false,
      message: `API连接失败：${getErrorMessage(error)}`,
    };
  }
};

/**
 * 测试音频上传
 * @param audioBlob 音频数据
 * @returns 测试结果
 */
export const testAudioUpload = async (audioBlob: Blob): Promise<{ success: boolean; message: string }> => {
  try {
    const formData = new FormData();
    formData.append('title', '测试音频');
    formData.append('audio', audioBlob, 'test.wav');

    const response = await axios.post('/api/audio/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    return {
      success: true,
      message: `音频上传成功，转录文本：${response.data.meeting.transcript.substring(0, 100)}...`,
    };
  } catch (error) {
    return {
      success: false,
      message: `音频上传失败：${getErrorMessage(error)}`,
    };
  }
};

/**
 * 测试转录分析
 * @returns 测试结果
 */
export const testTranscriptAnalysis = async (): Promise<{ success: boolean; message: string }> => {
  try {
    const testTranscript = `
      张三：大家好，今天我们讨论项目进度。
      李四：我已经完成了前端开发，需要王五测试一下。
      王五：好的，我明天会进行测试。
      张三：那我们决定下周一发布第一个版本。
    `;

    const response = await axios.post('/api/meetings/analyze', {
      title: '测试会议',
      transcript: testTranscript,
    });

    return {
      success: true,
      message: `分析成功，提取了${response.data.meeting.todoItems.length}个待办事项和${response.data.meeting.decisions.length}个决策点`,
    };
  } catch (error) {
    return {
      success: false,
      message: `分析失败：${getErrorMessage(error)}`,
    };
  }
};

/**
 * 运行所有测试
 * @returns 测试结果
 */
export const runAllTests = async (): Promise<Array<{ name: string; success: boolean; message: string }>> => {
  const results = [];

  // 测试API连接
  const connectionResult = await testApiConnection();
  results.push({
    name: 'API连接测试',
    ...connectionResult,
  });

  // 如果API连接成功，继续其他测试
  if (connectionResult.success) {
    // 测试转录分析
    const analysisResult = await testTranscriptAnalysis();
    results.push({
      name: '转录分析测试',
      ...analysisResult,
    });

    // 测试音频上传需要用户提供音频数据，这里不自动测试
  }

  return results;
}; 