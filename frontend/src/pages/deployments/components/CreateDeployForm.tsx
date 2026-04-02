import { useState, useEffect } from 'react';
import {
  Modal,
  Steps,
  Button,
  Table,
  Radio,
  Form,
  Select,
  Input,
  Space,
  Tag,
  message,
  Card,
  Descriptions,
  Spin,
} from 'antd';
import { PlusOutlined, MinusCircleOutlined } from '@ant-design/icons';
import type { Certificate, CloudCredential, DeployProviderType, DeployTargetType, DeployTarget } from '@/types';
import { getCertificates } from '@/api/certificate';
import { getCredentials } from '@/api/credential';
import { createDeployTask } from '@/api/deployment';

const { Step } = Steps;
const { Option } = Select;

interface CreateDeployFormProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

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

// 目标类型选项（根据 Provider 联动）
const targetTypeOptions: Record<DeployProviderType, { value: DeployTargetType; label: string }[]> = {
  aliyun: [
    { value: 'cdn', label: 'CDN' },
    { value: 'slb', label: 'SLB' },
    { value: 'dcdn', label: 'DCDN' },
  ],
  tencent: [
    { value: 'cdn', label: 'CDN' },
    { value: 'clb', label: 'CLB' },
  ],
  volcengine: [
    { value: 'cdn', label: 'CDN' },
    { value: 'clb', label: 'CLB' },
  ],
  wangsu: [
    { value: 'cdn', label: 'CDN' },
  ],
  aws: [
    { value: 'cloudfront', label: 'CloudFront' },
    { value: 'elb', label: 'ELB' },
  ],
  azure: [
    { value: 'cdn', label: 'CDN' },
    { value: 'appgateway', label: 'App Gateway' },
  ],
};

