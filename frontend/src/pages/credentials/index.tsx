import { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Button,
  Table,
  Tag,
  Select,
  Space,
  Popconfirm,
  message,
  Empty,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ApiOutlined,
} from '@ant-design/icons';
import { getCredentials, deleteCredential, testCredential } from '@/api/credential';
import CredentialForm from './components/CredentialForm';
import type { CloudCredential } from '@/types';
import type { ColumnsType } from 'antd/es/table';

const { Option } = Select;

// Provider 类型选项
const providerOptions = [
  { value: '', label: '全部' },
  { value: 'aliyun', label: '阿里云' },
  { value: 'tencent', label: '腾讯云' },
  { value: 'volcengine', label: '火山云' },
  { value: 'wangsu', label: '网宿' },
  { value: 'aws', label: 'AWS' },
  { value: 'azure', label: 'Azure' },
];

// Provider 标签颜色和显示名称映射
const providerConfig: Record<string, { color: string; label: string }> = {
  aliyun: { color: 'orange', label: '阿里云' },
  tencent: { color: 'blue', label: '腾讯云' },
  volcengine: { color: 'red', label: '火山云' },
  wangsu: { color: 'green', label: '网宿' },
  aws: { color: 'purple', label: 'AWS' },
  azure: { color: 'geekblue', label: 'Azure' },
};

// AccessKey 脱敏处理
const maskAccessKey = (accessKey: string): string => {
  if (!accessKey || accessKey.length <= 8) {
    return '****';
  }
  return accessKey.substring(0, 4) + '****' + accessKey.substring(accessKey.length - 4);
};

const CredentialPage: React.FC = () => {
  const [credentials, setCredentials] = useState<CloudCredential[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [providerType, setProviderType] = useState<string>('');
  const [formOpen, setFormOpen] = useState(false);
  const [editingCredential, setEditingCredential] = useState<CloudCredential | null>(null);
  const [testingId, setTestingId] = useState<string | null>(null);

  // 获取凭证列表
  const fetchCredentials = useCallback(async () => {
    setLoading(true);
    try {
      const params: { page: number; pageSize: number; providerType?: string } = {
        page,
        pageSize,
      };
      if (providerType) {
        params.providerType = providerType;
      }
      const res = await getCredentials(params);
      setCredentials(res.list);
      setTotal(res.total);
    } catch (error) {
      console.error('获取凭证列表失败:', error);
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, providerType]);

  useEffect(() => {
    fetchCredentials();
  }, [fetchCredentials]);

  // 处理 Provider 筛选变化
  const handleProviderChange = (value: string) => {
    setProviderType(value);
    setPage(1);
  };

  // 处理分页变化
  const handlePageChange = (newPage: number, newPageSize?: number) => {
    setPage(newPage);
    if (newPageSize) {
      setPageSize(newPageSize);
    }
  };

  // 打开新建抽屉
  const handleCreate = () => {
    setEditingCredential(null);
    setFormOpen(true);
  };

  // 打开编辑抽屉
  const handleEdit = (record: CloudCredential) => {
    setEditingCredential(record);
    setFormOpen(true);
  };

  // 关闭抽屉
  const handleCloseForm = () => {
    setFormOpen(false);
    setEditingCredential(null);
  };

  // 表单提交成功回调
  const handleFormSuccess = () => {
    fetchCredentials();
  };

  // 删除凭证
  const handleDelete = async (id: string) => {
    try {
      await deleteCredential(id);
      message.success('凭证删除成功');
      fetchCredentials();
    } catch (error) {
      console.error('删除凭证失败:', error);
    }
  };

  // 测试凭证连通性
  const handleTest = async (id: string) => {
    setTestingId(id);
    try {
      const result = await testCredential(id);
      if (result.success) {
        message.success(result.message || '连接测试成功');
      } else {
        message.error(result.message || '连接测试失败');
      }
    } catch (error) {
      message.error('连接测试失败');
    } finally {
      setTestingId(null);
    }
  };

  const columns: ColumnsType<CloudCredential> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 180,
    },
    {
      title: 'Provider 类型',
      dataIndex: 'providerType',
      key: 'providerType',
      width: 120,
      render: (providerType: string) => {
        const config = providerConfig[providerType] || { color: 'default', label: providerType };
        return <Tag color={config.color}>{config.label}</Tag>;
      },
    },
    {
      title: 'Access Key',
      dataIndex: 'accessKey',
      key: 'accessKey',
      width: 180,
      render: (accessKey: string) => maskAccessKey(accessKey),
    },
    {
      title: '状态',
      dataIndex: 'isValid',
      key: 'isValid',
      width: 100,
      render: (isValid: boolean) => (
        <Tag color={isValid ? 'success' : 'error'}>
          {isValid ? '有效' : '无效'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right',
      render: (_, record) => (
        <Space size="small">
          <Button
            type="text"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Button
            type="text"
            size="small"
            icon={<ApiOutlined />}
            loading={testingId === record.id}
            onClick={() => handleTest(record.id)}
          >
            测试连通
          </Button>
          <Popconfirm
            title="确定删除该凭证吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="text"
              size="small"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card>
      {/* 页面标题和操作栏 */}
      <div style={{ marginBottom: 24 }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 16,
          }}
        >
          <h2 style={{ margin: 0, fontSize: 20, fontWeight: 500 }}>云凭证管理</h2>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新建凭证
          </Button>
        </div>

        {/* 筛选栏 */}
        <Space>
          <span>Provider 类型：</span>
          <Select
            value={providerType}
            onChange={handleProviderChange}
            style={{ width: 160 }}
            placeholder="全部"
          >
            {providerOptions.map((option) => (
              <Option key={option.value} value={option.value}>
                {option.label}
              </Option>
            ))}
          </Select>
        </Space>
      </div>

      {/* 凭证列表表格 */}
      <Table
        columns={columns}
        dataSource={credentials}
        rowKey="id"
        loading={loading}
        pagination={{
          current: page,
          pageSize,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条`,
          onChange: handlePageChange,
        }}
        locale={{
          emptyText: <Empty description="暂无凭证数据" />,
        }}
        scroll={{ x: 800 }}
      />

      {/* 新建/编辑抽屉 */}
      <CredentialForm
        open={formOpen}
        credential={editingCredential}
        onClose={handleCloseForm}
        onSuccess={handleFormSuccess}
      />
    </Card>
  );
};

export default CredentialPage;
