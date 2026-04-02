import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Modal,
  Steps,
  Form,
  Radio,
  Select,
  Table,
  Button,
  Space,
  Typography,
  Tag,
  Alert,
  message,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { CloudProvider, CSRRecord, CloudCredential } from '@/types';
import { applyCertificate } from '@/api/certificate';
import { getCSRList } from '@/api/csr';
import { getCredentials } from '@/api/credential';

const { Step } = Steps;
const { Text } = Typography;

interface ApplyCertFormProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

interface FormData {
  caProvider: CloudProvider;
  credentialId: string;
  csrId: number;
  validateType: string;
  productType: string;   // DV / OV / EV
  domainType: string;    // single / wildcard / multi
}

const providerOptions: { value: CloudProvider; label: string }[] = [
  { value: 'aliyun', label: '阿里云' },
  { value: 'tencent', label: '腾讯云' },
  { value: 'volcengine', label: '火山云' },
];

const validateTypeOptions = [
  { value: 'DNS', label: 'DNS验证（推荐）', description: '需要在域名DNS中添加TXT记录' },
  { value: 'FILE', label: '文件验证', description: '需要在Web服务器根目录放置验证文件' },
];

const productTypeOptions = [
  { value: 'DV', label: 'DV（域名验证）', description: '仅验证域名所有权，颁发速度快' },
  { value: 'OV', label: 'OV（组织验证）', description: '验证组织身份，显示企业信息' },
];

const domainTypeOptions = [
  { value: 'single', label: '单域名', description: '保护单个域名' },
  { value: 'wildcard', label: '通配符', description: '保护主域名及其所有子域名' },
  { value: 'multi', label: '多域名', description: '保护多个不同域名' },
];

