import { get, post, del, getPage } from './request';
import type { Certificate, PageResponse } from '@/types';

export interface CertificateQueryParams {
  page?: number;
  pageSize?: number;
  search?: string;
  status?: string;
  sortBy?: string;
}

export interface CertificateOption {
  id: string;
  name: string;
  domain: string;
}

export interface ApplyCertificateData {
  caProvider: string;
  domain: string;
  csrId: number;
  credentialId: string;
  validateType: string;   // DNS / FILE
  productType: string;    // DV / OV / EV
  domainType: string;     // single / wildcard / multi
}

export interface ImportCertificateData {
  certPEM: string;
  chainPEM?: string;
  privateKeyPEM?: string;
}

// 获取证书列表（分页）
export function getCertificates(params: CertificateQueryParams): Promise<PageResponse<Certificate>> {
  return getPage<Certificate>('/certificates', params as Record<string, unknown>);
}

// 获取单个证书详情
export function getCertificate(id: string): Promise<Certificate> {
  return get<Certificate>(`/certificates/${id}`);
}

// 获取证书选项列表（简化列表，用于下拉选择）
export function getCertificateOptions(): Promise<CertificateOption[]> {
  return get<CertificateOption[]>('/certificates/options');
}

// 删除证书
export function deleteCertificate(id: string): Promise<void> {
  return del<void>(`/certificates/${id}`);
}

// 上传证书
export function uploadCertificate(data: FormData): Promise<Certificate> {
  return post<Certificate>('/certificates/upload', data, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}

// 导入证书
export function importCertificate(data: ImportCertificateData): Promise<Certificate> {
  return post<Certificate>('/certificates/import', data);
}

// 申请证书
export function applyCertificate(data: ApplyCertificateData): Promise<Certificate> {
  return post<Certificate>('/certificates/apply', data);
}

// 同步证书状态
export function syncCertStatus(id: string): Promise<Certificate> {
  return post<Certificate>(`/certificates/${id}/sync`);
}
