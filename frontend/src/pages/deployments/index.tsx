import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Button,
  Table,
  Tag,
  Space,
  Select,
  Popconfirm,
  message,
  Badge,
  Spin,
  Typography,
} from 'antd';
import {
  PlusOutlined,
  EyeOutlined,
  PlayCircleOutlined,
  RollbackOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import type { DeployTask, DeployTaskStatus } from '@/types';
import { getDeployTasks, executeDeployTask, rollbackDeployTask, deleteDeployTask } from '@/api/deployment';
import CreateDeployForm from './components/CreateDeployForm';

const { Title } = Typography;
const { Option } = Select;

// 状态映射
const statusMap: Record<DeployTaskStatus, { text: string; color: string }> = {
  pending: { text: '待执行', color: 'processing' },
  deploying: { text: '部署中', color: 'processing' },
  success: { text: '成功', color: 'success' },
  partial_success: { text: '部分成功', color: 'warning' },
  failed: { text: '失败', color: 'error' },
};

const DeploymentListPage: React.FC = () => {
  const navigate = useNavigate();
  const [tasks, setTasks] = useState<DeployTask[]>([]);
  const [loading, setLoading] = useState(false);
  const [status, setStatus] = useState<DeployTaskStatus | undefined>();
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [executingId, setExecutingId] = useState<string | null>(null);
  const [rollingBackId, setRollingBackId] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const fetchTasks = useCallback(async () => {
    setLoading(true);
    try {
      const res = await getDeployTasks({
        page: pagination.current,
        pageSize: pagination.pageSize,
        status,
      });
      setTasks(res.list);
      setPagination((prev) => ({ ...prev, total: res.total }));
    } catch (error) {
      message.error('获取部署任务列表失败');
    } finally {
      setLoading(false);
    }
  }, [pagination.current, pagination.pageSize, status]);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const handleTableChange = (newPagination: { current?: number; pageSize?: number }) => {
    setPagination((prev) => ({
      ...prev,
      current: newPagination.current || 1,
      pageSize: newPagination.pageSize || 10,
    }));
  };

  const handleStatusChange = (value: DeployTaskStatus | undefined) => {
    setStatus(value);
    setPagination((prev) => ({ ...prev, current: 1 }));
  };

  const handleExecute = async (id: string) => {
    setExecutingId(id);
    try {
      await executeDeployTask(id);
      message.success('部署任务已开始执行');
      fetchTasks();
    } catch (error) {
      message.error('执行部署任务失败');
    } finally {
      setExecutingId(null);
    }
  };

  const handleRollback = async (id: string) => {
    setRollingBackId(id);
    try {
      await rollbackDeployTask(id);
      message.success('回滚任务已提交');
      fetchTasks();
    } catch (error) {
      message.error('回滚部署任务失败');
    } finally {
      setRollingBackId(null);
    }
  };

  const handleDelete = async (id: string) => {
    setDeletingId(id);
    try {
      await deleteDeployTask(id);
      message.success('删除部署任务成功');
      fetchTasks();
    } catch (error) {
      message.error('删除部署任务失败');
    } finally {
      setDeletingId(null);
    }
  };

  const columns = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: DeployTask) => (
        <Button
          type="link"
          onClick={() => navigate(`/deployments/${record.id}`)}
          style={{ padding: 0 }}
        >
          {text}
        </Button>
      ),
    },
    {
      title: '关联证书',
      dataIndex: 'certificateDomain',
      key: 'certificateDomain',
    },
    {
      title: '目标数量',
      dataIndex: 'totalItems',
      key: 'totalItems',
      render: (total: number) => <Tag>{total}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: DeployTaskStatus) => {
        const config = statusMap[status];
        if (status === 'deploying') {
          return (
            <Badge status="processing" text={<Spin size="small" style={{ marginLeft: 8 }} />}>
              <span style={{ marginLeft: 24 }}>{config.text}</span>
            </Badge>
          );
        }
        return <Badge status={config.color as 'processing' | 'success' | 'warning' | 'error'} text={config.text} />;
      },
    },
    {
      title: '成功/失败',
      key: 'result',
      render: (_: unknown, record: DeployTask) => (
        <Space>
          <Tag color="success">{record.successItems}</Tag>
          <span>/</span>
          <Tag color="error">{record.failedItems}</Tag>
        </Space>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (text: string) => new Date(text).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: unknown, record: DeployTask) => (
        <Space size="small">
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => navigate(`/deployments/${record.id}`)}
            title="详情"
          />
          {record.status === 'pending' && (
            <Button
              type="text"
              icon={<PlayCircleOutlined />}
              onClick={() => handleExecute(record.id)}
              loading={executingId === record.id}
              title="执行"
            />
          )}
          {(record.status === 'success' || record.status === 'partial_success') && (
            <Popconfirm
              title="确认回滚"
              description="确定要回滚此部署任务吗？"
              onConfirm={() => handleRollback(record.id)}
              okText="确认"
              cancelText="取消"
            >
              <Button
                type="text"
                icon={<RollbackOutlined />}
                loading={rollingBackId === record.id}
                title="回滚"
              />
            </Popconfirm>
          )}
          <Popconfirm
            title="确认删除"
            description="确定要删除此部署任务吗？此操作不可恢复。"
            onConfirm={() => handleDelete(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              loading={deletingId === record.id}
              title="删除"
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={2} style={{ margin: 0 }}>部署任务</Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setCreateModalVisible(true)}
        >
          创建部署任务
        </Button>
      </div>

      <div style={{ marginBottom: 16 }}>
        <Space>
          <span>状态筛选：</span>
          <Select
            style={{ width: 150 }}
            placeholder="全部状态"
            allowClear
            value={status}
            onChange={handleStatusChange}
          >
            <Option value="pending">待执行</Option>
            <Option value="deploying">部署中</Option>
            <Option value="success">成功</Option>
            <Option value="partial_success">部分成功</Option>
            <Option value="failed">失败</Option>
          </Select>
        </Space>
      </div>

      <Table
        rowKey="id"
        columns={columns}
        dataSource={tasks}
        loading={loading}
        pagination={{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        onChange={handleTableChange}
      />

      <CreateDeployForm
        visible={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onSuccess={() => {
          setCreateModalVisible(false);
          fetchTasks();
        }}
      />
    </Card>
  );
};

export default DeploymentListPage;
