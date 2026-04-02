import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Button,
  Descriptions,
  Table,
  Tag,
  Space,
  Progress,
  Popconfirm,
  message,
  Badge,
  Tooltip,
  Spin,
  Typography,
} from 'antd';

const { Title } = Typography;
import {
  RollbackOutlined,
} from '@ant-design/icons';
import type { DeployTask, DeployTaskItem, DeployTaskItemStatus, DeployProviderType } from '@/types';
import { getDeployTask, rollbackDeployTask, rollbackDeployTaskItem, connectDeployWS } from '@/api/deployment';

// Provider 类型映射
const providerTypeMap: Record<DeployProviderType, string> = {
  aliyun: '阿里云',
  tencent: '腾讯云',
  volcengine: '火山云',
  wangsu: '网宿',
  aws: 'AWS',
  azure: 'Azure',
};

// Provider 颜色映射
const providerColorMap: Record<DeployProviderType, string> = {
  aliyun: 'orange',
  tencent: 'blue',
  volcengine: 'green',
  wangsu: 'cyan',
  aws: 'purple',
  azure: 'geekblue',
};

// 目标类型映射
const targetTypeMap: Record<string, string> = {
  cdn: 'CDN',
  slb: 'SLB',
  dcdn: 'DCDN',
  clb: 'CLB',
  cloudfront: 'CloudFront',
  elb: 'ELB',
  appgateway: 'App Gateway',
};

// 状态映射
const statusMap: Record<DeployTaskItemStatus, { text: string; color: string }> = {
  pending: { text: '待执行', color: 'default' },
  deploying: { text: '部署中', color: 'processing' },
  success: { text: '成功', color: 'success' },
  failed: { text: '失败', color: 'error' },
  rolled_back: { text: '已回滚', color: 'warning' },
};

// 任务状态映射
const taskStatusMap: Record<string, { text: string; color: string }> = {
  pending: { text: '待执行', color: 'default' },
  deploying: { text: '部署中', color: 'processing' },
  success: { text: '成功', color: 'success' },
  partial_success: { text: '部分成功', color: 'warning' },
  failed: { text: '失败', color: 'error' },
};

const DeploymentDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [task, setTask] = useState<DeployTask | null>(null);
  const [loading, setLoading] = useState(false);
  const [rollingBack, setRollingBack] = useState(false);
  const [rollingBackItemId, setRollingBackItemId] = useState<string | null>(null);

  const fetchTask = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const res = await getDeployTask(id);
      setTask(res);
    } catch (error) {
      message.error('获取部署任务详情失败');
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchTask();
  }, [fetchTask]);

  // WebSocket 实时更新
  useEffect(() => {
    if (!id) return;

    const ws = connectDeployWS(id);

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        // 更新子任务状态
        if (data.type === 'item_update' && data.item) {
          setTask((prev) => {
            if (!prev || !prev.items) return prev;
            return {
              ...prev,
              items: prev.items.map((item) =>
                item.id === data.item.id ? { ...item, ...data.item } : item
              ),
            };
          });
        }
        // 更新整体任务状态
        if (data.type === 'task_update' && data.task) {
          setTask((prev) => (prev ? { ...prev, ...data.task } : prev));
        }
      } catch (error) {
        console.error('WebSocket message parse error:', error);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    return () => {
      ws.close();
    };
  }, [id]);

  const handleRollbackAll = async () => {
    if (!id) return;
    setRollingBack(true);
    try {
      await rollbackDeployTask(id);
      message.success('回滚任务已提交');
      fetchTask();
    } catch (error) {
      message.error('回滚部署任务失败');
    } finally {
      setRollingBack(false);
    }
  };

  const handleRollbackItem = async (itemId: string) => {
    setRollingBackItemId(itemId);
    try {
      await rollbackDeployTaskItem(itemId);
      message.success('回滚任务项已提交');
      fetchTask();
    } catch (error) {
      message.error('回滚任务项失败');
    } finally {
      setRollingBackItemId(null);
    }
  };

  const calculateProgress = () => {
    if (!task || task.totalItems === 0) return 0;
    return Math.round((task.successItems / task.totalItems) * 100);
  };

  const columns = [
    {
      title: '目标类型',
      dataIndex: 'targetType',
      key: 'targetType',
      render: (type: string) => <Tag>{targetTypeMap[type] || type}</Tag>,
    },
    {
      title: '云服务商',
      dataIndex: 'providerType',
      key: 'providerType',
      render: (type: DeployProviderType) => (
        <Tag color={providerColorMap[type]}>{providerTypeMap[type]}</Tag>
      ),
    },
    {
      title: '资源 ID',
      dataIndex: 'resourceId',
      key: 'resourceId',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: DeployTaskItemStatus) => {
        const config = statusMap[status];
        return <Badge status={config.color as 'default' | 'processing' | 'success' | 'error' | 'warning'} text={config.text} />;
      },
    },
    {
      title: '错误信息',
      dataIndex: 'errorMessage',
      key: 'errorMessage',
      render: (text: string) => {
        if (!text) return '-';
        return (
          <Tooltip title={text}>
            <span style={{ color: '#ff4d4f', cursor: 'pointer' }}>查看错误</span>
          </Tooltip>
        );
      },
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: DeployTaskItem) => (
        record.status === 'success' && (
          <Popconfirm
            title="确认回滚"
            description="确定要回滚此部署目标吗？"
            onConfirm={() => handleRollbackItem(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button
              type="text"
              icon={<RollbackOutlined />}
              loading={rollingBackItemId === record.id}
              title="回滚"
            />
          </Popconfirm>
        )
      ),
    },
  ];

  if (loading && !task) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: 50 }}>
          <Spin size="large" />
        </div>
      </Card>
    );
  }

  return (
    <Card>
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <Button onClick={() => navigate('/deployments')}>
            返回
          </Button>
          <Title level={4} style={{ margin: 0 }}>{task?.name || '部署详情'}</Title>
        </Space>
        {task && (task.status === 'success' || task.status === 'partial_success') && (
          <Popconfirm
            title="确认回滚所有已部署的目标"
            description="确定要回滚所有已部署的目标吗？此操作不可恢复。"
            onConfirm={handleRollbackAll}
            okText="确认"
            cancelText="取消"
          >
            <Button
              type="primary"
              danger
              icon={<RollbackOutlined />}
              loading={rollingBack}
            >
              一键回滚
            </Button>
          </Popconfirm>
        )}
      </div>

      {task && (
        <>
          <Descriptions title="基本信息" bordered size="small" column={2}>
            <Descriptions.Item label="任务名称">{task.name}</Descriptions.Item>
            <Descriptions.Item label="关联证书">{task.certificateDomain}</Descriptions.Item>
            <Descriptions.Item label="状态">
              {taskStatusMap[task.status] && (
                <Badge
                  status={taskStatusMap[task.status].color as 'default' | 'processing' | 'success' | 'error' | 'warning'}
                  text={taskStatusMap[task.status].text}
                />
              )}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {new Date(task.createdAt).toLocaleString()}
            </Descriptions.Item>
            <Descriptions.Item label="目标统计">
              <Space>
                <span>总数: <Tag>{task.totalItems}</Tag></span>
                <span>成功: <Tag color="success">{task.successItems}</Tag></span>
                <span>失败: <Tag color="error">{task.failedItems}</Tag></span>
              </Space>
            </Descriptions.Item>
          </Descriptions>

          <div style={{ marginTop: 24 }}>
            <h4>总体进度</h4>
            <Progress
              percent={calculateProgress()}
              status={task.status === 'failed' ? 'exception' : task.status === 'success' ? 'success' : 'active'}
              strokeColor={task.status === 'partial_success' ? '#faad14' : undefined}
            />
          </div>

          <div style={{ marginTop: 24 }}>
            <h4>子任务列表</h4>
            <Table
              rowKey="id"
              columns={columns}
              dataSource={task.items || []}
              pagination={false}
              size="small"
            />
          </div>
        </>
      )}
    </Card>
  );
};

export default DeploymentDetailPage;
