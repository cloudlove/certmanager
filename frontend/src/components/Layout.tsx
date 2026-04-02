import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { ProLayout } from '@ant-design/pro-layout';
import {
  DashboardOutlined,
  SafetyCertificateOutlined,
  KeyOutlined,
  GlobalOutlined,
  CloudOutlined,
  DeploymentUnitOutlined,
  ClusterOutlined,
  BellOutlined,
  BulbOutlined,
  BulbFilled,
  SettingOutlined,
  TeamOutlined,
  SafetyOutlined,
  UserOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import { useThemeStore } from '@/store/theme';
import { useAuthStore } from '@/store/auth';
import type { MenuDataItem } from '@ant-design/pro-layout';
import { Switch, Space, Tooltip, Avatar, Dropdown, Button } from 'antd';
import type { MenuProps } from 'antd';

const logo = (
  <svg viewBox="0 0 1024 1024" width="32" height="32">
    <path
      fill="currentColor"
      d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64zm0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"
    />
    <path
      fill="currentColor"
      d="M686.7 638.6L544.1 535.5V288c0-4.4-3.6-8-8-8H488c-4.4 0-8 3.6-8 8v275.4c0 2.6 1.2 5 3.3 6.5l165.4 120.6c3.6 2.6 8.6 1.8 11.2-1.7l28.6-39.2c2.6-3.7 1.8-8.7-1.8-11.2z"
    />
  </svg>
);

const menuData: MenuDataItem[] = [
  {
    path: '/',
    name: '大盘总览',
    icon: <DashboardOutlined />,
  },
  {
    path: '/certificates',
    name: '证书管理',
    icon: <SafetyCertificateOutlined />,
  },
  {
    path: '/csr',
    name: 'CSR 管理',
    icon: <KeyOutlined />,
  },
  {
    path: '/domains',
    name: '域名管理',
    icon: <GlobalOutlined />,
  },
  {
    path: '/credentials',
    name: '云凭证管理',
    icon: <CloudOutlined />,
  },
  {
    path: '/deployments',
    name: '部署任务',
    icon: <DeploymentUnitOutlined />,
  },
  {
    path: '/nginx',
    name: 'Nginx 集群',
    icon: <ClusterOutlined />,
  },
  {
    path: '/notifications',
    name: '通知管理',
    icon: <BellOutlined />,
  },
  {
    path: '/system',
    name: '系统管理',
    icon: <SettingOutlined />,
    children: [
      {
        path: '/system/users',
        name: '用户管理',
        icon: <TeamOutlined />,
      },
      {
        path: '/system/roles',
        name: '角色管理',
        icon: <SafetyOutlined />,
      },
    ],
  },
];

const Layout: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { theme, toggleTheme } = useThemeStore();
  const { user, logout } = useAuthStore();

  // 处理登出
  const handleLogout = async () => {
    await logout();
    navigate('/login', { replace: true });
  };

  // 用户下拉菜单项
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人信息',
      onClick: () => navigate('/profile'),
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ];

  // 根据权限过滤菜单
  const filterMenuByPermission = (menus: MenuDataItem[]): MenuDataItem[] => {
    return menus.filter((menu) => {
      // 系统管理菜单需要特殊权限检查
      if (menu.path === '/system') {
        // 只有管理员能看到系统管理
        return user?.role === 'admin';
      }
      return true;
    }).map((menu) => {
      if (menu.children) {
        return {
          ...menu,
          children: filterMenuByPermission(menu.children),
        };
      }
      return menu;
    });
  };

  const filteredMenuData = filterMenuByPermission(menuData);

  return (
    <ProLayout
      layout="mix"
      navTheme={theme === 'dark' ? 'realDark' : 'light'}
      fixedHeader
      fixSiderbar
      logo={logo}
      title="CertManager"
      className="modern-layout"
      route={{
        path: '/',
        routes: filteredMenuData,
      }}
      location={{
        pathname: location.pathname,
      }}
      menuItemRender={(item, dom) => (
        <a
          onClick={() => {
            navigate(item.path || '/');
          }}
        >
          {dom}
        </a>
      )}
      rightContentRender={() => (
        <Space style={{ marginRight: 16 }}>
          <Tooltip title={theme === 'dark' ? '切换到亮色主题' : '切换到暗色主题'}>
            <Space>
              {theme === 'dark' ? <BulbFilled /> : <BulbOutlined />}
              <Switch
                checked={theme === 'dark'}
                onChange={toggleTheme}
                checkedChildren="暗"
                unCheckedChildren="亮"
              />
            </Space>
          </Tooltip>
          
          {/* 用户信息 */}
          {user && (
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <Button type="text" style={{ padding: '0 8px' }}>
                <Space>
                  <Avatar size="small" icon={<UserOutlined />} style={{ backgroundColor: '#667eea', border: 'none' }}>
                    {user.username?.charAt(0).toUpperCase()}
                  </Avatar>
                  <span>{user.nickname || user.username}</span>
                </Space>
              </Button>
            </Dropdown>
          )}
        </Space>
      )}
      style={{
        minHeight: '100vh',
      }}
    >
      <Outlet />
    </ProLayout>
  );
};

export default Layout;
