import axios from 'axios';
import { ApiError } from '../models/Meeting';

/**
 * 从错误对象中提取错误信息
 * @param error 错误对象
 * @returns 格式化的错误信息
 */
export const getErrorMessage = (error: unknown): string => {
  if (axios.isAxiosError(error)) {
    // 处理Axios错误
    if (error.response) {
      // 服务器返回了错误响应
      const data = error.response.data as ApiError | string;
      if (typeof data === 'object' && data.message) {
        return `${error.response.status}: ${data.message}`;
      } else if (typeof data === 'object' && data.error) {
        return `${error.response.status}: ${data.error}`;
      }
      return `${error.response.status}: ${typeof data === 'string' ? data : JSON.stringify(data)}`;
    } else if (error.request) {
      // 请求已发送但没有收到响应
      return '服务器无响应，请检查网络连接或服务器状态';
    } else {
      // 请求设置时出错
      return `请求错误: ${error.message}`;
    }
  }
  
  // 处理非Axios错误
  if (error instanceof Error) {
    return error.message;
  }
  
  // 未知错误类型
  return String(error);
};

/**
 * 显示错误通知
 * @param error 错误对象或错误消息
 */
export const showErrorNotification = (error: unknown): void => {
  const message = getErrorMessage(error);
  console.error('错误:', message);
  // 这里可以集成通知组件，如Material-UI的Snackbar
  // 例如: enqueueSnackbar(message, { variant: 'error' });
};

/**
 * 记录错误到控制台
 * @param error 错误对象
 * @param context 错误上下文
 */
export const logError = (error: unknown, context?: string): void => {
  const message = getErrorMessage(error);
  if (context) {
    console.error(`[${context}] 错误:`, message, error);
  } else {
    console.error('错误:', message, error);
  }
}; 