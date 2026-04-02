import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Card,
  Button,
  Typography,
  Descriptions,
  Table,
  Tag,
  Modal,
  Form,
  Input,
  message,
  Popconfirm,
  Space,
  Spin,
  Select,
  Badge,
  Empty,
} from 'antd';
import {
  ArrowLeftOutlined,
  PlusOutlined,
  DeploymentUnitOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import {
  getCluster,
  addNode,
  removeNode,
  deployToCluster,
  type NginxCluster,
  type NginxNode,
  type DeployResult,
} from '@/api/nginx';
import { getCertificates } from '@/api/certificate';
import type { Certificate } from '@/types';

const { Title, Text } = Typography;

const NginxClusterDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const clusterId = parseInt(id || '0', 10);

  const [cluster, setCluster] = useState<NginxCluster | null>(null);
  const [loading, setLoading] = useState(false);
  
  // 添加节点 Modal
  const [addNodeModalVisible, setAddNodeModalVisible] = useState(false);
  const [addNodeForm] = Form.useForm();
  const [addNodeLoading, setAddNodeLoading] = useState(false);

  // 部署证书 Modal
  const [deployModalVisible, setDeployModalVisible] = useState(false);
  const [deployForm] = Form.useForm();
  const [deployLoading, setDeployLoading] = useState(false);
  const [certificates, setCertificates] = useState<Certificate[]>([]);
  const [certLoading, setCertLoading] = useState(false);

  // 部署结果 Modal
  const [deployResultModalVisible, setDeployResultModalVisible] = useState(false);
  const [deployResults, setDeployResults] = useState<DeployResult[]>([]);

  // 获取集群详情
  const fetchCluster = async () => {
    if (!clusterId) return;
    setLoading(true);
    try {
      const res = await getCluster(clusterId);
      setCluster(res);
    } catch (error) {
      message.error('获取集群详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchCluster();
  }, [clusterId]);

  // 获取证书列表
  const fetchCertificates = async () => {
    setCertLoading(true);
    try {
      const res = await getCertificates({ page: 1, pageSize: 100, status: 'active' });
      setCertificates(res.list);
    } catch (error) {
      message.error('获取证书列表失败');
    } finally {
      setCertLoading(false);
    }
  };

  // 添加节点
  const handleAddNode = async (values: { ip: string; port: string }) => {
    setAddNodeLoading(true);
    try {
      await addNode(clusterId, values);
      message.success('添加节点成功');
      setAddNodeModalVisible(false);
      addNodeForm.resetFields();
      fetchCluster();
    } catch (error) {
      message.error('添加节点失败');
    } finally {
      setAddNodeLoading(false);
    }
  };

  // 移除节点
  const handleRemoveNode = async (nodeId: number) => {
    try {
      await removeNode(nodeId);
      message.success('移除节点成功');
      fetchCluster();
    } catch (error) {
      message.error('移除节点失败');
    }
  };

  // 打开部署证书 Modal
  const handleOpenDeployModal = () => {
    fetchCertificates();
    setDeployModalVisible(true);
  };

  // 部署证书
  const handleDeploy = async (values: { certificateId: number }) => {
    setDeployLoading(true);
    try {
      const results = await deployToCluster(clusterId, { certificateId: values.certificateId });
      setDeployResults(results);
      setDeployModalVisible(false);
      deployForm.resetFields();
      setDeployResultModalVisible(true);
      
      // 检查是否有失败
      const hasFailure = results.some(r => !r.success);
      if (hasFailure) {
        message.warning('部分节点部署失败，请查看详情');
      } else {
        message.success('证书部署成功');
      }
    } catch (error) {
      message.error('部署证书失败');
    } finally {
      setDeployLoading(false);
    }
  };

  // 获取状态标签
  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'online':
        return <Badge status="success" text="在线" />;
      case 'offline':
        return <Badge status="error" text="离线" />;
      case 'busy':
        return <Badge status="processing" text="忙碌" />;
      case 'error':
        return <Badge status="warning" text="错误" />;
      default:
        return <Badge status="default" text={status} />;
    }
  };

  // 节点表格列
  const nodeColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: 'IP 地址',
      dataIndex: 'ip',
      key: 'ip',
    },
    {
      title: '端口',
      dataIndex: 'port',
      key: 'port',
      width: 100,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => getStatusBadge(status),
    },
    {
      title: '最后心跳时间',
      dataIndex: 'lastHeartbeat',
      key: 'lastHeartbeat',
      render: (time: string) => time || '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: NginxNode) => (
        <Popconfirm
          title="确定要移除此节点吗？"
          onConfirm={() => handleRemoveNode(record.id)}
          okText="确定"
          cancelText="取消"
        >
          <Button type="link" danger icon={<DeleteOutlined />}>
            移除
          </Button>
        </Popconfirm>
      ),
    },
  ];

  if (!cluster && !loading) {
    return (
      <Card>
        <Empty description="集群不存在" />
      </Card>
    );
  }

  return (
    <div>
      {/* 页面头部 */}
      <Card style={{ marginBottom: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/nginx')}>
              返回
            </Button>
            <Title level={3} style={{ margin: 0 }}>
              {cluster?.name || '集群详情'}
            </Title>
          </Space>
          <Space>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setAddNodeModalVisible(true)}
            >
              添加节点
            </Button>
            <Button
              type="primary"
              icon={<DeploymentUnitOutlined />}
              onClick={handleOpenDeployModal}
              disabled={!cluster?.nodes?.length}
            >
              部署证书
            </Button>
          </Space>
        </div>
      </Card>

      <Spin spinning={loading}>
        {/* 基本信息 */}
        <Card title="基本信息" style={{ marginBottom: 16 }}>
          <Descriptions column={2}>
            <Descriptions.Item label="集群名称">{cluster?.name}</Descriptions.Item>
            <Descriptions.Item label="节点数量">
              <Tag color="blue">{cluster?.nodeCount || 0}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="描述">{cluster?.description || '-'}</Descriptions.Item>
            <Descriptions.Item label="在线节点">
              <Tag color="green">{cluster?.onlineCount || 0}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">{cluster?.createdAt}</Descriptions.Item>
            <Descriptions.Item label="更新时间">{cluster?.updatedAt}</Descriptions.Item>
          </Descriptions>
        </Card>

        {/* 节点列表 */}
        <Card title="节点列表">
          <Table
            dataSource={cluster?.nodes || []}
            columns={nodeColumns}
            rowKey="id"
            pagination={false}
            locale={{ emptyText: '暂无节点，请点击"添加节点"按钮添加' }}
          />
        </Card>
      </Spin>

      {/* 添加节点 Modal */}
      <Modal
        title="添加节点"
        open={addNodeModalVisible}
        onOk={() => addNodeForm.submit()}
        onCancel={() => {
          setAddNodeModalVisible(false);
          addNodeForm.resetFields();
        }}
        confirmLoading={addNodeLoading}
        okText="添加"
        cancelText="取消"
      >
        <Form
          form={addNodeForm}
          layout="vertical"
          onFinish={handleAddNode}
        >
          <Form.Item
            name="ip"
            label="IP 地址"
            rules={[
              { required: true, message: '请输入 IP 地址' },
              { pattern: /^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$/, message: '请输入有效的 IP 地址' },
            ]}
          >
            <Input placeholder="例如: 192.168.1.1" />
          </Form.Item>
          
          <Form.Item
            name="port"
            label="Agent 端口"
            rules={[
              { required: true, message: '请输入端口' },
              { pattern: /^\d+$/, message: '端口必须是数字' },
            ]}
            initialValue="9090"
          >
            <Input placeholder="例如: 9090" />
          </Form.Item>
        </Form>
      </Modal>

      {/* 部署证书 Modal */}
      <Modal
        title="部署证书"
        open={deployModalVisible}
        onOk={() => deployForm.submit()}
        onCancel={() => {
          setDeployModalVisible(false);
          deployForm.resetFields();
        }}
        confirmLoading={deployLoading}
        okText="部署"
        cancelText="取消"
      >
        <Form
          form={deployForm}
          layout="vertical"
          onFinish={handleDeploy}
        >
          <Form.Item
            name="certificateId"
            label="选择证书"
            rules={[{ required: true, message: '请选择要部署的证书' }]}
          >
            <Select
              placeholder="请选择证书"
              loading={certLoading}
              showSearch
              optionFilterProp="children"
            >
              {certificates.map((cert) => (
                <Select.Option key={cert.id} value={cert.id}>
                  {cert.domain} ({cert.issuer})
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          
          <Text type="secondary">
            将证书部署到集群的所有节点，采用滚动部署策略，一个节点失败则停止部署。
          </Text>
        </Form>
      </Modal>

      {/* 部署结果 Modal */}
      <Modal
        title="部署结果"
        open={deployResultModalVisible}
        onOk={() => setDeployResultModalVisible(false)}
        onCancel={() => setDeployResultModalVisible(false)}
        footer={[
          <Button key="ok" type="primary" onClick={() => setDeployResultModalVisible(false)}>
            确定
          </Button>,
        ]}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          {deployResults.map((result) => (
            <Card
              key={result.nodeId}
              size="small"
              style={{
                borderLeft: `4px solid ${result.success ? '#52c41a' : '#ff4d4f'}`,
              }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <Text strong>{result.ip}:{result.port}</Text>
                  <br />
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {result.message}
                  </Text>
                </div>
                <Tag color={result.success ? 'success' : 'error'}>
                  {result.success ? '成功' : '失败'}
                </Tag>
              </div>
            </Card>
          ))}
        </Space>
      </Modal>
    </div>
  );
};

export default NginxClusterDetailPage;
