import { useEffect, useState } from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { Spin } from 'antd';
import { useAuthStore } from '@/store/auth';

export default function AuthGuard() {
  const location = useLocation();
  const { isLoggedIn, token, fetchCurrentUser, user } = useAuthStore();
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const initAuth = async () => {
      // 如果有 token 但没有用户信息，尝试获取用户信息
      if (token && !user) {
        try {
          await fetchCurrentUser();
        } catch {
          // 获取失败会在 fetchCurrentUser 中清除状态
        }
      }
      setIsLoading(false);
    };

    initAuth();
  }, [token, user, fetchCurrentUser]);

  // 加载中显示 loading
  if (isLoading) {
    return (
      <div
        style={{
          height: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  // 未登录，跳转到登录页
  if (!isLoggedIn || !token) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // 已登录，渲染子路由
  return <Outlet />;
}
