import { useEffect, useState } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Spin,
  Empty,
  Table,
  Badge,
  Alert,
  Typography,
} from 'antd';
import {
  SafetyCertificateOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  BarChart,
  Bar,
} from 'recharts';
import { getDashboardOverview } from '@/api/dashboard';
import type {
  DashboardFullOverview,
} from '@/types';
import dayjs from 'dayjs';

const { Title } = Typography;

// 证书状态饼图颜色
const CERT_STATUS_COLORS = {
  issued: '#52c41a',
  pending: '#1890ff',
  expired: '#ff4d4f',
  revoked: '#d9d9d9',
};

// 告警级别对应颜色
const ALERT_LEVEL_COLORS = {
  error: 'red',
  warning: 'orange',
  info: 'blue',
};

const DashboardPage: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [data, setData] = useState<DashboardFullOverview | null>(null);

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      const overview = await getDashboardOverview();
      setData(overview);
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '400px' }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  if (!data) {
    return <Empty description="暂无数据" />;
  }

  const { certOverview, deployOverview, cloudDistribution, expiryTrend, recentTasks, alerts } = data;

  // 证书状态饼图数据
  const certPieData = [
    { name: '已颁发', value: certOverview.issued, key: 'issued' },
    { name: '待处理', value: certOverview.pending, key: 'pending' },
    { name: '已过期', value: certOverview.expired, key: 'expired' },
    { name: '已吊销', value: certOverview.revoked, key: 'revoked' },
  ].filter(item => item.value > 0);

  // 统计卡片数据
  const statCards = [
    {
      title: '证书总数',
      value: certOverview.total,
      icon: <SafetyCertificateOutlined style={{ fontSize: 32, color: '#1890ff' }} />,
      color: '#1890ff',
    },
    {
      title: '即将过期',
      value: certOverview.expiring30,
      icon: <ClockCircleOutlined style={{ fontSize: 32, color: '#fa8c16' }} />,
      color: '#fa8c16',
      suffix: '个(30天内)',
    },
    {
      title: '部署成功率',
      value: deployOverview.successRate?.toFixed(1) || 0,
      icon: <CheckCircleOutlined style={{ fontSize: 32, color: '#52c41a' }} />,
      color: '#52c41a',
      suffix: '%',
    },
    {
      title: '异常告警',
      value: alerts.length,
      icon: <ExclamationCircleOutlined style={{ fontSize: 32, color: '#ff4d4f' }} />,
      color: '#ff4d4f',
    },
  ];

  // 最近部署任务表格列
  const taskColumns = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
      ellipsis: true,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusMap: Record<string, { color: string; text: string }> = {
          pending: { color: 'default', text: '待处理' },
          deploying: { color: 'processing', text: '部署中' },
          success: { color: 'success', text: '成功' },
          partial_success: { color: 'warning', text: '部分成功' },
          failed: { color: 'error', text: '失败' },
        };
        const { color, text } = statusMap[status] || { color: 'default', text: status };
        return <Badge status={color as any} text={text} />;
      },
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (time: string) => dayjs(time).format('MM-DD HH:mm'),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2} style={{ marginBottom: 24 }}>大盘总览</Title>

      {/* 第一行：统计卡片 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        {statCards.map((card, index) => (
          <Col xs={24} sm={12} md={12} lg={6} xl={6} key={index}>
            <Card hoverable>
              <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                {card.icon}
                <Statistic
                  title={card.title}
                  value={card.value}
                  suffix={card.suffix}
                  valueStyle={{ color: card.color, fontWeight: 'bold' }}
                />
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      {/* 第二行：证书状态饼图 + 到期趋势折线图 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={24} md={24} lg={12} xl={12}>
          <Card title="证书状态分布" hoverable>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={certPieData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name}: ${((percent || 0) * 100).toFixed(0)}%`}
                  outerRadius={100}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {certPieData.map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={CERT_STATUS_COLORS[entry.key as keyof typeof CERT_STATUS_COLORS]}
                    />
                  ))}
                </Pie>
                <Tooltip />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col xs={24} sm={24} md={24} lg={12} xl={12}>
          <Card title="证书到期趋势 (未来90天)" hoverable>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={expiryTrend}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis
                  dataKey="date"
                  tickFormatter={(value) => dayjs(value).format('MM-DD')}
                  interval="preserveStartEnd"
                />
                <YAxis />
                <Tooltip
                  labelFormatter={(value) => dayjs(value).format('YYYY-MM-DD')}
                  formatter={(value) => [`${value} 个`, '到期证书']}
                />
                <Line
                  type="monotone"
                  dataKey="count"
                  stroke="#1890ff"
                  strokeWidth={2}
                  dot={false}
                  activeDot={{ r: 6 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </Card>
        </Col>
      </Row>

      {/* 第三行：云资源分布柱状图 + 最近部署任务 */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={24} md={24} lg={12} xl={12}>
          <Card title="云资源分布" hoverable>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={cloudDistribution}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="provider" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="certCount" name="证书数" fill="#1890ff" />
                <Bar dataKey="deployCount" name="部署数" fill="#52c41a" />
              </BarChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col xs={24} sm={24} md={24} lg={12} xl={12}>
          <Card title="最近部署任务" hoverable>
            <Table
              dataSource={recentTasks}
              columns={taskColumns}
              rowKey="id"
              size="small"
              pagination={false}
              scroll={{ y: 240 }}
            />
          </Card>
        </Col>
      </Row>

      {/* 第四行：异常告警列表 */}
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card title="异常告警" hoverable>
            {alerts.length === 0 ? (
              <Empty description="暂无告警" />
            ) : (
              <div style={{ maxHeight: 300, overflow: 'auto' }}>
                {alerts.map((alert, index) => (
                  <Alert
                    key={index}
                    message={alert.title}
                    description={alert.detail}
                    type={ALERT_LEVEL_COLORS[alert.level] as 'error' | 'warning' | 'info'}
                    showIcon
                    style={{ marginBottom: 8 }}
                  />
                ))}
              </div>
            )}
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default DashboardPage;
