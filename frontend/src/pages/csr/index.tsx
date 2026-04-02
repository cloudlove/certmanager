import React, { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Button,
  Table,
  Tag,
  Space,
  Input,
  Popconfirm,
  message,
  Tooltip,
} from 'antd';
import {
  PlusOutlined,
  EyeOutlined,
  DownloadOutlined,
  KeyOutlined,
  DeleteOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import type { CSRRecord } from '@/types';
import { getCSRList, getCSR, deleteCSR, downloadCSR, downloadPrivateKey } from '@/api/csr';
import GenerateCSRForm from './components/GenerateCSRForm';
import CSRDetail from './components/CSRDetail';

const CSRPage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<CSRRecord[]>([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [searchKeyword, setSearchKeyword] = useState('');
  const [generateModalOpen, setGenerateModalOpen] = useState(false);
  const [detailDrawerOpen, setDetailDrawerOpen] = useState(false);
  const [selectedRecord, setSelectedRecord] = useState<CSRRecord | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getCSRList({
        page: pagination.current,
        pageSize: pagination.pageSize,
        search: searchKeyword || undefined,
      });
      setData(res.list);
      setPagination((prev) => ({
        ...prev,
        total: res.total,
      }));
    } catch {
      message.error('获取 CSR 列表失败');
    } finally {
      setLoading(false);
    }
  }, [pagination.current, pagination.pageSize, searchKeyword]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleSearch = (value: string) => {
    setSearchKeyword(value);
    setPagination((prev) => ({ ...prev, current: 1 }));
  };

  const handleDelete = async (id: number) => {
    try {
      await deleteCSR(id);
      message.success('删除成功');
      fetchData();
    } catch {
      message.error('删除失败');
    }
  };

  const handleDownloadCSR = async (record: CSRRecord) => {
    try {
      const blob = await downloadCSR(record.id);
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${record.commonName}.csr.pem`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      message.success('CSR 下载成功');
    } catch {
      message.error('下载失败');
    }
  };

  const handleDownloadPrivateKey = async (record: CSRRecord) => {
    try {
      const blob = await downloadPrivateKey(record.id);
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${record.commonName}.key.pem`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      message.success('私钥下载成功');
    } catch {
      message.error('下载失败');
    }
  };

  const handleViewDetail = async (record: CSRRecord) => {
    try {
      const detail = await getCSR(record.id);
      setSelectedRecord(detail);
      setDetailDrawerOpen(true);
    } catch {
      message.error('获取 CSR 详情失败');
    }
  };

  const handleGenerateSuccess = () => {
    setGenerateModalOpen(false);
    fetchData();
  };

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      active: { color: 'success', text: '有效' },
      revoked: { color: 'error', text: '已吊销' },
    };
    const config = statusMap[status] || { color: 'default', text: status };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const getKeyTypeTag = (keyAlgorithm: string) => {
    const color = keyAlgorithm === 'RSA' ? 'blue' : 'green';
    return <Tag color={color}>{keyAlgorithm}</Tag>;
  };

  const columns = [
    {
      title: 'Common Name',
      dataIndex: 'commonName',
      key: 'commonName',
      ellipsis: true,
    },
    {
      title: 'SAN',
      dataIndex: 'san',
      key: 'san',
      render: (san: string[]) => (
        <Space size={[0, 4]} wrap style={{ maxWidth: 300 }}>
          {san?.slice(0, 3).map((s, index) => (
            <Tag key={index}>{s}</Tag>
          ))}
          {san?.length > 3 && (
            <Tooltip title={san.slice(3).join(', ')}>
              <Tag>+{san.length - 3}</Tag>
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: '密钥算法',
      dataIndex: 'keyAlgorithm',
      key: 'keyAlgorithm',
      width: 100,
      render: (keyAlgorithm: string) => getKeyTypeTag(keyAlgorithm),
    },
    {
      title: '密钥长度',
      dataIndex: 'keySize',
      key: 'keySize',
      width: 100,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (createdAt: string) => new Date(createdAt).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right' as const,
      render: (_: unknown, record: CSRRecord) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          <Tooltip title="下载 CSR">
            <Button
              type="text"
              icon={<DownloadOutlined />}
              onClick={() => handleDownloadCSR(record)}
            />
          </Tooltip>
          <Tooltip title="下载私钥">
            <Button
              type="text"
              icon={<KeyOutlined />}
              onClick={() => handleDownloadPrivateKey(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确认删除"
            description="删除后无法恢复，是否继续？"
            onConfirm={() => handleDelete(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button type="text" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: '24px', background: '#fafafa', minHeight: '100vh' }}>
      <Card 
        style={{ borderRadius: 12, boxShadow: '0 4px 12px rgba(0,0,0,0.05)' }}
        styles={{ body: { padding: '24px' } }}
      >
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 24,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div style={{
              width: 48,
              height: 48,
              borderRadius: '50%',
              background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center'
            }}>
              <KeyOutlined style={{ fontSize: 24, color: 'white' }} />
            </div>
            <div>
              <h2 style={{ margin: 0, fontSize: 24, fontWeight: 600, color: '#1d1d1d' }}>CSR 管理</h2>
              <p style={{ margin: '4px 0 0 0', color: '#888', fontSize: 14 }}>
                管理您的证书签名请求 (Certificate Signing Request)
              </p>
            </div>
          </div>
          <Button
            type="primary"
            size="large"
            icon={<PlusOutlined />}
            onClick={() => setGenerateModalOpen(true)}
            style={{ borderRadius: 8, height: 44 }}
          >
            生成 CSR
          </Button>
        </div>

        <div style={{ marginBottom: 24 }} role="search">
          <Input.Search
            placeholder="按 Common Name 搜索 CSR"
            allowClear
            enterButton={<><SearchOutlined aria-label="搜索" /> 搜索</>}
            onSearch={handleSearch}
            style={{ width: 400, borderRadius: 8 }}
            size="large"
            aria-label="CSR 搜索框"
          />
        </div>

        <Table
          columns={columns}
          dataSource={data}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条记录`,
            onChange: (page, pageSize) => {
              setPagination({ current: page, pageSize: pageSize || 10, total: pagination.total });
            },
          }}
          scroll={{ x: 1200 }}
          size="middle"
          style={{ borderRadius: 8, overflow: 'hidden' }}
        />

        <GenerateCSRForm
          open={generateModalOpen}
          onCancel={() => setGenerateModalOpen(false)}
          onSuccess={handleGenerateSuccess}
        />

        <CSRDetail
          open={detailDrawerOpen}
          onClose={() => {
            setDetailDrawerOpen(false);
            setSelectedRecord(null);
          }}
          record={selectedRecord}
        />
      </Card>
    </div>
  );
};

export default CSRPage;
