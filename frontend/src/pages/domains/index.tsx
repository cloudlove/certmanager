import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Typography,
  Button,
  Space,
  Table,
  Input,
  Tag,
  Popconfirm,
  message,
  Tooltip,
} from 'antd';
import type { TablePaginationConfig } from 'antd/es/table';
import {
  PlusOutlined,
  SafetyOutlined,
  EditOutlined,
  DeleteOutlined,
  SafetyCertificateOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import DomainForm from './components/DomainForm';
import {
  getDomains,
  deleteDomain,
  verifyDomain,
  batchVerifyDomains,
  type DomainWithVerify,
  type VerifyStatus,
} from '@/api/domain';

const { Title } = Typography;
const { Search } = Input;

// 校验状态映射
const VERIFY_STATUS_MAP: Record<VerifyStatus, { color: string; text: string }> = {
  normal: { color: 'success', text: '正常' },
  mismatch: { color: 'error', text: '不匹配' },
  expired: { color: 'error', text: '已过期' },
  error: { color: 'warning', text: '校验失败' },
  unchecked: { color: 'default', text: '未校验' },
};

const DomainPage: React.FC = () => {
  const navigate = useNavigate();
  const [domains, setDomains] = useState<DomainWithVerify[]>([]);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState<TablePaginationConfig>({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [searchKeyword, setSearchKeyword] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [verifyingIds, setVerifyingIds] = useState<Set<number>>(new Set());
  const [batchVerifying, setBatchVerifying] = useState(false);

  // 表单弹窗状态
  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<'create' | 'edit'>('create');
  const [editingDomain, setEditingDomain] = useState<DomainWithVerify | undefined>();

  // 获取域名列表
  const fetchDomains = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getDomains({
        page: pagination.current,
        pageSize: pagination.pageSize,
        search: searchKeyword || undefined,
      });
      setDomains(res.list);
      setPagination({
        ...pagination,
        total: res.total,
      });
    } catch {
      // 错误已在 request.ts 中处理
    } finally {
      setLoading(false);
    }
  }, [pagination.current, pagination.pageSize, searchKeyword]);

  useEffect(() => {
    fetchDomains();
  }, [fetchDomains]);

  // 处理分页变化
  const handleTableChange = (newPagination: TablePaginationConfig) => {
    setPagination(newPagination);
  };

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchKeyword(value);
    setPagination({ ...pagination, current: 1 });
  };

  // 打开添加弹窗
  const handleAdd = () => {
    setFormMode('create');
    setEditingDomain(undefined);
    setFormOpen(true);
  };

  // 打开编辑弹窗
  const handleEdit = (record: DomainWithVerify) => {
    setFormMode('edit');
    setEditingDomain(record);
    setFormOpen(true);
  };

  // 关闭弹窗
  const handleFormCancel = () => {
    setFormOpen(false);
    setEditingDomain(undefined);
  };

  // 表单提交成功
  const handleFormSuccess = () => {
    setFormOpen(false);
    setEditingDomain(undefined);
    fetchDomains();
  };

  // 删除域名
  const handleDelete = async (id: number) => {
    try {
      await deleteDomain(id);
      message.success('域名删除成功');
      fetchDomains();
    } catch {
      // 错误已在 request.ts 中处理
    }
  };

  // 单个校验
  const handleVerify = async (record: DomainWithVerify) => {
    const id = parseInt(record.id);
    setVerifyingIds((prev) => new Set(prev).add(id));
    try {
      const result = await verifyDomain(id);
      message.success(result.message || '校验完成');
      // 刷新当前行数据
      fetchDomains();
    } catch {
      // 错误已在 request.ts 中处理
    } finally {
      setVerifyingIds((prev) => {
        const newSet = new Set(prev);
        newSet.delete(id);
        return newSet;
      });
    }
  };

  // 批量校验
  const handleBatchVerify = async () => {
    if (selectedRowKeys.length === 0) return;

    setBatchVerifying(true);
    try {
      const ids = selectedRowKeys.map((key) => parseInt(key as string));
      const result = await batchVerifyDomains(ids);

      // 统计结果
      const successCount = result.results.filter(
        (r) => r.status === 'normal'
      ).length;
      const failCount = result.results.length - successCount;

      if (failCount === 0) {
        message.success(`全部 ${successCount} 个域名校验成功`);
      } else {
        message.warning(`${successCount} 个成功，${failCount} 个失败`);
      }

      setSelectedRowKeys([]);
      fetchDomains();
    } catch {
      // 错误已在 request.ts 中处理
    } finally {
      setBatchVerifying(false);
    }
  };

  // 表格列定义
  const columns = [
    {
      title: '域名',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: '关联证书',
      dataIndex: 'certificateName',
      key: 'certificateName',
      render: (text: string, record: DomainWithVerify) => {
        if (!text || !record.certificateId) return '-';
        return (
          <a
            onClick={() => navigate(`/certificates/${record.certificateId}`)}
          >
            {text}
          </a>
        );
      },
    },
    {
      title: '校验状态',
      dataIndex: 'verifyStatus',
      key: 'verifyStatus',
      render: (status: VerifyStatus) => {
        const { color, text } = VERIFY_STATUS_MAP[status] || VERIFY_STATUS_MAP.unchecked;
        return <Tag color={color}>{text}</Tag>;
      },
    },
    {
      title: '上次校验时间',
      dataIndex: 'lastCheckAt',
      key: 'lastCheckAt',
      render: (text: string) => {
        return text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-';
      },
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (text: string) => dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      render: (_: unknown, record: DomainWithVerify) => {
        const id = parseInt(record.id);
        const isVerifying = verifyingIds.has(id);

        return (
          <Space size="small">
            <Tooltip title="编辑">
              <Button
                type="text"
                icon={<EditOutlined />}
                onClick={() => handleEdit(record)}
              />
            </Tooltip>
            <Tooltip title="校验">
              <Button
                type="text"
                icon={<SafetyOutlined />}
                loading={isVerifying}
                onClick={() => handleVerify(record)}
              />
            </Tooltip>
            <Popconfirm
              title="确认删除"
              description={`确定要删除域名 "${record.name}" 吗？`}
              onConfirm={() => handleDelete(id)}
              okText="删除"
              cancelText="取消"
              okButtonProps={{ danger: true }}
            >
              <Tooltip title="删除">
                <Button type="text" danger icon={<DeleteOutlined />} />
              </Tooltip>
            </Popconfirm>
          </Space>
        );
      },
    },
  ];

  // 行选择配置
  const rowSelection = {
    selectedRowKeys,
    onChange: (newSelectedRowKeys: React.Key[]) => {
      setSelectedRowKeys(newSelectedRowKeys);
    },
  };

  return (
    <>
      <Card>
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          {/* 标题栏 */}
          <div
            style={{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}
          >
            <Title level={2} style={{ margin: 0 }}>
              域名管理
            </Title>
            <Space>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleAdd}
              >
                添加域名
              </Button>
              <Button
                icon={<SafetyCertificateOutlined />}
                onClick={handleBatchVerify}
                disabled={selectedRowKeys.length === 0}
                loading={batchVerifying}
              >
                批量校验
                {selectedRowKeys.length > 0 && ` (${selectedRowKeys.length})`}
              </Button>
            </Space>
          </div>

          {/* 搜索栏 */}
          <Search
            placeholder="搜索域名"
            allowClear
            enterButton
            onSearch={handleSearch}
            style={{ width: 300 }}
          />

          {/* 表格 */}
          <Table
            rowKey="id"
            rowSelection={rowSelection}
            columns={columns}
            dataSource={domains}
            loading={loading}
            pagination={pagination}
            onChange={handleTableChange}
          />
        </Space>
      </Card>

      {/* 表单弹窗 */}
      <DomainForm
        open={formOpen}
        mode={formMode}
        initialData={editingDomain}
        onSuccess={handleFormSuccess}
        onCancel={handleFormCancel}
      />
    </>
  );
};

export default DomainPage;
