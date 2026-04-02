import { useEffect } from 'react';
import { Drawer, Form, Input, Select, Button, Space, message } from 'antd';
import { createCredential, updateCredential } from '@/api/credential';
import type { CloudCredential } from '@/types';

const { Option } = Select;
const { TextArea } = Input;

interface CredentialFormProps {
  open: boolean;
  credential: CloudCredential | null;
  onClose: () => void;
  onSuccess: () => void;
}

const providerOptions = [
  { value: 'aliyun', label: '阿里云' },
  { value: 'tencent', label: '腾讯云' },
  { value: 'volcengine', label: '火山云' },
  { value: 'wangsu', label: '网宿' },
  { value: 'aws', label: 'AWS' },
  { value: 'azure', label: 'Azure' },
];

const CredentialForm: React.FC<CredentialFormProps> = ({
  open,
  credential,
  onClose,
  onSuccess,
}) => {
  const [form] = Form.useForm();
  const isEdit = !!credential;

  useEffect(() => {
    if (open && credential) {
      form.setFieldsValue({
        name: credential.name,
        providerType: credential.providerType,
        accessKey: credential.accessKey,
        secretKey: '',
        extraConfig: credential.description || '',
      });
    } else if (open) {
      form.resetFields();
    }
  }, [open, credential, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      if (isEdit && credential) {
        // 编辑模式：如果 secretKey 为空，则不提交
        const updateData: Record<string, unknown> = {
          name: values.name,
          providerType: values.providerType,
          accessKey: values.accessKey,
          extraConfig: values.extraConfig,
        };
        if (values.secretKey) {
          updateData.secretKey = values.secretKey;
        }
        await updateCredential(credential.id, updateData);
        message.success('凭证更新成功');
      } else {
        // 新建模式
        await createCredential({
          name: values.name,
          providerType: values.providerType,
          accessKey: values.accessKey,
          secretKey: values.secretKey,
          extraConfig: values.extraConfig,
        });
        message.success('凭证创建成功');
      }
      
      onClose();
      onSuccess();
    } catch (error) {
      // 表单验证失败或请求失败
      console.error('提交失败:', error);
    }
  };

  const handleClose = () => {
    form.resetFields();
    onClose();
  };

  return (
    <Drawer
      title={isEdit ? '编辑凭证' : '新建凭证'}
      width={480}
      open={open}
      onClose={handleClose}
      footer={
        <Space>
          <Button onClick={handleClose}>取消</Button>
          <Button type="primary" onClick={handleSubmit}>
            确定
          </Button>
        </Space>
      }
    >
      <Form
        form={form}
        layout="vertical"
        autoComplete="off"
      >
        <Form.Item
          name="name"
          label="名称"
          rules={[{ required: true, message: '请输入凭证名称' }]}
        >
          <Input placeholder="请输入凭证名称" maxLength={100} />
        </Form.Item>

        <Form.Item
          name="providerType"
          label="Provider 类型"
          rules={[{ required: true, message: '请选择 Provider 类型' }]}
        >
          <Select placeholder="请选择 Provider 类型">
            {providerOptions.map((option) => (
              <Option key={option.value} value={option.value}>
                {option.label}
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          name="accessKey"
          label="Access Key"
          rules={[{ required: true, message: '请输入 Access Key' }]}
        >
          <Input placeholder="请输入 Access Key" />
        </Form.Item>

        <Form.Item
          name="secretKey"
          label="Secret Key"
          rules={[{ required: !isEdit, message: '请输入 Secret Key' }]}
        >
          <Input.Password
            placeholder={isEdit ? '不修改请留空' : '请输入 Secret Key'}
          />
        </Form.Item>

        <Form.Item
          name="extraConfig"
          label="额外配置"
          extra="可选，JSON 格式"
        >
          <TextArea
            rows={4}
            placeholder="请输入额外配置（JSON 格式）"
          />
        </Form.Item>
      </Form>
    </Drawer>
  );
};

export default CredentialForm;
