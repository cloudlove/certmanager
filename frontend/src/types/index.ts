// API 响应类型
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface PageResponse<T> {
  list: T[];
  total: number;
  page: number;
  pageSize: number;
}

// 证书状态
export type CertificateStatus = 'active' | 'expiring' | 'expired' | 'revoked' | 'domain_verify' | 'process' | 'issued' | 'failed';

// 证书类型
export type CertificateType = 'RSA' | 'ECC';

// 证书
export interface Certificate {
  id: string;
  domain: string;
  caProvider: string;
  status: CertificateStatus;
  expireAt: string;
  issuer: string;
  fingerprint: string;
  keyAlgorithm: string;
  serialNumber: string;
  csrId?: number;
  certPem?: string;
  chainPem?: string;
  orderId?: string;       // CA订单ID
  verifyType?: string;    // DNS / FILE
  verifyInfo?: string;    // JSON格式验证信息
  productType?: string;   // DV / OV / EV 证书级别
  domainType?: string;    // single / wildcard / multi 域名类型
  createdAt: string;
  updatedAt: string;
}

// CSR 记录
export interface CSRRecord {
  id: number;
  commonName: string;
  san: string[];
  keyAlgorithm: string;
  keySize: string;
  csrPem?: string;
  status: string;
  createdAt: string;
  updatedAt: string;
  // 组织信息
  countryCode: string;
  province: string;
  locality: string;
  corpName?: string;
  department?: string;
}

// 云提供商类型
export type CloudProvider = 'aliyun' | 'tencent' | 'volcengine' | 'wangsu' | 'aws' | 'azure' | 'gcp' | 'huawei';

// 云凭证
export interface CloudCredential {
  id: string;
  name: string;
  providerType: CloudProvider;
  accessKey: string;
  secretKey: string;
  region: string;
  description: string;
  isValid: boolean;
  lastVerifiedAt: string;
  createdAt: string;
  updatedAt: string;
}

// 域名
export interface Domain {
  id: string;
  name: string;
  rootDomain: string;
  registrar: string;
  dnsProvider: CloudProvider;
  credentialId: string;
  expirationDate: string;
  autoRenew: boolean;
  status: 'active' | 'expiring' | 'expired';
  createdAt: string;
  updatedAt: string;
}

// 部署任务状态
export type DeployTaskStatus = 'pending' | 'deploying' | 'success' | 'partial_success' | 'failed';

// 部署任务项状态
export type DeployTaskItemStatus = 'pending' | 'deploying' | 'success' | 'failed' | 'rolled_back';

// 云提供商类型（用于部署）
export type DeployProviderType = 'aliyun' | 'tencent' | 'volcengine' | 'wangsu' | 'aws' | 'azure';

// 部署目标类型
export type DeployTargetType = 'cdn' | 'slb' | 'dcdn' | 'clb' | 'cloudfront' | 'elb' | 'appgateway';

// 部署目标
export interface DeployTarget {
  providerType: DeployProviderType;
  targetType: DeployTargetType;
  resourceId: string;
  credentialId: string;
}

// 部署任务
export interface DeployTask {
  id: string;
  name: string;
  certificateId: string;
  certificateDomain: string;
  status: DeployTaskStatus;
  totalItems: number;
  successItems: number;
  failedItems: number;
  items?: DeployTaskItem[];
  createdAt: string;
  updatedAt: string;
}

// 部署任务项
export interface DeployTaskItem {
  id: string;
  deployTaskId: string;
  providerType: DeployProviderType;
  targetType: DeployTargetType;
  resourceId: string;
  credentialId: string;
  status: DeployTaskItemStatus;
  errorMessage: string;
  startedAt?: string;
  completedAt?: string;
  createdAt: string;
  updatedAt: string;
}

// 创建部署任务请求
export interface CreateDeployTaskRequest {
  name: string;
  certificateId: string;
  targets: DeployTarget[];
}

// 部署任务查询参数
export interface DeployTaskQueryParams {
  page?: number;
  pageSize?: number;
  status?: DeployTaskStatus;
}

// 部署快照
export interface DeploySnapshot {
  id: string;
  taskId: string;
  certificateId: string;
  certificatePem: string;
  privateKeyPem: string;
  chainPem: string;
  targetConfig: Record<string, unknown>;
  deployedAt: string;
  rolledBackAt: string;
  isRollbackAvailable: boolean;
}

// Nginx 节点状态
export type NginxNodeStatus = 'online' | 'offline' | 'busy' | 'error';

