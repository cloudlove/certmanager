import { useState, useEffect, useMemo } from 'react';
import {
  Card,
  Table,
  Button,
  Input,
  Space,
  Tag,
  Popconfirm,
  Modal,
  Form,
  Tree,
  message,
  Typography,
  Row,
  Col,

} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SafetyOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { getRoles, createRole, updateRole, deleteRole, assignRolePermissions, getPermissions } from '@/api/user';
import type { Role, Permission } from '@/types';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { TextArea } = Input;

// 权限资源分组配置
const PERMISSION_GROUPS = [
  { resource: 'certificate', label: '证书管理' },
  { resource: 'csr', label: 'CSR 管理' },
  { resource: 'domain', label: '域名管理' },
  { resource: 'credential', label: '凭证管理' },
  { resource: 'deployment', label: '部署管理' },
  { resource: 'nginx', label: 'Nginx 管理' },
  { resource: 'notification', label: '通知管理' },
  { resource: 'user', label: '用户管理' },
  { resource: 'role', label: '角色管理' },
  { resource: 'audit', label: '审计日志' },
];

// 操作类型
const ACTION_LABELS: Record<string, string> = {
  read: '查看',
  write: '编辑',
  delete: '删除',
  execute: '执行',
};

export default function RoleManagePage() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');

  // 模态框状态
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isPermissionModalOpen, setIsPermissionModalOpen] = useState(false);
  const [currentRole, setCurrentRole] = useState<Role | null>(null);

  // 表单
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();
  const [permissionForm] = Form.useForm();

  // 选中的权限
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>([]);
  const [expandedKeys, setExpandedKeys] = useState<string[]>([]);

  // 加载角色列表
  const loadRoles = async () => {
    setLoading(true);
    try {
      const res = await getRoles();
      setRoles(res);
    } finally {
      setLoading(false);
    }
  };

  // 加载权限列表
  const loadPermissions = async () => {
    try {
      const res = await getPermissions();
      setPermissions(res);
    } catch {
      // 错误已由拦截器处理
    }
  };

  useEffect(() => {
    loadRoles();
    loadPermissions();
  }, []);

  // 按资源分组的权限
  const permissionTreeData = useMemo(() => {
    const groups = PERMISSION_GROUPS.map((group) => {
      const groupPermissions = permissions.filter((p) => p.resource === group.resource);
      if (groupPermissions.length === 0) return null;

      return {
        title: group.label,
        key: group.resource,
        children: groupPermissions.map((p) => ({
          title: `${ACTION_LABELS[p.action] || p.action} (${p.name})`,
          key: p.id,
          isLeaf: true,
          permission: p,
        })),
      };
    }).filter(Boolean);

    return groups;
  }, [permissions]);

  // 过滤后的角色列表
  const filteredRoles = useMemo(() => {
    if (!keyword) return roles;
    return roles.filter(
      (r) =>
        r.name.toLowerCase().includes(keyword.toLowerCase()) ||
        r.description?.toLowerCase().includes(keyword.toLowerCase())
    );
  }, [roles, keyword]);

  // 创建角色
  const handleCreate = async (values: { name: string; description?: string }) => {
    try {
      await createRole(values);
      message.success('创建角色成功');
      setIsCreateModalOpen(false);
      createForm.resetFields();
      loadRoles();
    } catch {
      // 错误已由拦截器处理
    }
  };

  // 更新角色
  const handleUpdate = async (values: { name?: string; description?: string }) => {
    if (!currentRole) return;
    try {
      await updateRole(currentRole.id, values);
      message.success('更新角色成功');
      setIsEditModalOpen(false);
      loadRoles();
    } catch {
      // 错误已由拦截器处理
    }
  };

  // 删除角色
  const handleDelete = async (id: string) => {
    try {
      await deleteRole(id);
      message.success('删除角色成功');
      loadRoles();
    } catch {
      // 错误已由拦截器处理
    }
  };

  // 分配权限
  const handleAssignPermissions = async () => {
    if (!currentRole) return;
    try {
      await assignRolePermissions(currentRole.id, { permissionIds: selectedPermissions });
      message.success('分配权限成功');
      setIsPermissionModalOpen(false);
      loadRoles();
    } catch {
      // 错误已由拦截器处理
    }
  };

  // 打开编辑模态框
  const openEditModal = (role: Role) => {
    setCurrentRole(role);
    editForm.setFieldsValue({
      name: role.name,
      description: role.description,
    });
    setIsEditModalOpen(true);
  };

  // 打开权限分配模态框
  const openPermissionModal = (role: Role) => {
    setCurrentRole(role);
    const permissionIds = role.permissions?.map((p) => p.id) || [];
    setSelectedPermissions(permissionIds);
    permissionForm.setFieldsValue({
      permissionIds,
    });
    // 展开所有分组
    setExpandedKeys(PERMISSION_GROUPS.map((g) => g.resource));
    setIsPermissionModalOpen(true);
  };

  // 处理权限选择
  const handlePermissionCheck = (checkedKeys: string[]) => {
    setSelectedPermissions(checkedKeys);
  };

  // 全选/取消全选
  const handleSelectAll = () => {
    const allIds = permissions.map((p) => p.id);
    setSelectedPermissions(allIds);
  };

  const handleDeselectAll = () => {
    setSelectedPermissions([]);
  };

  const columns: ColumnsType<Role> = [
    {
      title: '角色名称',
      dataIndex: 'name',
      key: 'name',
      render: (text) => (
        <Space>
          <SafetyOutlined />
          <span>{text}</span>
        </Space>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      render: (text) => text || '-',
    },
    {
      title: '权限数量',
      dataIndex: 'permissions',
      key: 'permissions',
      render: (permissions?: Permission[]) => (
        <Tag color="blue">{permissions?.length || 0} 个权限</Tag>
      ),
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
            onClick={() => openPermissionModal(record)}
          >
            权限
          </Button>
          <Popconfirm
            title="确认删除"
            description={`确定要删除角色 "${record.name}" 吗？`}
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
            角色管理
          </Title>
        </Col>
        <Col>
          <Space>
            <Input
              placeholder="搜索角色名称"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              prefix={<SearchOutlined />}
              style={{ width: 220 }}
              allowClear
            />
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setIsCreateModalOpen(true)}>
              创建角色
            </Button>
          </Space>
        </Col>
      </Row>

      <Table
        columns={columns}
        dataSource={filteredRoles}
        rowKey="id"
        loading={loading}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
      />

      {/* 创建角色模态框 */}
      <Modal
        title="创建角色"
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
            name="name"
            label="角色名称"
            rules={[{ required: true, message: '请输入角色名称' }]}
          >
            <Input placeholder="请输入角色名称" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <TextArea rows={3} placeholder="请输入角色描述" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑角色模态框 */}
      <Modal
        title="编辑角色"
        open={isEditModalOpen}
        onCancel={() => setIsEditModalOpen(false)}
        onOk={() => editForm.submit()}
        okText="保存"
        cancelText="取消"
      >
        <Form form={editForm} layout="vertical" onFinish={handleUpdate}>
          <Form.Item name="name" label="角色名称">
            <Input placeholder="请输入角色名称" disabled />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <TextArea rows={3} placeholder="请输入角色描述" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 权限分配模态框 */}
      <Modal
        title={`分配权限 - ${currentRole?.name}`}
        open={isPermissionModalOpen}
        onCancel={() => setIsPermissionModalOpen(false)}
        onOk={handleAssignPermissions}
        okText="保存"
        cancelText="取消"
        width={600}
      >
        <Space style={{ marginBottom: 16 }}>
          <Button size="small" onClick={handleSelectAll}>
            全选
          </Button>
          <Button size="small" onClick={handleDeselectAll}>
            取消全选
          </Button>
          <Text type="secondary">已选择 {selectedPermissions.length} 个权限</Text>
        </Space>
        <Form form={permissionForm}>
          <Form.Item name="permissionIds">
            <Tree
              checkable
              checkedKeys={selectedPermissions}
              onCheck={(checked) => handlePermissionCheck(checked as string[])}
              expandedKeys={expandedKeys}
              onExpand={(keys) => setExpandedKeys(keys as string[])}
              treeData={permissionTreeData as unknown as Parameters<typeof Tree>[0]['treeData']}
              style={{ maxHeight: 400, overflow: 'auto' }}
            />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
