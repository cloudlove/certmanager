import axios from 'axios';
import type { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError, InternalAxiosRequestConfig } from 'axios';
import { message } from 'antd';
import type { ApiResponse, PageResponse } from '@/types';
import { toCamelCase, toSnakeCase } from '@/utils/transform';
import { refreshToken } from './auth';

// 是否正在刷新 token
let isRefreshing = false;
// 等待刷新 token 的请求队列
let refreshSubscribers: Array<(token: string) => void> = [];

// 订阅 token 刷新
function subscribeTokenRefresh(callback: (token: string) => void) {
  refreshSubscribers.push(callback);
}

// 通知所有订阅者新 token
function onTokenRefreshed(newToken: string) {
  refreshSubscribers.forEach((callback) => callback(newToken));
  refreshSubscribers = [];
}

// 创建 axios 实例
const request: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 从 localStorage 获取 token
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    // 自动转换请求数据 camelCase -> snake_case
    if (config.data && typeof config.data === 'object' && !(config.data instanceof FormData) && !(config.data instanceof Blob)) {
      config.data = toSnakeCase(config.data);
    }
    if (config.params && typeof config.params === 'object') {
      config.params = toSnakeCase(config.params);
    }

    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse<unknown>>) => {
    // 先做 snake_case -> camelCase 转换
    // 注意: ApiResponse 的 code, message, data 本身就是 camelCase, 转换不会影响
    if (response.data && typeof response.data === 'object') {
      response.data = toCamelCase(response.data) as ApiResponse<unknown>;
    }

    const { data } = response;

    // 业务状态码处理
    if (data.code !== 0) {
      message.error(data.message || '请求失败');
      return Promise.reject(new Error(data.message));
    }

    return response;
  },
  async (error: AxiosError<ApiResponse<unknown>>) => {
    const { response, config } = error;
    
    if (response) {
      const { status, data } = response;
      
      // 401 处理 - token 过期，尝试刷新
      if (status === 401) {
        const originalRequest = config as InternalAxiosRequestConfig & { _retry?: boolean };
        
        // 避免重复刷新
        if (originalRequest._retry) {
          message.error('登录已过期，请重新登录');
          localStorage.removeItem('token');
          localStorage.removeItem('refreshToken');
          window.location.href = '/login';
          return Promise.reject(error);
        }
        
        originalRequest._retry = true;
        
        const refreshTokenValue = localStorage.getItem('refreshToken');
        
        if (!refreshTokenValue) {
          message.error('登录已过期，请重新登录');
          localStorage.removeItem('token');
          window.location.href = '/login';
          return Promise.reject(error);
        }
        
        // 如果正在刷新，将请求加入队列等待
        if (isRefreshing) {
          return new Promise((resolve) => {
            subscribeTokenRefresh((newToken: string) => {
              if (originalRequest.headers) {
                originalRequest.headers.Authorization = `Bearer ${newToken}`;
              }
              resolve(request(originalRequest));
            });
          });
        }
        
        isRefreshing = true;
        
        try {
          const res = await refreshToken({ refreshToken: refreshTokenValue });
          const { accessToken, refreshToken: newRefreshToken } = res;
                  
          // 更新存储的 token
          localStorage.setItem('token', accessToken);
          localStorage.setItem('refreshToken', newRefreshToken);
                  
          // 通知所有等待的请求
          onTokenRefreshed(accessToken);
                  
          // 重试原请求
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${accessToken}`;
          }
          return request(originalRequest);
        } catch (refreshError) {
          // 刷新失败，清除状态并跳转登录
          message.error('登录已过期，请重新登录');
          localStorage.removeItem('token');
          localStorage.removeItem('refreshToken');
          window.location.href = '/login';
          return Promise.reject(refreshError);
        } finally {
          isRefreshing = false;
        }
      }
      
      switch (status) {
        case 403:
          message.error('没有权限执行此操作');
          break;
        case 404:
          message.error('请求的资源不存在');
          break;
        case 500:
          message.error('服务器内部错误');
          break;
        default:
          message.error(data?.message || `请求失败 (${status})`);
      }
    } else {
      message.error('网络错误，请检查网络连接');
    }
    
    return Promise.reject(error);
  }
);

// 封装 GET 请求
export function get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return request.get<ApiResponse<T>>(url, config).then((res) => res.data.data);
}

// 封装 POST 请求
export function post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return request.post<ApiResponse<T>>(url, data, config).then((res) => res.data.data);
}

// 封装 PUT 请求
export function put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  return request.put<ApiResponse<T>>(url, data, config).then((res) => res.data.data);
}

// 封装 DELETE 请求
export function del<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return request.delete<ApiResponse<T>>(url, config).then((res) => res.data.data);
}

// 分页请求封装
export function getPage<T>(url: string, params?: Record<string, unknown>): Promise<PageResponse<T>> {
  return get<PageResponse<T>>(url, { params });
}

export default request;
export type { ApiResponse, PageResponse };
