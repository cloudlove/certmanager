import { get, post, put, del, getPage } from './request';
import type { Domain, PageResponse } from '@/types';

export type VerifyStatus = 'normal' | 'mismatch' | 'expired' | 'error' | 'unchecked';

export interface DomainWithVerify extends Domain {
  certificateId?: string;
  certificateName?: string;
  verifyStatus: VerifyStatus;
  lastCheckAt?: string;
}

export interface DomainQueryParams {
  page?: number;
  pageSize?: number;
  search?: string;
}

export interface CreateDomainData {
  name: string;
}

export interface UpdateDomainData {
  certificateId?: number | null;
}

export interface VerifyResult {
  verifyStatus: VerifyStatus;
  message: string;
}

export interface BatchVerifyResult {
  results: Array<{
    id: number;
    status: VerifyStatus;
    message: string;
  }>;
}

// 获取域名列表（分页）
export function getDomains(params: DomainQueryParams): Promise<PageResponse<DomainWithVerify>> {
  return getPage<DomainWithVerify>('/domains', params as Record<string, unknown>);
}

// 获取单个域名详情
export function getDomain(id: number): Promise<DomainWithVerify> {
  return get<DomainWithVerify>(`/domains/${id}`);
}

// 创建域名
export function createDomain(data: CreateDomainData): Promise<DomainWithVerify> {
  return post<DomainWithVerify>('/domains', data);
}

// 更新域名
export function updateDomain(id: number, data: UpdateDomainData): Promise<void> {
  return put<void>(`/domains/${id}`, data);
}

// 删除域名
export function deleteDomain(id: number): Promise<void> {
  return del<void>(`/domains/${id}`);
}

// 校验域名
export function verifyDomain(id: number): Promise<VerifyResult> {
  return post<VerifyResult>(`/domains/${id}/verify`);
}

// 批量校验域名
export function batchVerifyDomains(ids: number[]): Promise<BatchVerifyResult> {
  return post<BatchVerifyResult>('/domains/batch-verify', { ids });
}