const ApplyCertForm: React.FC<ApplyCertFormProps> = ({
  visible,
  onCancel,
  onSuccess,
}) => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [form] = Form.useForm<FormData>();
  const [submitting, setSubmitting] = useState(false);

  const [csrList, setCsrList] = useState<CSRRecord[]>([]);
  const [csrLoading, setCsrLoading] = useState(false);
  const [csrPagination, setCsrPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  const [credentials, setCredentials] = useState<CloudCredential[]>([]);
  const [credentialLoading, setCredentialLoading] = useState(false);

  const [selectedCsr, setSelectedCsr] = useState<CSRRecord | null>(null);

  // 获取凭证列表
  const fetchCredentials = async (provider: CloudProvider) => {
    setCredentialLoading(true);
    try {
      const res = await getCredentials({
        page: 1,
        pageSize: 100,
        providerType: provider,
      });
      setCredentials(res.list);
    } finally {
      setCredentialLoading(false);
    }
  };

  // 获取 CSR 列表
  const fetchCSRList = async (page = 1, pageSize = 10) => {
    setCsrLoading(true);
    try {
      const res = await getCSRList({ page, pageSize });
      setCsrList(res.list);
      setCsrPagination({
        current: res.page,
        pageSize: res.pageSize,
        total: res.total,
      });
    } finally {
      setCsrLoading(false);
    }
  };

  // 当选择 CA 时，获取对应凭证
  const handleProviderChange = (provider: CloudProvider) => {
    form.setFieldsValue({ credentialId: undefined });
    fetchCredentials(provider);
  };

  // 步骤切换
  const nextStep = async () => {
    if (currentStep === 0) {
      const values = await form.validateFields(['caProvider', 'credentialId']);
      if (values.caProvider && values.credentialId) {
        setCurrentStep(1);
        fetchCSRList();
      }
    } else if (currentStep === 1) {
      const values = await form.validateFields(['csrId', 'validateType']);
      if (!values.csrId) {
        message.warning('请选择一个 CSR');
        return;
      }
      if (!values.validateType) {
        message.warning('请选择域名验证方式');
        return;
      }
      const csr = csrList.find((c) => c.id === values.csrId);
      setSelectedCsr(csr || null);
      setCurrentStep(2);
    }
  };

  const prevStep = () => {
    setCurrentStep(currentStep - 1);
  };

  // 提交申请
  const handleSubmit = async () => {
    const values = await form.validateFields();
    if (!selectedCsr) {
      message.error('未选择 CSR');
      return;
    }

    setSubmitting(true);
    try {
      const result = await applyCertificate({
        caProvider: values.caProvider,
        domain: selectedCsr.commonName,
        csrId: values.csrId,
        credentialId: values.credentialId,
        validateType: values.validateType,
        productType: values.productType,
        domainType: values.domainType,
      });
      message.success('证书申请已提交');
      onSuccess();
      handleClose();
      // 跳转到证书详情页
      if (result.id) {
        navigate(`/certificates/${result.id}`);
      }
    } finally {
      setSubmitting(false);
    }
  };

  // 关闭并重置
  const handleClose = () => {
    form.resetFields();
    setCurrentStep(0);
    setSelectedCsr(null);
    setCredentials([]);
    setCsrList([]);
    onCancel();
  };

  // CSR 表格列
  const csrColumns: ColumnsType<CSRRecord> = [
    {
      title: '域名 (CN)',
      dataIndex: 'commonName',
      key: 'commonName',
    },
    {
      title: 'SAN 列表',
      dataIndex: 'san',
      key: 'san',
      render: (sans: string[]) => (
        <Space wrap>
          {sans?.map((san) => (
            <Tag key={san}>{san}</Tag>
          ))}
        </Space>
      ),
    },
    {
      title: '密钥算法',
      dataIndex: 'keyAlgorithm',
      key: 'keyAlgorithm',
      render: (type: string, record: CSRRecord) =>
        `${type}-${record.keySize}`,
    },
  ];

  // 步骤内容
  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <Form form={form} layout="vertical">
            <Form.Item
              name="caProvider"
              label="选择 CA 机构"
              rules={[{ required: true, message: '请选择 CA 机构' }]}
            >
              <Radio.Group onChange={(e) => handleProviderChange(e.target.value)}>
                <Space direction="vertical">
                  {providerOptions.map((opt) => (
                    <Radio.Button key={opt.value} value={opt.value}>
                      {opt.label}
                    </Radio.Button>
                  ))}
                </Space>
              </Radio.Group>
            </Form.Item>

            <Form.Item
              name="credentialId"
              label="选择云凭证"
              rules={[{ required: true, message: '请选择云凭证' }]}
            >
              <Select
                placeholder="请选择凭证"
                loading={credentialLoading}
                options={credentials.map((c) => ({
                  value: c.id,
                  label: `${c.name} (${c.accessKey.slice(0, 8)}...)`,
                }))}
              />
            </Form.Item>

            {!credentialLoading && credentials.length === 0 && form.getFieldValue('caProvider') && (
              <Alert
                type="warning"
                message="未找到该云厂商的凭证"
                description="请先在 云凭证管理 页面添加凭证"
                showIcon
              />
            )}

            <Form.Item
              name="productType"
              label="证书级别"
              rules={[{ required: true, message: '请选择证书级别' }]}
              initialValue="DV"
            >
              <Radio.Group>
                <Space direction="vertical">
                  {productTypeOptions.map((opt) => (
                    <Radio.Button key={opt.value} value={opt.value}>
                      <div>
                        <Text strong>{opt.label}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {opt.description}
                        </Text>
                      </div>
                    </Radio.Button>
                  ))}
                </Space>
              </Radio.Group>
            </Form.Item>

            <Form.Item
              name="domainType"
              label="证书类型"
              rules={[{ required: true, message: '请选择证书类型' }]}
              initialValue="single"
            >
              <Radio.Group>
                <Space direction="vertical">
                  {domainTypeOptions.map((opt) => (
                    <Radio.Button key={opt.value} value={opt.value}>
                      <div>
                        <Text strong>{opt.label}</Text>
                        <br />
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {opt.description}
                        </Text>
                      </div>
                    </Radio.Button>
                  ))}
                </Space>
              </Radio.Group>
            </Form.Item>
          </Form>
        );

      case 1:
        return (
          <div>
            <Alert
              type="info"
              message="请选择一个 CSR 用于申请证书"
              style={{ marginBottom: 16 }}
            />
            <Form form={form}>
              <Form.Item
                name="csrId"
                rules={[{ required: true, message: '请选择 CSR' }]}
                hidden
              >
                <input type="hidden" />
              </Form.Item>
            </Form>
            <Table
              rowKey="id"
              columns={csrColumns}
              dataSource={csrList}
              loading={csrLoading}
              pagination={{
                ...csrPagination,
                onChange: (page, pageSize) => fetchCSRList(page, pageSize),
              }}
              rowSelection={{
                type: 'radio',
                selectedRowKeys: form.getFieldValue('csrId') ? [form.getFieldValue('csrId')] : [],
                onChange: (selectedRowKeys) => {
                  form.setFieldsValue({ csrId: Number(selectedRowKeys[0]) });
                },
              }}
            />

            {/* 域名验证方式选择 */}
            <div style={{ marginTop: 24 }}>
              <Typography.Title level={5}>域名验证方式</Typography.Title>
              <Form form={form} layout="vertical">
                <Form.Item
                  name="validateType"
                  rules={[{ required: true, message: '请选择域名验证方式' }]}
                  initialValue="DNS"
                >
                  <Radio.Group>
                    <Space direction="vertical">
                      {validateTypeOptions.map((opt) => (
                        <Radio.Button key={opt.value} value={opt.value}>
                          <div>
                            <Text strong>{opt.label}</Text>
                            <br />
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              {opt.description}
                            </Text>
                          </div>
                        </Radio.Button>
                      ))}
                    </Space>
                  </Radio.Group>
                </Form.Item>
              </Form>
            </div>
          </div>
        );

      case 2:
        const values = form.getFieldsValue();
        const providerLabel = providerOptions.find(
          (p) => p.value === values.caProvider
        )?.label;
        const credential = credentials.find((c) => c.id === values.credentialId);
        const validateTypeLabel = validateTypeOptions.find(
          (v) => v.value === values.validateType
        )?.label;
        const productTypeLabel = productTypeOptions.find(
          (p) => p.value === values.productType
        )?.label;
        const domainTypeLabel = domainTypeOptions.find(
          (d) => d.value === values.domainType
        )?.label;

        return (
          <div>
            <Alert
              type="info"
              message="请确认以下信息"
              style={{ marginBottom: 16 }}
            />

            <div style={{ marginBottom: 24 }}>
              <Typography.Title level={5}>基本信息</Typography.Title>
              <Space direction="vertical" style={{ width: '100%' }}>
                <div>
                  <Text type="secondary">CA 机构: </Text>
                  <Text strong>{providerLabel}</Text>
                </div>
                <div>
                  <Text type="secondary">云凭证: </Text>
                  <Text strong>{credential?.name}</Text>
                </div>
                <div>
                  <Text type="secondary">证书级别: </Text>
                  <Text strong>{productTypeLabel}</Text>
                </div>
                <div>
                  <Text type="secondary">证书类型: </Text>
                  <Text strong>{domainTypeLabel}</Text>
                </div>
                <div>
                  <Text type="secondary">验证方式: </Text>
                  <Text strong>{validateTypeLabel}</Text>
                </div>
              </Space>
            </div>

            <div style={{ marginBottom: 24 }}>
              <Typography.Title level={5}>CSR 信息</Typography.Title>
              {selectedCsr && (
                <Space direction="vertical" style={{ width: '100%' }}>
                  <div>
                    <Text type="secondary">域名: </Text>
                    <Text strong>{selectedCsr.commonName}</Text>
                  </div>
                  <div>
                    <Text type="secondary">密钥算法: </Text>
                    <Text>{selectedCsr.keyAlgorithm}</Text>
                  </div>
                  <div>
                    <Text type="secondary">SAN 列表: </Text>
                    <Space wrap>
                      {selectedCsr.san?.map((s: string) => (
                        <Tag key={s}>{s}</Tag>
                      ))}
                    </Space>
                  </div>
                </Space>
              )}
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <Modal
      title="申请证书"
      open={visible}
      onCancel={handleClose}
      width={800}
      footer={null}
      destroyOnClose
    >
      <Steps current={currentStep} style={{ marginBottom: 24 }}>
        <Step title="选择 CA" />
        <Step title="选择 CSR + 验证方式" />
        <Step title="确认" />
      </Steps>

      <div style={{ minHeight: 300 }}>{renderStepContent()}</div>

      <div style={{ marginTop: 24, textAlign: 'right' }}>
        <Space>
          {currentStep > 0 && (
            <Button onClick={prevStep}>上一步</Button>
          )}
          {currentStep < 2 && (
            <Button type="primary" onClick={nextStep}>
              下一步
            </Button>
          )}
          {currentStep === 2 && (
            <Button
              type="primary"
              onClick={handleSubmit}
              loading={submitting}
            >
              提交申请
            </Button>
          )}
        </Space>
      </div>
    </Modal>
  );
};

export default ApplyCertForm;
