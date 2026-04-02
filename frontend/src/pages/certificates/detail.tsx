import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Typography,
  Button,
  Descriptions,
  Badge,
  Tag,
  Timeline,
  Table,
  Collapse,
  Space,
  message,
  Spin,
  Alert,
} from 'antd';
import {
  ArrowLeftOutlined,
  CopyOutlined,
  SafetyCertificateOutlined,
  CheckCircleOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import dayjs from 'dayjs';
import type { Certificate, DeployTask } from '@/types';
import { getCertificate, syncCertStatus } from '@/api/certificate';

const { Title, Text, Paragraph } = Typography;
const { Panel } = Collapse;

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

const CertificateDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [certificate, setCertificate] = useState<Certificate | null>(null);
  const [syncLoading, setSyncLoading] = useState(false);

  // 获取证书详情
  const fetchCertificate = async () => {
    if (!id) return;
    setLoading(true);
    try {
      const data = await getCertificate(id);
      setCertificate(data);
    } catch (error) {
      // 错误已在 request.ts 中处理
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCertificate();
  }, [id]);

  // 复制到剪贴板
  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      message.success('已复制到剪贴板');
    });
  };

  // 同步证书状态
  const handleSync = async () => {
    if (!id) return;
    setSyncLoading(true);
    try {
      const data = await syncCertStatus(id);
      setCertificate(data);
      message.success('状态同步成功');
    } catch (error) {
      // 错误已在 request.ts 中处理
    } finally {
      setSyncLoading(false);
    }
  };

  // 解析验证信息 JSON
  const parseVerifyInfo = (verifyInfo?: string): Record<string, string> => {
    if (!verifyInfo) return {};
    try {
      return JSON.parse(verifyInfo);
    } catch {
      return {};
    }
  };

  // 渲染域名验证指引
  const renderVerifyGuide = () => {
    if (!certificate) return null;

    if (certificate.status === 'domain_verify') {
      const verifyInfo = parseVerifyInfo(certificate.verifyInfo);

      if (certificate.verifyType === 'DNS') {
        return (
          <Alert
            type="warning"
            message="域名验证"
            description={
              <div>
                <p style={{ marginBottom: 16 }}>请在您的域名DNS管理后台添加以下记录，完成域名验证：</p>
                <Descriptions bordered column={1} size="small">
                  <Descriptions.Item label="主机记录">
                    <Text copyable>{verifyInfo.recordDomain || '-'}</Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="记录类型">
                    <Text>{verifyInfo.recordType || '-'}</Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="记录值">
                    <Text copyable>{verifyInfo.recordValue || '-'}</Text>
                  </Descriptions.Item>
                </Descriptions>
                <div style={{ marginTop: 16 }}>
                  <Button type="primary" onClick={handleSync} loading={syncLoading}>
                    我已完成验证，检查状态
                  </Button>
                </div>
              </div>
            }
            showIcon
            style={{ marginBottom: 24 }}
          />
        );
      }

      if (certificate.verifyType === 'FILE') {
        return (
          <Alert
            type="warning"
            message="域名验证"
            description={
              <div>
                <p style={{ marginBottom: 16 }}>请在您的Web服务器根目录创建以下文件：</p>
                <Descriptions bordered column={1} size="small">
                  <Descriptions.Item label="文件路径">
                    <Text copyable>{verifyInfo.uri || '-'}</Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="文件内容">
                    <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all', background: '#f5f5f5', padding: 8, borderRadius: 4 }}>
                      {verifyInfo.content || '-'}
                    </pre>
                    <Text copyable={{ text: verifyInfo.content || '' }} style={{ marginTop: 8, display: 'inline-block' }}>
                      复制内容
                    </Text>
                  </Descriptions.Item>
                </Descriptions>
                <div style={{ marginTop: 16 }}>
                  <Button type="primary" onClick={handleSync} loading={syncLoading}>
                    我已完成验证，检查状态
                  </Button>
                </div>
              </div>
            }
            showIcon
            style={{ marginBottom: 24 }}
          />
        );
      }
    }

    if (certificate.status === 'process') {
      return (
        <Alert
          type="info"
          message="CA机构审核中"
          description={
            <div>
              <p>CA机构审核中，请耐心等待（通常1-2个工作日）</p>
              <Button onClick={handleSync} loading={syncLoading} style={{ marginTop: 8 }}>
                刷新状态
              </Button>
            </div>
          }
          showIcon
          style={{ marginBottom: 24 }}
        />
      );
    }

    if (certificate.status === 'failed') {
      return (
        <Alert
          type="error"
          message="证书申请失败"
          description="证书申请失败，请删除后重新申请"
          showIcon
          style={{ marginBottom: 24 }}
        />
      );
    }

    return null;
  };

  if (loading) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px 0' }}>
          <Spin size="large" />
        </div>
      </Card>
    );
  }

  if (!certificate) {
    return (
      <Card>
        <div style={{ textAlign: 'center', padding: '50px 0' }}>
          <Text type="secondary">证书不存在或已被删除</Text>
          <div style={{ marginTop: 16 }}>
            <Button onClick={() => navigate('/certificates')}>
              <ArrowLeftOutlined /> 返回列表
            </Button>
          </div>
        </div>
      </Card>
    );
  }

  const statusConfig = statusMap[certificate.status] || { text: certificate.status, status: 'default' };

  // 模拟证书链数据（实际应从后端获取）
  const certChain = [
    {
      title: '根 CA 证书',
      issuer: certificate.issuer,
      valid: true,
    },
    {
      title: '中间 CA 证书',
      issuer: 'Intermediate CA',
      valid: true,
    },
    {
      title: '终端证书',
      issuer: certificate.domain,
      valid: certificate.status !== 'expired' && certificate.status !== 'revoked',
    },
  ];

  // 模拟部署任务数据
  const deployTasks: DeployTask[] = [];

  const deployColumns = [
    {
      title: '任务名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '目标类型',
      dataIndex: 'targetType',
      key: 'targetType',
    },
    {
      title: '目标名称',
      dataIndex: 'targetName',
      key: 'targetName',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusConfig: Record<string, { color: string; text: string }> = {
          pending: { color: 'default', text: '待执行' },
          running: { color: 'processing', text: '执行中' },
          success: { color: 'success', text: '成功' },
          failed: { color: 'error', text: '失败' },
          partial: { color: 'warning', text: '部分成功' },
        };
        const config = statusConfig[status] || { color: 'default', text: status };
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: '执行时间',
      dataIndex: 'executedAt',
      key: 'executedAt',
      render: (time: string) => time ? dayjs(time).format('YYYY-MM-DD HH:mm') : '-',
    },
  ];

  return (
    <Card>
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        {/* 页面头部 */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Button onClick={() => navigate('/certificates')}>
            <ArrowLeftOutlined /> 返回列表
          </Button>
          <Title level={3} style={{ margin: 0 }}>
            <SafetyCertificateOutlined style={{ marginRight: 8 }} />
            证书详情
          </Title>
          <div style={{ width: 80 }} /> {/* 占位保持居中 */}
        </div>

        {/* 域名验证指引 */}
        {renderVerifyGuide()}

        {/* 基本信息 */}
        <Descriptions
          title="基本信息"
          bordered
          column={{ xxl: 3, xl: 3, lg: 2, md: 2, sm: 1, xs: 1 }}
        >
          <Descriptions.Item label="域名">{certificate.domain}</Descriptions.Item>
          <Descriptions.Item label="CA机构">
            <Tag color={providerColorMap[certificate.caProvider] || 'default'}>
              {providerNameMap[certificate.caProvider] || certificate.caProvider}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            <Badge status={statusConfig.status} text={statusConfig.text} />
          </Descriptions.Item>
          <Descriptions.Item label="颁发者" span={2}>
            {certificate.issuer}
          </Descriptions.Item>
          <Descriptions.Item label="有效期">
            {certificate.expireAt ? dayjs(certificate.expireAt).format('YYYY-MM-DD') : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="指纹" span={2}>
            <Space>
              <Text copyable={{ text: certificate.fingerprint }}>
                {certificate.fingerprint?.slice(0, 40)}...
              </Text>
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="密钥算法">
            {certificate.keyAlgorithm || '-'}
          </Descriptions.Item>
          <Descriptions.Item label="序列号" span={2}>
            {certificate.serialNumber}
          </Descriptions.Item>
        </Descriptions>

        {/* 证书链可视化 */}
        <div>
          <Title level={5}>证书链</Title>
          <Timeline
            items={certChain.map((cert) => ({
              dot: cert.valid ? (
                <CheckCircleOutlined style={{ fontSize: 16, color: '#52c41a' }} />
              ) : (
                <InfoCircleOutlined style={{ fontSize: 16, color: '#faad14' }} />
              ),
              children: (
                <div>
                  <Text strong>{cert.title}</Text>
                  <div>
                    <Text type="secondary">{cert.issuer}</Text>
                  </div>
                </div>
              ),
            }))}
          />
        </div>

        {/* 关联域名 */}
        <div>
          <Title level={5}>关联域名 (SAN)</Title>
          <Space wrap>
            {certificate.domain ? [certificate.domain].map((d) => (
              <Tag key={d} color="blue">{d}</Tag>
            )) : null}
            {!certificate.domain && (
              <Text type="secondary">无 SAN 信息</Text>
            )}
          </Space>
        </div>

        {/* 关联部署任务 */}
        <div>
          <Title level={5}>关联部署任务</Title>
          {deployTasks.length > 0 ? (
            <Table
              rowKey="id"
              columns={deployColumns}
              dataSource={deployTasks}
              pagination={false}
              size="small"
            />
          ) : (
            <Text type="secondary">暂无部署记录</Text>
          )}
        </div>

        {/* PEM 内容 */}
        <div>
          <Title level={5}>PEM 内容</Title>
          <Collapse defaultActiveKey={[]}>
            <Panel
              header="证书 PEM"
              key="1"
              extra={
                <Button
                  type="text"
                  size="small"
                  icon={<CopyOutlined />}
                  onClick={(e) => {
                    e.stopPropagation();
                    handleCopy(certificate.fingerprint);
                  }}
                >
                  复制
                </Button>
              }
            >
              <Paragraph copyable={{ text: certificate.fingerprint }}>
                <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                  {certificate.fingerprint}
                </pre>
              </Paragraph>
            </Panel>
            <Panel header="证书链 PEM" key="2">
              <Text type="secondary">暂无证书链信息</Text>
            </Panel>
          </Collapse>
        </div>
      </Space>
    </Card>
  );
};

export default CertificateDetailPage;
