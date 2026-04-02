import { useEffect, useState } from 'react';
import { Modal, Form, Input, Select, message } from 'antd';

import { createDomain, updateDomain, type DomainWithVerify } from '@/api/domain';
import { getCertificateOptions, type CertificateOption } from '@/api/certificate';

const { Option } = Select;

interface DomainFormProps {
  open: boolean;
  mode: 'create' | 'edit';
  initialData?: DomainWithVerify;
  onSuccess: () => void;
  onCancel: () => void;
}

// 域名格式校验正则
const DOMAIN_PATTERN = /^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])$/;

const DomainForm: React.FC<DomainFormProps> = ({
  open,
  mode,
  initialData,
  onSuccess,
  onCancel,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [certificateOptions, setCertificateOptions] = useState<CertificateOption[]>([]);
  const [fetchingCertificates, setFetchingCertificates] = useState(false);

  // 加载证书选项
  useEffect(() => {
    if (mode === 'edit' && open) {
      setFetchingCertificates(true);
      getCertificateOptions()
        .then((options) => {
          setCertificateOptions(options);
        })
        .catch(() => {
          message.error('加载证书列表失败');
        })
        .finally(() => {
          setFetchingCertificates(false);
        });
    }
  }, [mode, open]);

  // 设置初始值
  useEffect(() => {
    if (open && initialData) {
      form.setFieldsValue({
        name: initialData.name,
        certificateId: initialData.certificateId ? parseInt(initialData.certificateId) : undefined,
      });
    } else if (open) {
      form.resetFields();
    }
  }, [open, initialData, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      if (mode === 'create') {
        await createDomain({ name: values.name });
        message.success('域名添加成功');
      } else {
        const updateData: { certificateId?: number | null } = {};
        if (values.certificateId !== undefined) {
          updateData.certificateId = values.certificateId;
        }
        await updateDomain(parseInt(initialData!.id), updateData);
        message.success('域名更新成功');
      }

      form.resetFields();
      onSuccess();
    } catch (error) {
      // 表单校验失败或请求失败
      if (error instanceof Error) {
        message.error(error.message);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    form.resetFields();
    onCancel();
  };

  return (
    <Modal
      title={mode === 'create' ? '添加域名' : '编辑域名'}
      open={open}
      onOk={handleSubmit}
      onCancel={handleCancel}
      confirmLoading={loading}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        autoComplete="off"
      >
        {mode === 'create' ? (
          <Form.Item
            name="name"
            label="域名"
            rules={[
              { required: true, message: '请输入域名' },
              {
                pattern: DOMAIN_PATTERN,
                message: '请输入有效的域名格式',
              },
            ]}
          >
            <Input placeholder="例如: example.com" />
          </Form.Item>
        ) : (
          <>
            <Form.Item label="域名">
              <Input value={initialData?.name} disabled />
            </Form.Item>
            <Form.Item
              name="certificateId"
              label="关联证书"
            >
              <Select
                placeholder="请选择关联证书"
                allowClear
                loading={fetchingCertificates}
                showSearch
                optionFilterProp="children"
                filterOption={(input, option) =>
                  (option?.children as unknown as string)
                    ?.toLowerCase()
                    .includes(input.toLowerCase())
                }
              >
                {certificateOptions.map((cert) => (
                  <Option key={cert.id} value={parseInt(cert.id)}>
                    {cert.name} ({cert.domain})
                  </Option>
                ))}
              </Select>
            </Form.Item>
          </>
        )}
      </Form>
    </Modal>
  );
};

export default DomainForm;
