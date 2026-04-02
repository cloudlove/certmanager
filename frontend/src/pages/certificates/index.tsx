import { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Typography,
  Button,
  Space,
  Table,
  Tag,
  Badge,
  Input,
  Select,
  Popconfirm,
  message,
  Tooltip,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  PlusOutlined,
  ImportOutlined,
  EyeOutlined,
  SyncOutlined,
  DeleteOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import type { Certificate } from '@/types';
import {
  getCertificates,
  deleteCertificate,
  syncCertStatus,
  type CertificateQueryParams,
} from '@/api/certificate';
import ApplyCertForm from './components/ApplyCertForm';
import ImportCertForm from './components/ImportCertForm';

const { Title } = Typography;
const { Search } = Input;

// 证书状态映射
const statusMap: Record<string, { text: string; status: 'processing' | 'success' | 'error' | 'default' | 'warning' }> = {
  pending: { text: '待签发', status: 'processing' },
  issued: { text: '已签发', status: 'success' },
  active: { text: '正常', status: 'success' },
  expired: { text: '已过期', status: 'error' },
  revoked: { text: '已吊销', status: 'default' },
  expiring: { text: '即将过期', status: 'warning' },
  domain_verify: { text: '待验证', status: 'warning' },
  process: { text: '审核中', status: 'processing' },
  failed: { text: '失败', status: 'error' },
};

// CA 提供商颜色映射
const providerColorMap: Record<string, string> = {
  aliyun: 'orange',
  tencent: 'blue',
  volcengine: 'red',
  wangsu: 'purple',
  aws: 'gold',
  azure: 'cyan',
  gcp: 'green',
  huawei: 'geekblue',
};

// CA 提供商名称映射
const providerNameMap: Record<string, string> = {
  aliyun: '阿里云',
  tencent: '腾讯云',
  volcengine: '火山云',
  wangsu: '网宿',
  aws: 'AWS',
  azure: 'Azure',
  gcp: 'GCP',
  huawei: '华为云',
};

const CertificateListPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<Certificate[]>([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [queryParams, setQueryParams] = useState<CertificateQueryParams>({
    page: 1,
    pageSize: 10,
    status: '',
    search: '',
    sortBy: 'notAfter_desc',
  });

  const [applyModalVisible, setApplyModalVisible] = useState(false);
  const [importModalVisible, setImportModalVisible] = useState(false);
  const [syncLoading, setSyncLoading] = useState<Record<string, boolean>>({});

  // 获取证书列表
  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getCertificates(queryParams);
      setData(res.list);
      setPagination({
        current: res.page,
        pageSize: res.pageSize,
        total: res.total,
      });
    } finally {
      setLoading(false);
    }
  }, [queryParams]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // 处理搜索
  const handleSearch = (value: string) => {
    setQueryParams((prev) => ({
      ...prev,
      search: value,
      page: 1,
    }));
  };

  // 处理状态筛选
  const handleStatusChange = (value: string) => {
    setQueryParams((prev) => ({
      ...prev,
      status: value,
      page: 1,
    }));
  };

  // 处理排序
  const handleSortChange = (value: string) => {
    setQueryParams((prev) => ({
      ...prev,
      sortBy: value,
      page: 1,
    }));
  };

  // 处理分页
  const handlePageChange = (page: number, pageSize: number) => {
    setQueryParams((prev) => ({
      ...prev,
      page,
      pageSize,
    }));
  };

  // 删除证书
  const handleDelete = async (id: string) => {
    try {
      await deleteCertificate(id);
      message.success('删除成功');
      fetchData();
    } catch (error) {
      // 错误已在 request.ts 中处理
    }
  };

  // 同步证书状态
  const handleSync = async (id: string) => {
    setSyncLoading((prev) => ({ ...prev, [id]: true }));
    try {
      await syncCertStatus(id);
      message.success('同步成功');
      fetchData();
    } catch (error) {
      // 错误已在 request.ts 中处理
    } finally {
      setSyncLoading((prev) => ({ ...prev, [id]: false }));
    }
  };

  // 渲染过期时间（带过期警告）
  const renderExpireTime = (notAfter: string) => {
    const expireDate = dayjs(notAfter);
    const now = dayjs();
    const daysUntilExpire = expireDate.diff(now, 'day');

    let color: string | undefined;
    let icon = null;

    if (daysUntilExpire < 0) {
      color = '#ff4d4f'; // 已过期，红色
    } else if (daysUntilExpire <= 7) {
      color = '#ff4d4f'; // 7天内过期，红色
      icon = <WarningOutlined style={{ color: '#faad14', marginRight: 4 }} />;
    } else if (daysUntilExpire <= 30) {
      color = '#ff4d4f'; // 30天内过期，红色
    }

    return (
      <span style={{ color }}>
        {icon}
        {expireDate.format('YYYY-MM-DD HH:mm')}
        {daysUntilExpire >= 0 && (
          <span style={{ marginLeft: 8, fontSize: 12 }}>
            ({daysUntilExpire}天后过期)
          </span>
        )}
      </span>
    );
  };

  const columns: ColumnsType<Certificate> = [
    {
      title: '域名',
      dataIndex: 'domain',
      key: 'domain',
      render: (domain: string, record: Certificate) => (
        <a onClick={() => navigate(`/certificates/${record.id}`)}>{domain}</a>
      ),
    },
    {
      title: 'CA机构',
      dataIndex: 'caProvider',
      key: 'caProvider',
      render: (caProvider: string) => (
        <Tag color={providerColorMap[caProvider] || 'default'}>
          {providerNameMap[caProvider] || caProvider}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const config = statusMap[status] || { text: status, status: 'default' };
        return <Badge status={config.status} text={config.text} />;
      },
    },
    {
      title: '过期时间',
      dataIndex: 'expireAt',
      key: 'expireAt',
      render: renderExpireTime,
    },
    {
      title: '颁发者',
      dataIndex: 'issuer',
      key: 'issuer',
      ellipsis: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record: Certificate) => (
        <Space size="small">
          <Tooltip title="详情">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/certificates/${record.id}`)}
            />
          </Tooltip>
          <Tooltip title="同步状态">
            <Button
              type="text"
              icon={<SyncOutlined spin={syncLoading[record.id]} />}
              onClick={() => handleSync(record.id)}
              loading={syncLoading[record.id]}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Popconfirm
              title="确认删除"
              description="删除后无法恢复，是否继续？"
              onConfirm={() => handleDelete(record.id)}
              okText="删除"
              cancelText="取消"
              okButtonProps={{ danger: true }}
            >
              <Button type="text" danger icon={<DeleteOutlined />} />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <Card>
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        {/* 标题和操作按钮 */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Title level={2} style={{ margin: 0 }}>证书管理</Title>
          <Space>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setApplyModalVisible(true)}
            >
              申请证书
            </Button>
            <Button
              icon={<ImportOutlined />}
              onClick={() => setImportModalVisible(true)}
            >
              导入证书
            </Button>
          </Space>
        </div>

        {/* 筛选栏 */}
        <Space wrap>
          <Select
            placeholder="状态筛选"
            value={queryParams.status || ''}
            onChange={handleStatusChange}
            style={{ width: 120 }}
            options={[
              { value: '', label: '全部' },
              { value: 'pending', label: '待签发' },
              { value: 'domain_verify', label: '待验证' },
              { value: 'process', label: '审核中' },
              { value: 'issued', label: '已签发' },
              { value: 'expired', label: '已过期' },
              { value: 'failed', label: '失败' },
              { value: 'revoked', label: '已吊销' },
            ]}
          />
          <Search
            placeholder="搜索域名"
            allowClear
            onSearch={handleSearch}
            style={{ width: 250 }}
          />
          <Select
            placeholder="排序方式"
            value={queryParams.sortBy}
            onChange={handleSortChange}
            style={{ width: 150 }}
            options={[
              { value: 'notAfter_desc', label: '过期时间降序' },
              { value: 'notAfter_asc', label: '过期时间升序' },
              { value: 'createdAt_desc', label: '创建时间降序' },
              { value: 'createdAt_asc', label: '创建时间升序' },
            ]}
          />
        </Space>

        {/* 表格 */}
        <Table
          rowKey="id"
          columns={columns}
          dataSource={data}
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: handlePageChange,
          }}
        />
      </Space>

      {/* 申请证书弹窗 */}
      <ApplyCertForm
        visible={applyModalVisible}
        onCancel={() => setApplyModalVisible(false)}
        onSuccess={fetchData}
      />

      {/* 导入证书弹窗 */}
      <ImportCertForm
        visible={importModalVisible}
        onCancel={() => setImportModalVisible(false)}
        onSuccess={fetchData}
      />
    </Card>
  );
};

export default CertificateListPage;
