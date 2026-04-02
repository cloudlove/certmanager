import { get, post, del, getPage } from './request';
import type { DeployTask, CreateDeployTaskRequest, PageResponse, DeployTaskQueryParams } from '@/types';

// 创建部署任务
export function createDeployTask(data: CreateDeployTaskRequest): Promise<DeployTask> {
  return post<DeployTask>('/deploy-tasks', data);
}

// 执行部署任务
export function executeDeployTask(id: string): Promise<void> {
  return post<void>(`/deploy-tasks/${id}/execute`);
}

// 回滚整个部署任务
export function rollbackDeployTask(id: string): Promise<void> {
  return post<void>(`/deploy-tasks/${id}/rollback`);
}

// 回滚单个部署任务项
export function rollbackDeployTaskItem(itemId: string): Promise<void> {
  return post<void>(`/deploy-task-items/${itemId}/rollback`);
}

// 获取部署任务列表（分页）
export function getDeployTasks(params: DeployTaskQueryParams): Promise<PageResponse<DeployTask>> {
  return getPage<DeployTask>('/deploy-tasks', params as Record<string, unknown>);
}

// 获取单个部署任务详情
export function getDeployTask(id: string): Promise<DeployTask> {
  return get<DeployTask>(`/deploy-tasks/${id}`);
}

// 删除部署任务
export function deleteDeployTask(id: string): Promise<void> {
  return del<void>(`/deploy-tasks/${id}`);
}

// 连接部署任务 WebSocket
export function connectDeployWS(taskId: string): WebSocket {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return new WebSocket(`${protocol}//${window.location.host}/ws/deploy/${taskId}`);
}
