import { useState } from 'react';
import {
  Modal,
  Form,
  Input,
  Button,
  Space,
  message,
} from 'antd';
import { importCertificate } from '@/api/certificate';

interface ImportCertFormProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

interface ImportFormData {
  certPEM: string;
  chainPEM?: string;
  privateKeyPEM?: string;
}

const ImportCertForm: React.FC<ImportCertFormProps> = ({
  visible,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm<ImportFormData>();
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);
      await importCertificate({
        certPEM: values.certPEM,
        chainPEM: values.chainPEM,
        privateKeyPEM: values.privateKeyPEM,
      });
      message.success('证书导入成功');
      onSuccess();
      handleClose();
    } catch (error) {
      // 表单验证错误或 API 错误
      if (error instanceof Error) {
        message.error(error.message);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    form.resetFields();
    onCancel();
  };

  return (
    <Modal
      title="导入证书"
      open={visible}
      onCancel={handleClose}
      width={600}
      footer={null}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        autoComplete="off"
      >
        <Form.Item
          name="certPEM"
          label="证书 PEM"
          rules={[
            { required: true, message: '请输入证书 PEM 内容' },
            {
              validator: (_, value) => {
                if (!value) return Promise.resolve();
                if (value.includes('-----BEGIN CERTIFICATE-----')) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error('证书格式不正确，应包含 -----BEGIN CERTIFICATE-----'));
              },
            },
          ]}
        >
          <Input.TextArea
            rows={8}
            placeholder={`-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
...
-----END CERTIFICATE-----`}
          />
        </Form.Item>

        <Form.Item
          name="chainPEM"
          label="证书链 PEM (可选)"
        >
          <Input.TextArea
            rows={6}
            placeholder={`-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
...
-----END CERTIFICATE-----`}
          />
        </Form.Item>

        <Form.Item
          name="privateKeyPEM"
          label="私钥 PEM (可选)"
        >
          <Input.TextArea
            rows={6}
            placeholder={`-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7MhgwMbRvI0MBZhpJ
...
-----END RSA PRIVATE KEY-----`}
          />
        </Form.Item>

        <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
          <Space>
            <Button onClick={handleClose}>取消</Button>
            <Button
              type="primary"
              onClick={handleSubmit}
              loading={loading}
            >
              导入
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default ImportCertForm;
