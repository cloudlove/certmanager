import { get, post, put, del } from './request';
import type { 
  User, 
  Role, 
  Permission,
  CreateUserRequest, 
  UpdateUserRequest, 
  ResetPasswordRequest,
  CreateRoleRequest,
  UpdateRoleRequest,
  AssignPermissionsRequest,
  PageResponse 
} from '@/types';

// ========== 用户管理 API ==========

// 获取用户列表
export function getUsers(params?: { page?: number; pageSize?: number; keyword?: string }): Promise<PageResponse<User>> {
  return get<PageResponse<User>>('/users', { params });
}

// 获取用户详情
export function getUser(id: string): Promise<User> {
  return get<User>(`/users/${id}`);
}

// 创建用户
export function createUser(data: CreateUserRequest): Promise<User> {
  return post<User>('/users', data);
}

// 更新用户
export function updateUser(id: string, data: UpdateUserRequest): Promise<User> {
  return put<User>(`/users/${id}`, data);
}

// 删除用户
export function deleteUser(id: string): Promise<void> {
  return del<void>(`/users/${id}`);
}

// 分配角色（单个角色）
export function assignUserRole(id: string, data: { roleId: number }): Promise<void> {
  return put<void>(`/users/${id}/role`, data);
}

// 重置密码
export function resetUserPassword(id: string, data: ResetPasswordRequest): Promise<void> {
  return put<void>(`/users/${id}/password`, data);
}

// ========== 角色管理 API ==========

// 获取角色列表（全部）
export function getRoles(): Promise<Role[]> {
  return get<Role[]>('/roles/all');
}

// 获取角色详情
export function getRole(id: string): Promise<Role> {
  return get<Role>(`/roles/${id}`);
}

// 创建角色
export function createRole(data: CreateRoleRequest): Promise<Role> {
  return post<Role>('/roles', data);
}

// 更新角色
export function updateRole(id: string, data: UpdateRoleRequest): Promise<Role> {
  return put<Role>(`/roles/${id}`, data);
}

// 删除角色
export function deleteRole(id: string): Promise<void> {
  return del<void>(`/roles/${id}`);
}

// 分配权限
export function assignRolePermissions(id: string, data: AssignPermissionsRequest): Promise<void> {
  return put<void>(`/roles/${id}/permissions`, data);
}

// ========== 权限管理 API ==========

// 获取权限列表
export function getPermissions(): Promise<Permission[]> {
  return get<Permission[]>('/roles/permissions');
}
