import { get, post, put, del, getPage } from './request';
import type { CloudCredential, PageResponse } from '@/types';

export interface CredentialQueryParams {
  page?: number;
  pageSize?: number;
  providerType?: string;
}

export interface CreateCredentialData {
  name: string;
  providerType: string;
  accessKey: string;
  secretKey: string;
  extraConfig?: string;
}

export interface UpdateCredentialData {
  name?: string;
  providerType?: string;
  accessKey?: string;
  secretKey?: string;
  extraConfig?: string;
}

export interface TestCredentialResult {
  success: boolean;
  message: string;
}

// 获取凭证列表（分页）
export function getCredentials(params: CredentialQueryParams): Promise<PageResponse<CloudCredential>> {
  return getPage<CloudCredential>('/credentials', params as Record<string, unknown>);
}

// 获取单个凭证详情
export function getCredential(id: string): Promise<CloudCredential> {
  return get<CloudCredential>(`/credentials/${id}`);
}

// 创建凭证
export function createCredential(data: CreateCredentialData): Promise<void> {
  return post<void>('/credentials', data);
}

// 更新凭证
export function updateCredential(id: string, data: UpdateCredentialData): Promise<void> {
  // 过滤掉 undefined 字段
  const filteredData: Partial<UpdateCredentialData> = {};
  for (const key in data) {
    if (data[key as keyof UpdateCredentialData] !== undefined) {
      filteredData[key as keyof UpdateCredentialData] = data[key as keyof UpdateCredentialData];
    }
  }
  return put<void>(`/credentials/${id}`, filteredData);
}

// 删除凭证
export function deleteCredential(id: string): Promise<void> {
  return del<void>(`/credentials/${id}`);
}

// 测试凭证连通性
export function testCredential(id: string): Promise<TestCredentialResult> {
  return post<TestCredentialResult>(`/credentials/${id}/test`);
}
