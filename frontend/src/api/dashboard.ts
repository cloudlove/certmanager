import { get } from './request';
import type {
  CertOverview,
  DeployOverview,
  CloudDistItem,
  ExpiryTrendItem,
  AlertItem,
  DashboardFullOverview,
} from '@/types';

// 获取综合概览数据
export function getDashboardOverview(): Promise<DashboardFullOverview> {
  return get<DashboardFullOverview>('/dashboard/overview');
}

// 获取证书概览
export function getCertOverview(): Promise<CertOverview> {
  return get<CertOverview>('/dashboard/cert-overview');
}

// 获取部署概览
export function getDeployOverview(): Promise<DeployOverview> {
  return get<DeployOverview>('/dashboard/deploy-overview');
}

// 获取云资源分布
export function getCloudDistribution(): Promise<CloudDistItem[]> {
  return get<CloudDistItem[]>('/dashboard/cloud-distribution');
}

// 获取证书到期趋势
export function getExpiryTrend(days?: number): Promise<ExpiryTrendItem[]> {
  return get<ExpiryTrendItem[]>('/dashboard/expiry-trend', { params: { days: days || 90 } });
}

// 获取告警列表
export function getAlerts(): Promise<AlertItem[]> {
  return get<AlertItem[]>('/dashboard/alerts');
}
