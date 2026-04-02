import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  Button,
  Row,
  Col,
  Typography,
  Modal,
  Form,
  Input,
  message,
  Popconfirm,
  Progress,
  Space,
  Empty,
  Spin,
} from 'antd';
import { PlusOutlined, SettingOutlined, DeleteOutlined } from '@ant-design/icons';
import { getClusters, createCluster, deleteCluster, type NginxCluster } from '@/api/nginx';

const { Title, Text } = Typography;
const { TextArea } = Input;

const NginxClusterPage: React.FC = () => {
  const navigate = useNavigate();
  const [clusters, setClusters] = useState<NginxCluster[]>([]);
  const [loading, setLoading] = useState(false);
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [createForm] = Form.useForm();
  const [createLoading, setCreateLoading] = useState(false);

  // 获取集群列表
  const fetchClusters = async () => {
    setLoading(true);
    try {
      const res = await getClusters({ page: 1, pageSize: 100 });
      setClusters(res.list);
    } catch (error) {
      message.error('获取集群列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchClusters();
  }, []);

  // 创建集群
  const handleCreate = async (values: { name: string; description?: string }) => {
    setCreateLoading(true);
    try {
      await createCluster(values);
      message.success('创建集群成功');
      setCreateModalVisible(false);
      createForm.resetFields();
      fetchClusters();
    } catch (error) {
      message.error('创建集群失败');
    } finally {
      setCreateLoading(false);
    }
  };

  // 删除集群
  const handleDelete = async (id: number) => {
    try {
      await deleteCluster(id);
      message.success('删除集群成功');
      fetchClusters();
    } catch (error) {
      message.error('删除集群失败');
    }
  };

  // 计算在线率
  const getOnlineRate = (cluster: NginxCluster) => {
    if (cluster.nodeCount === 0) return 0;
    return Math.round((cluster.onlineCount / cluster.nodeCount) * 100);
  };

  // 获取在线率颜色
  const getOnlineRateColor = (rate: number) => {
    if (rate >= 90) return '#52c41a';
    if (rate >= 60) return '#faad14';
    return '#ff4d4f';
  };

  return (
    <div>
      {/* 页面标题 */}
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={2} style={{ margin: 0 }}>Nginx 集群管理</Title>
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={() => setCreateModalVisible(true)}
        >
          创建集群
        </Button>
      </div>

      {/* 集群列表 */}
      <Spin spinning={loading}>
        {clusters.length === 0 ? (
          <Empty description="暂无集群" style={{ marginTop: 64 }} />
        ) : (
          <Row gutter={[16, 16]}>
            {clusters.map((cluster) => {
              const onlineRate = getOnlineRate(cluster);
              return (
                <Col xs={24} sm={12} lg={8} xl={6} key={cluster.id}>
                  <Card
                    hoverable
                    title={cluster.name}
                    actions={[
                      <Button
                        key="manage"
                        type="link"
                        icon={<SettingOutlined />}
                        onClick={() => navigate(`/nginx/${cluster.id}`)}
                      >
                        管理
                      </Button>,
                      <Popconfirm
                        key="delete"
                        title="确定要删除此集群吗？"
                        description="删除后无法恢复，集群下的所有节点也将被删除。"
                        onConfirm={() => handleDelete(cluster.id)}
                        okText="确定"
                        cancelText="取消"
                      >
                        <Button type="link" danger icon={<DeleteOutlined />}>
                          删除
                        </Button>
                      </Popconfirm>,
                    ]}
                  >
                    <Space direction="vertical" style={{ width: '100%' }}>
                      <Text type="secondary" ellipsis={{ tooltip: cluster.description }}>
                        {cluster.description || '暂无描述'}
                      </Text>
                      
                      <div style={{ marginTop: 8 }}>
                        <Text>节点数量: <strong>{cluster.nodeCount}</strong></Text>
                      </div>
                      
                      <div style={{ marginTop: 8 }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                          <Text>在线率</Text>
                          <Text strong style={{ color: getOnlineRateColor(onlineRate) }}>
                            {onlineRate}%
                          </Text>
                        </div>
                        <Progress
                          percent={onlineRate}
                          size="small"
                          strokeColor={getOnlineRateColor(onlineRate)}
                          showInfo={false}
                        />
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {cluster.onlineCount} / {cluster.nodeCount} 节点在线
                        </Text>
                      </div>
                    </Space>
                  </Card>
                </Col>
              );
            })}
          </Row>
        )}
      </Spin>

      {/* 创建集群 Modal */}
      <Modal
        title="创建集群"
        open={createModalVisible}
        onOk={() => createForm.submit()}
        onCancel={() => {
          setCreateModalVisible(false);
          createForm.resetFields();
        }}
        confirmLoading={createLoading}
        okText="创建"
        cancelText="取消"
      >
        <Form
          form={createForm}
          layout="vertical"
          onFinish={handleCreate}
        >
          <Form.Item
            name="name"
            label="集群名称"
            rules={[
              { required: true, message: '请输入集群名称' },
              { max: 100, message: '名称长度不能超过100个字符' },
            ]}
          >
            <Input placeholder="请输入集群名称" />
          </Form.Item>
          
          <Form.Item
            name="description"
            label="描述"
            rules={[{ max: 500, message: '描述长度不能超过500个字符' }]}
          >
            <TextArea
              rows={3}
              placeholder="请输入集群描述（可选）"
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default NginxClusterPage;
