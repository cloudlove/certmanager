import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ConfigProvider, theme, Spin } from 'antd';
import Layout from '@/components/Layout';
import AuthGuard from '@/components/AuthGuard';
import LoginPage from '@/pages/login';
import { useThemeStore } from '@/store/theme';

// 路由级懒加载 - 减少首屏加载体积
const DashboardPage = lazy(() => import('@/pages/dashboard'));
const CertificateListPage = lazy(() => import('@/pages/certificates'));
const CertificateDetailPage = lazy(() => import('@/pages/certificates/detail'));
const CSRPage = lazy(() => import('@/pages/csr'));
const CredentialPage = lazy(() => import('@/pages/credentials'));
const DomainPage = lazy(() => import('@/pages/domains'));
const DeploymentListPage = lazy(() => import('@/pages/deployments'));
const DeploymentDetailPage = lazy(() => import('@/pages/deployments/detail'));
const NginxClusterPage = lazy(() => import('@/pages/nginx'));
const NginxClusterDetailPage = lazy(() => import('@/pages/nginx/detail'));
const NotificationPage = lazy(() => import('@/pages/notifications'));
const UserManagePage = lazy(() => import('@/pages/users'));
const RoleManagePage = lazy(() => import('@/pages/roles'));

// 页面加载中的占位组件
const PageLoading = () => (
  <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%', minHeight: 200 }}>
    <Spin size="large" />
  </div>
);

function App() {
  const { theme: themeMode } = useThemeStore();
  const isDark = themeMode === 'dark';

  return (
    <ConfigProvider
      theme={{
        algorithm: isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
      }}
    >
      <BrowserRouter>
        <Suspense fallback={<PageLoading />}>
          <Routes>
            {/* 登录页 - 不需要认证，直接加载 */}
            <Route path="/login" element={<LoginPage />} />
            
            {/* 需要认证的路由 - 页面组件懒加载 */}
            <Route element={<AuthGuard />}>
              <Route path="/" element={<Layout />}>
                <Route index element={<DashboardPage />} />
                <Route path="certificates" element={<CertificateListPage />} />
                <Route path="certificates/:id" element={<CertificateDetailPage />} />
                <Route path="csr" element={<CSRPage />} />
                <Route path="credentials" element={<CredentialPage />} />
                <Route path="domains" element={<DomainPage />} />
                <Route path="deployments" element={<DeploymentListPage />} />
                <Route path="deployments/:id" element={<DeploymentDetailPage />} />
                <Route path="nginx" element={<NginxClusterPage />} />
                <Route path="nginx/:id" element={<NginxClusterDetailPage />} />
                <Route path="notifications" element={<NotificationPage />} />
                <Route path="system/users" element={<UserManagePage />} />
                <Route path="system/roles" element={<RoleManagePage />} />
              </Route>
            </Route>
          </Routes>
        </Suspense>
      </BrowserRouter>
    </ConfigProvider>
  );
}

export default App;
