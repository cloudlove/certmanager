import { get, post, del, getPage } from './request';
import type { PageResponse } from '@/types';

// Nginx 节点
export interface NginxNode {
  id: number;
  clusterId: number;
  ip: string;
  port: string;
  status: 'online' | 'offline' | 'busy' | 'error';
  lastHeartbeat: string;
  createdAt: string;
  updatedAt: string;
}

// Nginx 集群
export interface NginxCluster {
  id: number;
  name: string;
  description: string;
  nodeCount: number;
  onlineCount: number;
  nodes?: NginxNode[];
  createdAt: string;
  updatedAt: string;
}

// 创建集群请求
export interface CreateClusterData {
  name: string;
  description?: string;
}

// 添加节点请求
export interface AddNodeData {
  ip: string;
  port: string;
}

// 部署证书请求
export interface DeployCertificateData {
  certificateId: number;
}

// 部署结果
export interface DeployResult {
  nodeId: number;
  ip: string;
  port: string;
  success: boolean;
  message: string;
}

// 创建集群
export function createCluster(data: CreateClusterData): Promise<NginxCluster> {
  return post<NginxCluster>('/nginx/clusters', data);
}

// 获取集群列表（分页）
export function getClusters(params: { page?: number; pageSize?: number }): Promise<PageResponse<NginxCluster>> {
  return getPage<NginxCluster>('/nginx/clusters', params as Record<string, unknown>);
}

// 获取单个集群详情
export function getCluster(id: number): Promise<NginxCluster> {
  return get<NginxCluster>(`/nginx/clusters/${id}`);
}

// 删除集群
export function deleteCluster(id: number): Promise<void> {
  return del<void>(`/nginx/clusters/${id}`);
}

// 添加节点
export function addNode(clusterId: number, data: AddNodeData): Promise<NginxNode> {
  return post<NginxNode>(`/nginx/clusters/${clusterId}/nodes`, data);
}

// 移除节点
export function removeNode(nodeId: number): Promise<void> {
  return del<void>(`/nginx/nodes/${nodeId}`);
}

// 部署证书到集群
export function deployToCluster(clusterId: number, data: DeployCertificateData): Promise<DeployResult[]> {
  return post<DeployResult[]>(`/nginx/clusters/${clusterId}/deploy`, data);
}
