import { useNavigate } from 'react-router-dom';
import { Card, Form, Input, Button, Checkbox, message, Space, Typography } from 'antd';
import { UserOutlined, LockOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { useAuthStore } from '@/store/auth';

const { Title, Text } = Typography;

interface LoginFormValues {
  username: string;
  password: string;
  remember: boolean;
}

export default function LoginPage() {
  const navigate = useNavigate();
  const { login, isLoading } = useAuthStore();
  const [form] = Form.useForm();

  // 从 localStorage 读取记住的用户名
  const rememberedUsername = localStorage.getItem('rememberedUsername') || '';

  const handleSubmit = async (values: LoginFormValues) => {
    try {
      await login(values.username, values.password);
      
      // 处理记住我
      if (values.remember) {
        localStorage.setItem('rememberedUsername', values.username);
      } else {
        localStorage.removeItem('rememberedUsername');
      }
      
      message.success('登录成功');
      navigate('/', { replace: true });
    } catch (error) {
      // 错误已由 request.ts 拦截器处理
      console.error('登录失败:', error);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: '24px',
      }}
    >
      <Card
        style={{
          width: '100%',
          maxWidth: 420,
          borderRadius: 12,
          boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3)',
        }}
        bodyStyle={{ padding: '40px 32px' }}
      >
        <Space direction="vertical" align="center" style={{ width: '100%', marginBottom: 32 }}>
          <div
            style={{
              width: 64,
              height: 64,
              borderRadius: '50%',
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <SafetyCertificateOutlined style={{ fontSize: 32, color: '#fff' }} />
          </div>
          <Title level={3} style={{ margin: 0, color: '#1f2937' }}>
            CertManager
          </Title>
          <Text type="secondary">证书管理平台</Text>
        </Space>

        <Form
          form={form}
          name="login"
          initialValues={{ remember: true, username: rememberedUsername }}
          onFinish={handleSubmit}
          size="large"
        >
          <Form.Item
            name="username"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3个字符' },
            ]}
          >
            <Input
              prefix={<UserOutlined style={{ color: '#9ca3af' }} />}
              placeholder="用户名"
              autoComplete="username"
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6个字符' },
            ]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: '#9ca3af' }} />}
              placeholder="密码"
              autoComplete="current-password"
            />
          </Form.Item>

          <Form.Item>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Form.Item name="remember" valuePropName="checked" noStyle>
                <Checkbox>记住我</Checkbox>
              </Form.Item>
            </div>
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={isLoading}
              block
              size="large"
              style={{
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                border: 'none',
                height: 48,
                fontSize: 16,
              }}
            >
              登录
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: 'center', marginTop: 16 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            默认管理员账号: admin / admin
          </Text>
        </div>
      </Card>
    </div>
  );
}
