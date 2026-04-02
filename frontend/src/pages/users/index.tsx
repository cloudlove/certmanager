import { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Input,
  Space,
  Tag,
  Badge,
  Popconfirm,
  Modal,
  Form,
  Select,
  message,

  Typography,
  Row,
  Col,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  KeyOutlined,
  UserOutlined,
  SearchOutlined,
  SafetyOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { getUsers, createUser, updateUser, deleteUser, assignUserRole, resetUserPassword, getRoles } from '@/api/user';
import type { User, Role, UserStatus } from '@/types';
import dayjs from 'dayjs';

const { Title } = Typography;
const { Option } = Select;

export default function UserManagePage() {
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [keyword, setKeyword] = useState('');
  
  // 模态框状态
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isRoleModalOpen, setIsRoleModalOpen] = useState(false);
  const [isResetPwdModalOpen, setIsResetPwdModalOpen] = useState(false);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  
  // 表单
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [roleForm] = Form.useForm();
  const [resetPwdForm] = Form.useForm();

  // 加载用户列表
  const loadUsers = async () => {
    setLoading(true);
    try {
      const res = await getUsers({ page, pageSize, keyword });
      setUsers(res.list);
      setTotal(res.total);
    } finally {
      setLoading(false);
    }
  };

  // 加载角色列表
  const loadRoles = async () => {
    try {
      const res = await getRoles();
      setRoles(res);
    } catch {
      // 错误已由拦截器处理
    }
  };

  useEffect(() => {
    loadUsers();
    loadRoles();
  }, [page, pageSize]);

  // 搜索
  const handleSearch = () => {
    setPage(1);
    loadUsers();
  };

  // 创建用户
  const handleCreate = async (values: {
    username: string;
    password: string;
    nickname?: string;
    email?: string;
    roleId: string;
  }) => {
    try {
      await createUser({
        ...values,
        status: 'active',
      });
      message.success('创建用户成功');
      setIsCreateModalOpen(false);
      createForm.resetFields();
      loadUsers();
    } catch {
      // 错误已由拦截器处理
    }
  };

  // 更新用户
  const handleUpdate = async (values: {
    nickname?: string;
    email?: string;
    status: UserStatus;
  }) => {
    if (!currentUser) return;
    try {
      await updateUser(currentUser.id, values);
      message.success('更新用户成功');
      setIsEditModalOpen(false);
      loadUsers();
    } catch {
      // 错误已由拦截器处理
    }
  };

  // 删除用户
  const handleDelete = async (id: string) => {
    try {
      await deleteUser(id);
      message.success('删除用户成功');
      loadUsers();
    } catch {
      // 错误已由拦截器处理
    }
  };

  // 分配角色
  const handleAssignRole = async (values: { roleId: string }) => {
    if (!currentUser) return;
    try {
      await assignUserRole(currentUser.id, { roleId: parseInt(values.roleId, 10) });
      message.success('分配角色成功');
      setIsRoleModalOpen(false);
      loadUsers();
    } catch {
      // 错误已由拦截器处理
    }
  };

  // 重置密码
  const handleResetPassword = async (values: { newPassword: string }) => {
    if (!currentUser) return;
    try {
      await resetUserPassword(currentUser.id, { newPassword: values.newPassword });
      message.success('重置密码成功');
      setIsResetPwdModalOpen(false);
      resetPwdForm.resetFields();
    } catch {
      // 错误已由拦截器处理
    }
  };

  // 打开编辑模态框
  const openEditModal = (user: User) => {
    setCurrentUser(user);
    editForm.setFieldsValue({
      nickname: user.nickname,
      email: user.email,
      status: user.status,
    });
    setIsEditModalOpen(true);
  };

  // 打开分配角色模态框
  const openRoleModal = (user: User) => {
    setCurrentUser(user);
    roleForm.setFieldsValue({
      roleId: user.roleId || user.roles?.[0]?.id || '',
    });
    setIsRoleModalOpen(true);
  };

  // 打开重置密码模态框
  const openResetPwdModal = (user: User) => {
    setCurrentUser(user);
    setIsResetPwdModalOpen(true);
  };

  const columns: ColumnsType<User> = [
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      render: (text, record) => (
        <Space>
          <UserOutlined />
          <span>{text}</span>
          {record.role === 'admin' && <Tag color="red">管理员</Tag>}
        </Space>
      ),
    },
    {
      title: '昵称',
      dataIndex: 'nickname',
      key: 'nickname',
      render: (text) => text || '-',
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      render: (text) => text || '-',
    },
    {
      title: '角色',
      dataIndex: 'roleName',
      key: 'roleName',
      render: (roleName?: string, record?: User) => (
        <Space wrap>
          {roleName || record?.roles?.[0]?.name ? (
            <Tag color="blue">
              {roleName || record?.roles?.[0]?.name}
            </Tag>
          ) : '-'}
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: UserStatus) => (
        <Badge
          status={status === 'active' ? 'success' : 'error'}
          text={status === 'active' ? '正常' : '禁用'}
        />
      ),
    },
    {
      title: '最后登录',
      dataIndex: 'lastLoginAt',
      key: 'lastLoginAt',
      render: (text) => (text ? dayjs(text).format('YYYY-MM-DD HH:mm') : '-'),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (text) => dayjs(text).format('YYYY-MM-DD'),
    },
    {
      title: '操作',
      key: 'action',
      width: 240,
      render: (_, record) => (
        <Space size="small">
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEditModal(record)}
          >
            编辑
          </Button>
          <Button
            type="text"
            size="small"
            icon={<SafetyOutlined />}
            onClick={() => openRoleModal(record)}
          >
            角色
          </Button>
          <Button
            type="text"
            size="small"
            icon={<KeyOutlined />}
            onClick={() => openResetPwdModal(record)}
          >
            重置密码
          </Button>
          <Popconfirm
            title="确认删除"
            description={`确定要删除用户 "${record.username}" 吗？`}
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="text" danger size="small" icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card>
      <Row justify="space-between" align="middle" style={{ marginBottom: 16 }}>
        <Col>
          <Title level={4} style={{ margin: 0 }}>
            用户管理
          </Title>
        </Col>
        <Col>
          <Space>
            <Input
              placeholder="搜索用户名/邮箱"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onPressEnter={handleSearch}
              prefix={<SearchOutlined />}
              style={{ width: 220 }}
              allowClear
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsCreateModalOpen(true)}>
              创建用户
            </Button>
          </Space>
        </Col>
      </Row>

      <Table
        columns={columns}
        dataSource={users}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          onChange: (p, ps) => {
            setPage(p);
            if (ps) setPageSize(ps);
          },
        }}
      />

      {/* 创建用户模态框 */}
      <Modal
        title="创建用户"
        open={isCreateModalOpen}
        onCancel={() => {
          setIsCreateModalOpen(false);
          createForm.resetFields();
        }}
        onOk={() => createForm.submit()}
        okText="创建"
        cancelText="取消"
      >
        <Form form={createForm} layout="vertical" onFinish={handleCreate}>
          <Form.Item
            name="username"
            label="用户名"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3个字符' },
            ]}
          >
            <Input placeholder="请输入用户名" />
          </Form.Item>
          <Form.Item
            name="password"
            label="密码"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6个字符' },
            ]}
          >
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
          <Form.Item name="nickname" label="昵称">
            <Input placeholder="请输入昵称" />
          </Form.Item>
          <Form.Item name="email" label="邮箱">
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          <Form.Item name="roleId" label="角色">
            <Select placeholder="请选择角色" allowClear>
              {roles.map((role) => (
                <Option key={role.id} value={role.id}>
                  {role.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑用户模态框 */}
      <Modal
        title="编辑用户"
        open={isEditModalOpen}
        onCancel={() => setIsEditModalOpen(false)}
        onOk={() => editForm.submit()}
        okText="保存"
        cancelText="取消"
      >
        <Form form={editForm} layout="vertical" onFinish={handleUpdate}>
          <Form.Item name="nickname" label="昵称">
            <Input placeholder="请输入昵称" />
          </Form.Item>
          <Form.Item name="email" label="邮箱">
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          <Form.Item name="status" label="状态">
            <Select placeholder="请选择状态">
              <Option value="active">正常</Option>
              <Option value="disabled">禁用</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* 分配角色模态框 */}
      <Modal
        title={`分配角色 - ${currentUser?.username}`}
        open={isRoleModalOpen}
        onCancel={() => setIsRoleModalOpen(false)}
        onOk={() => roleForm.submit()}
        okText="保存"
        cancelText="取消"
      >
        <Form form={roleForm} layout="vertical" onFinish={handleAssignRole}>
          <Form.Item name="roleId" label="角色">
            <Select placeholder="请选择角色" allowClear style={{ width: '100%' }}>
              {roles.map((role) => (
                <Option key={role.id} value={role.id}>
                  {role.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Form>
      </Modal>

      {/* 重置密码模态框 */}
      <Modal
        title={`重置密码 - ${currentUser?.username}`}
        open={isResetPwdModalOpen}
        onCancel={() => {
          setIsResetPwdModalOpen(false);
          resetPwdForm.resetFields();
        }}
        onOk={() => resetPwdForm.submit()}
        okText="重置"
        cancelText="取消"
      >
        <Form form={resetPwdForm} layout="vertical" onFinish={handleResetPassword}>
          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 8, message: '密码至少8个字符' },
            ]}
          >
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="确认密码"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请确认密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password placeholder="请再次输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