// Nginx 节点
export interface NginxNode {
  id: string;
  clusterId: string;
  name: string;
  host: string;
  port: number;
  status: NginxNodeStatus;
  version: string;
  configPath: string;
  lastHeartbeat: string;
  createdAt: string;
  updatedAt: string;
}

// Nginx 集群
export interface NginxCluster {
  id: string;
  name: string;
  description: string;
  environment: 'production' | 'staging' | 'development';
  nodes: NginxNode[];
  nodeCount: number;
  onlineCount: number;
  createdAt: string;
  updatedAt: string;
}

// 通知渠道类型
export type NotificationChannel = 'email' | 'webhook' | 'sms' | 'dingtalk' | 'wecom' | 'lark';

// 通知规则
export interface NotificationRule {
  id: string;
  name: string;
  description: string;
  events: string[];
  channels: NotificationChannel[];
  recipients: string[];
  webhookUrl: string;
  template: string;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

// 通知日志
export interface NotificationLog {
  id: string;
  ruleId: string;
  ruleName: string;
  event: string;
  channel: NotificationChannel;
  recipient: string;
  subject: string;
  content: string;
  status: 'pending' | 'sent' | 'failed';
  errorMessage: string;
  sentAt: string;
  createdAt: string;
}

// 仪表盘概览
export interface DashboardOverview {
  certificateCount: number;
  expiringCount: number;
  expiredCount: number;
  domainCount: number;
  nginxClusterCount: number;
  nginxNodeCount: number;
  deployTaskCount: number;
  pendingDeployCount: number;
  failedDeployCount: number;
}

// 证书概览 - 大盘展示用
export interface CertOverview {
  total: number;
  issued: number;
  pending: number;
  expired: number;
  revoked: number;
  expiring7: number;
  expiring15: number;
  expiring30: number;
}

// 失败项统计
type FailedItem = {
  domain: string;
  failCount: number;
};

// 部署概览 - 大盘展示用
export interface DeployOverview {
  totalTasks30d: number;
  successRate: number;
  failedTop5: FailedItem[];
}

// 云资源分布项
export interface CloudDistItem {
  provider: string;
  certCount: number;
  deployCount: number;
}

// 到期趋势项
export interface ExpiryTrendItem {
  date: string;
  count: number;
}

// 告警项
export interface AlertItem {
  type: string;
  title: string;
  detail: string;
  level: 'error' | 'warning' | 'info';
}

// Nginx 集群状态
export interface NginxClusterStatus {
  clusterCount: number;
  nodeCount: number;
  onlineCount: number;
  onlineRate: number;
}

// 完整大盘概览
export interface DashboardFullOverview {
  certOverview: CertOverview;
  deployOverview: DeployOverview;
  cloudDistribution: CloudDistItem[];
  nginxStatus: NginxClusterStatus;
  expiryTrend: ExpiryTrendItem[];
  recentTasks: DeployTask[];
  alerts: AlertItem[];
}

// 用户状态
export type UserStatus = 'active' | 'disabled';

// 用户
export interface User {
  id: string;
  username: string;
  nickname?: string;
  email: string;
  role?: string;
  roleId?: string;
  roleName?: string;
  roles?: Role[];
  permissions?: string[];
  status: UserStatus;
  lastLoginAt?: string;
  createdAt: string;
  updatedAt: string;
}

// 登录请求
export interface LoginRequest {
  username: string;
  password: string;
}

// 登录响应
export interface LoginResponse {
  token: string;
  refreshToken: string;
  user: User;
}

// 角色
export interface Role {
  id: string;
  name: string;
  description?: string;
  permissions: Permission[];
  createdAt: string;
  updatedAt: string;
}

// 权限
export interface Permission {
  id: string;
  name: string;
  resource: string;
  action: string;
  description?: string;
}

// 权限分组（用于前端展示）
export interface PermissionGroup {
  resource: string;
  label: string;
  permissions: Permission[];
}

// 创建用户请求
export interface CreateUserRequest {
  username: string;
  password: string;
  nickname?: string;
  email?: string;
  roleId?: string;
  status?: UserStatus;
}

// 更新用户请求
export interface UpdateUserRequest {
  nickname?: string;
  email?: string;
  status?: UserStatus;
}

// 重置密码请求
export interface ResetPasswordRequest {
  newPassword: string;
}

// 分配角色请求
export interface AssignRolesRequest {
  roleId: number;
}

// 创建角色请求
export interface CreateRoleRequest {
  name: string;
  description?: string;
  permissionIds?: string[];
}

// 更新角色请求
export interface UpdateRoleRequest {
  name?: string;
  description?: string;
}

// 分配权限请求
export interface AssignPermissionsRequest {
  permissionIds: string[];
}
