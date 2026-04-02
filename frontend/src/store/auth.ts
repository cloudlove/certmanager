import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { login as loginApi, logout as logoutApi, getCurrentUser } from '@/api/auth';
import type { User } from '@/types';

export interface AuthState {
  // 状态
  token: string | null;
  refreshToken: string | null;
  user: User | null;
  permissions: string[];
  isLoggedIn: boolean;
  isLoading: boolean;

  // Actions
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  fetchCurrentUser: () => Promise<void>;
  setToken: (token: string, refreshToken: string) => void;
  clearAuth: () => void;
  
  // 权限检查
  hasPermission: (resource: string, action?: string) => boolean;
  hasAnyPermission: (permissions: string[]) => boolean;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // 初始状态
      token: null,
      refreshToken: null,
      user: null,
      permissions: [],
      isLoggedIn: false,
      isLoading: false,

      // 登录
      login: async (username: string, password: string) => {
        set({ isLoading: true });
        try {
          const response = await loginApi({ username, password });
          // 从 user 对象中提取 permissions
          const permissions = response.user?.permissions || [];
          set({
            token: response.accessToken,
            refreshToken: response.refreshToken,
            user: response.user,
            permissions,
            isLoggedIn: true,
            isLoading: false,
          });
          // 存储 token 到 localStorage 供 request.ts 使用
          localStorage.setItem('token', response.accessToken);
          localStorage.setItem('refreshToken', response.refreshToken);
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },

      // 登出
      logout: async () => {
        try {
          await logoutApi();
        } finally {
          get().clearAuth();
        }
      },

      // 获取当前用户信息
      fetchCurrentUser: async () => {
        try {
          const user = await getCurrentUser();
          // 从 user 对象中提取 permissions
          const permissions = user?.permissions || [];
          set({ user, permissions });
        } catch (error) {
          // 获取失败，清除登录状态
          get().clearAuth();
          throw error;
        }
      },

      // 设置 token
      setToken: (token: string, refreshToken: string) => {
        set({ token, refreshToken });
        localStorage.setItem('token', token);
        localStorage.setItem('refreshToken', refreshToken);
      },

      // 清除认证状态
      clearAuth: () => {
        set({
          token: null,
          refreshToken: null,
          user: null,
          permissions: [],
          isLoggedIn: false,
        });
        localStorage.removeItem('token');
        localStorage.removeItem('refreshToken');
      },

      // 检查是否有特定权限
      // resource: 资源名，如 'certificate', 'user'
      // action: 操作，如 'read', 'write', 'delete'
      hasPermission: (resource: string, action?: string): boolean => {
        const { permissions, user } = get();
        
        // 管理员拥有所有权限
        if (user?.role === 'admin') {
          return true;
        }
        
        if (!permissions || permissions.length === 0) {
          return false;
        }
        
        // 如果指定了 action，检查 resource:action 格式
        if (action) {
          const permissionKey = `${resource}:${action}`;
          return permissions.includes(permissionKey);
        }
        
        // 只指定了 resource，检查是否有该资源的任何权限
        return permissions.some(p => p.startsWith(`${resource}:`));
      },

      // 检查是否有任意一个权限
      hasAnyPermission: (perms: string[]): boolean => {
        const { permissions, user } = get();
        
        // 管理员拥有所有权限
        if (user?.role === 'admin') {
          return true;
        }
        
        if (!permissions || permissions.length === 0) {
          return false;
        }
        
        return perms.some(p => permissions.includes(p));
      },
    }),
    {
      name: 'certmanager-auth',
      partialize: (state) => ({
        token: state.token,
        refreshToken: state.refreshToken,
        user: state.user,
        permissions: state.permissions,
        isLoggedIn: state.isLoggedIn,
      }),
    }
  )
);

export default useAuthStore;