const CreateDeployForm: React.FC<CreateDeployFormProps> = ({ visible, onCancel, onSuccess }) => {
  const [currentStep, setCurrentStep] = useState(0);
  const [form] = Form.useForm();
  const [certificates, setCertificates] = useState<Certificate[]>([]);
  const [credentials, setCredentials] = useState<CloudCredential[]>([]);
  const [selectedCert, setSelectedCert] = useState<Certificate | null>(null);
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  // 获取证书列表（issued 状态）
  useEffect(() => {
    if (visible && currentStep === 0) {
      fetchCertificates();
    }
  }, [visible, currentStep]);

  // 获取凭证列表
  useEffect(() => {
    if (visible && currentStep === 1) {
      fetchCredentials();
    }
  }, [visible, currentStep]);

  const fetchCertificates = async () => {
    setLoading(true);
    try {
      const res = await getCertificates({ page: 1, pageSize: 100, status: 'issued' });
      setCertificates(res.list);
    } catch (error) {
      message.error('获取证书列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchCredentials = async () => {
    try {
      const res = await getCredentials({ page: 1, pageSize: 100 });
      setCredentials(res.list);
    } catch (error) {
      message.error('获取凭证列表失败');
    }
  };

  const handleNext = async () => {
    if (currentStep === 0) {
      if (!selectedCert) {
        message.warning('请选择证书');
        return;
      }
    }
    if (currentStep === 1) {
      try {
        const values = await form.validateFields();
        if (!values.targets || values.targets.length === 0) {
          message.warning('请至少添加一个部署目标');
          return;
        }
      } catch {
        return;
      }
    }
    setCurrentStep(currentStep + 1);
  };

  const handlePrev = () => {
    setCurrentStep(currentStep - 1);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);

      const targets: DeployTarget[] = values.targets.map((target: {
        providerType: DeployProviderType;
        targetType: DeployTargetType;
        resourceId: string;
        credentialId: string;
      }) => ({
        providerType: target.providerType,
        targetType: target.targetType,
        resourceId: target.resourceId,
        credentialId: target.credentialId,
      }));

      await createDeployTask({
        name: values.name,
        certificateId: selectedCert!.id,
        targets,
      });

      message.success('创建部署任务成功');
      onSuccess();
      handleReset();
    } catch (error) {
      message.error('创建部署任务失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleReset = () => {
    setCurrentStep(0);
    setSelectedCert(null);
    form.resetFields();
  };

  const handleCancel = () => {
    handleReset();
    onCancel();
  };

  // 证书列表列定义
  const certColumns = [
    {
      title: '选择',
      key: 'select',
      width: 60,
      render: (_: unknown, record: Certificate) => (
        <Radio
          checked={selectedCert?.id === record.id}
          onChange={() => setSelectedCert(record)}
        />
      ),
    },
    {
      title: '域名',
      dataIndex: 'domain',
      key: 'domain',
    },
    {
      title: '颁发者',
      dataIndex: 'issuer',
      key: 'issuer',
    },
    {
      title: '过期时间',
      dataIndex: 'expireAt',
      key: 'expireAt',
      render: (text: string) => new Date(text).toLocaleString(),
    },
  ];

  // 获取指定 provider 的凭证列表
  const getCredentialsByProvider = (provider: DeployProviderType) => {
    return credentials.filter((c) => c.providerType === provider);
  };

  // 步骤内容渲染
  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <div>
            <p style={{ marginBottom: 16 }}>请选择要部署的证书：</p>
            <Spin spinning={loading}>
              <Table
                rowKey="id"
                columns={certColumns}
                dataSource={certificates}
                pagination={false}
                size="small"
                scroll={{ y: 300 }}
              />
            </Spin>
          </div>
        );
      case 1:
        return (
          <Form form={form} layout="vertical">
            <Form.List name="targets">
              {(fields, { add, remove }) => (
                <>
                  {fields.map(({ key, name, ...restField }) => (
                    <Card
                      key={key}
                      size="small"
                      style={{ marginBottom: 16 }}
                      title={`目标 ${name + 1}`}
                      extra={
                        fields.length > 1 && (
                          <MinusCircleOutlined
                            onClick={() => remove(name)}
                            style={{ color: '#ff4d4f', cursor: 'pointer' }}
                          />
                        )
                      }
                    >
                      <Space align="start" wrap>
                        <Form.Item
                          {...restField}
                          name={[name, 'providerType']}
                          label="云服务商"
                          rules={[{ required: true, message: '请选择' }]}
                        >
                          <Select
                            style={{ width: 120 }}
                            placeholder="选择服务商"
                            onChange={() => {
                              form.setFieldValue(['targets', name, 'targetType'], undefined);
                              form.setFieldValue(['targets', name, 'credentialId'], undefined);
                            }}
                          >
                            {(Object.keys(providerTypeMap) as DeployProviderType[]).map((p) => (
                              <Option key={p} value={p}>
                                <Tag color={providerColorMap[p]}>{providerTypeMap[p]}</Tag>
                              </Option>
                            ))}
                          </Select>
                        </Form.Item>

                        <Form.Item
                          {...restField}
                          name={[name, 'targetType']}
                          label="资源类型"
                          rules={[{ required: true, message: '请选择' }]}
                        >
                          <Select style={{ width: 130 }} placeholder="选择类型">
                            {(() => {
                              const provider = form.getFieldValue(['targets', name, 'providerType']) as DeployProviderType;
                              if (!provider) return null;
                              return targetTypeOptions[provider]?.map((t) => (
                                <Option key={t.value} value={t.value}>
                                  {t.label}
                                </Option>
                              ));
                            })()}
                          </Select>
                        </Form.Item>

                        <Form.Item
                          {...restField}
                          name={[name, 'resourceId']}
                          label="资源 ID"
                          rules={[{ required: true, message: '请输入' }]}
                        >
                          <Input style={{ width: 200 }} placeholder="如：域名或实例ID" />
                        </Form.Item>

                        <Form.Item
                          {...restField}
                          name={[name, 'credentialId']}
                          label="云凭证"
                          rules={[{ required: true, message: '请选择' }]}
                        >
                          <Select style={{ width: 180 }} placeholder="选择凭证">
                            {(() => {
                              const provider = form.getFieldValue(['targets', name, 'providerType']) as DeployProviderType;
                              if (!provider) return null;
                              const creds = getCredentialsByProvider(provider);
                              return creds.map((c) => (
                                <Option key={c.id} value={c.id}>
                                  {c.name}
                                </Option>
                              ));
                            })()}
                          </Select>
                        </Form.Item>
                      </Space>
                    </Card>
                  ))}
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    block
                    icon={<PlusOutlined />}
                  >
                    添加目标
                  </Button>
                </>
              )}
            </Form.List>
          </Form>
        );
      case 2:
        const targets = form.getFieldValue('targets') || [];
        return (
          <Form form={form} layout="vertical">
            <Form.Item
              name="name"
              label="任务名称"
              rules={[{ required: true, message: '请输入任务名称' }]}
            >
              <Input placeholder="输入部署任务名称" />
            </Form.Item>

            <Descriptions title="已选证书" bordered size="small" column={1}>
              <Descriptions.Item label="域名">{selectedCert?.domain}</Descriptions.Item>
              <Descriptions.Item label="颁发者">{selectedCert?.issuer}</Descriptions.Item>
              <Descriptions.Item label="过期时间">
                {selectedCert?.expireAt ? new Date(selectedCert.expireAt).toLocaleString() : '-'}
              </Descriptions.Item>
            </Descriptions>

            <div style={{ marginTop: 16 }}>
              <Descriptions title="部署目标" bordered size="small" column={1}>
                {targets.map((target: DeployTarget & { providerType: DeployProviderType }, index: number) => (
                  <Descriptions.Item key={index} label={`目标 ${index + 1}`}>
                    <Space>
                      <Tag color={providerColorMap[target.providerType]}>
                        {providerTypeMap[target.providerType]}
                      </Tag>
                      <Tag>{targetTypeOptions[target.providerType]?.find(t => t.value === target.targetType)?.label || target.targetType}</Tag>
                      <span>资源: {target.resourceId}</span>
                    </Space>
                  </Descriptions.Item>
                ))}
              </Descriptions>
            </div>
          </Form>
        );
      default:
        return null;
    }
  };

  return (
    <Modal
      title="创建部署任务"
      open={visible}
      width={800}
      onCancel={handleCancel}
      footer={null}
      destroyOnClose
    >
      <Steps current={currentStep} style={{ marginBottom: 24 }}>
        <Step title="选择证书" />
        <Step title="选择部署目标" />
        <Step title="确认" />
      </Steps>

      <div style={{ minHeight: 300 }}>{renderStepContent()}</div>

      <div style={{ marginTop: 24, textAlign: 'right' }}>
        {currentStep > 0 && (
          <Button style={{ marginRight: 8 }} onClick={handlePrev}>
            上一步
          </Button>
        )}
        {currentStep < 2 && (
          <Button type="primary" onClick={handleNext}>
            下一步
          </Button>
        )}
        {currentStep === 2 && (
          <Button type="primary" onClick={handleSubmit} loading={submitting}>
            确认创建
          </Button>
        )}
      </div>
    </Modal>
  );
};

export default CreateDeployForm;
