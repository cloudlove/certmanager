import { post, get, put } from './request';
import type { User } from '@/types';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresAt: string;
  user: User;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface RefreshTokenResponse {
  accessToken: string;
  refreshToken: string;
  expiresAt: string;
}

export interface ChangePasswordRequest {
  oldPassword: string;
  newPassword: string;
}

// 登录
export function login(data: LoginRequest): Promise<LoginResponse> {
  return post<LoginResponse>('/auth/login', data);
}

// 登出
export function logout(): Promise<void> {
  return post<void>('/auth/logout');
}

// 刷新 token
export function refreshToken(data: RefreshTokenRequest): Promise<RefreshTokenResponse> {
  return post<RefreshTokenResponse>('/auth/refresh', data);
}

// 获取当前用户信息
export function getCurrentUser(): Promise<User> {
  return get<User>('/auth/me');
}

// 修改密码
export function changePassword(data: ChangePasswordRequest): Promise<void> {
  return put<void>('/auth/password', data);
}
