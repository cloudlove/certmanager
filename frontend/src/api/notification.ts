import { get, post, put, del, getPage } from './request';
import type { PageResponse } from '@/types';

// 通知规则类型
export interface NotificationRule {
  id: string;
  name: string;
  eventType: string;
  thresholdDays: number;
  channels: string[];
  recipients: string[];
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

// 通知日志类型
export interface NotificationLog {
  id: string;
  ruleId: string;
  eventType: string;
  content: string;
  status: 'success' | 'failed';
  sentAt: string;
  createdAt: string;
}

// 创建规则请求
export interface CreateRuleRequest {
  name: string;
  eventType: string;
  thresholdDays: number;
  channels: string[];
  recipients: string[];
  enabled: boolean;
}

// 更新规则请求
export interface UpdateRuleRequest {
  name: string;
  eventType: string;
  thresholdDays: number;
  channels: string[];
  recipients: string[];
  enabled: boolean;
}

// 切换规则状态请求
export interface ToggleRuleRequest {
  enabled: boolean;
}

// 创建通知规则
export function createRule(data: CreateRuleRequest): Promise<NotificationRule> {
  return post<NotificationRule>('/notification-rules', data);
}

// 更新通知规则
export function updateRule(id: string, data: UpdateRuleRequest): Promise<NotificationRule> {
  return put<NotificationRule>(`/notification-rules/${id}`, data);
}

// 删除通知规则
export function deleteRule(id: string): Promise<void> {
  return del<void>(`/notification-rules/${id}`);
}

// 获取通知规则详情
export function getRule(id: string): Promise<NotificationRule> {
  return get<NotificationRule>(`/notification-rules/${id}`);
}

// 获取通知规则列表
export function getRules(params: { page?: number; pageSize?: number }): Promise<PageResponse<NotificationRule>> {
  return getPage<NotificationRule>('/notification-rules', params);
}

// 切换通知规则启用状态
export function toggleRule(id: string, enabled: boolean): Promise<{ enabled: boolean }> {
  return put<{ enabled: boolean }>(`/notification-rules/${id}/toggle`, { enabled });
}

// 测试通知规则
export function testRule(id: string): Promise<{ message: string }> {
  return post<{ message: string }>(`/notification-rules/${id}/test`);
}

// 获取通知日志列表
export function getNotificationLogs(params: { 
  page?: number; 
  pageSize?: number; 
  eventType?: string;
}): Promise<PageResponse<NotificationLog>> {
  return getPage<NotificationLog>('/notification-logs', params);
}
