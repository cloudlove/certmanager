import { get, post, del } from './request';
import type { CSRRecord, PageResponse } from '@/types';

export interface GenerateCSRRequest {
  commonName: string;
  sans: string[];
  keyAlgorithm: 'RSA' | 'ECC';
  keySize: number;
  // 阿里云 CreateCsr 参数
  countryCode: string;     // 国家代码，如 CN
  province: string;        // 省份
  locality: string;        // 城市
  corpName?: string;       // 单位名称（可选）
  department?: string;     // 部门（可选）
}

export interface CSRInfo {
  commonName: string;
  sans: string[];
  keyAlgorithm: string;
  keySize: number;
}

export interface CSRListParams {
  page: number;
  pageSize: number;
  search?: string;
}

// 生成 CSR
export function generateCSR(data: GenerateCSRRequest): Promise<CSRRecord> {
  return post<CSRRecord>('/csrs/generate', data);
}

// 获取 CSR 列表
export function getCSRList(params: CSRListParams): Promise<PageResponse<CSRRecord>> {
  return get<PageResponse<CSRRecord>>('/csrs', { params });
}

// 获取单个 CSR
export function getCSR(id: number): Promise<CSRRecord> {
  return get<CSRRecord>(`/csrs/${id}`);
}

// 删除 CSR
export function deleteCSR(id: number): Promise<void> {
  return del<void>(`/csrs/${id}`);
}

// 解析 CSR
export function parseCSR(csrPEM: string): Promise<CSRInfo> {
  return post<CSRInfo>('/csrs/parse', { csrPEM });
}

// 下载 CSR
export function downloadCSR(id: number): Promise<Blob> {
  return get<Blob>(`/csrs/${id}/download-csr`, {
    responseType: 'blob',
  });
}

// 下载私钥
export function downloadPrivateKey(id: number): Promise<Blob> {
  return get<Blob>(`/csrs/${id}/download-key`, {
    responseType: 'blob',
  });
}
