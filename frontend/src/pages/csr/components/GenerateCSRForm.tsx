import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Radio, Select, Button, message, Divider, Row, Col, Typography } from 'antd';
import type { RadioChangeEvent } from 'antd/es/radio';
import { PlusOutlined, MinusCircleOutlined, KeyOutlined, EnvironmentOutlined, BankOutlined } from '@ant-design/icons';
import { generateCSR } from '@/api/csr';
import type { GenerateCSRRequest } from '@/api/csr';

const { Option } = Select;
const { Text } = Typography;

interface GenerateCSRFormProps {
  open: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

const GenerateCSRForm: React.FC<GenerateCSRFormProps> = ({
  open,
  onCancel,
  onSuccess,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [keyAlgorithm, setKeyAlgorithm] = useState<'RSA' | 'ECC'>('RSA');

  // 当算法改变时，重置密钥长度
  useEffect(() => {
    if (keyAlgorithm === 'RSA') {
      form.setFieldsValue({ keySize: 2048 });
    } else {
      form.setFieldsValue({ keySize: 256 });
    }
  }, [keyAlgorithm, form]);

  const handleAlgorithmChange = (e: RadioChangeEvent) => {
    setKeyAlgorithm(e.target.value as 'RSA' | 'ECC');
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      const data: GenerateCSRRequest = {
        commonName: values.commonName,
        sans: values.sans?.filter(Boolean) || [],
        keyAlgorithm: values.keyAlgorithm,
        keySize: values.keySize,
        countryCode: values.countryCode,
        province: values.province,
        locality: values.locality,
        corpName: values.corpName,
        department: values.department,
      };

      await generateCSR(data);
      message.success('CSR 生成成功');
      form.resetFields();
      onSuccess();
    } catch (error) {
      // 表单验证失败或请求失败
      if (error instanceof Error) {
        message.error(error.message || '生成失败');
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
      title={
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <KeyOutlined style={{ fontSize: 20, color: '#1890ff' }} />
          <span>生成 CSR 请求</span>
        </div>
      }
      open={open}
      onOk={handleSubmit}
      onCancel={handleCancel}
      confirmLoading={loading}
      width={720}
      destroyOnClose
      styles={{
        body: {
          padding: '24px',
          background: 'linear-gradient(135deg, #f5f7fa 0%, #e4edf5 100%)',
        },
      }}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          keyAlgorithm: 'RSA',
          keySize: 2048,
          countryCode: 'CN',
        }}
        size="middle"
      >
        <Row gutter={16}>
          <Col span={24}>
            <Form.Item
              name="commonName"
              label={<Text strong>通用名称 (Common Name)</Text>}
              rules={[{ required: true, message: '请输入 Common Name' }]}
              tooltip="证书的主要域名，通常为网站主域名"
            >
              <Input 
                placeholder="例如：example.com" 
                prefix={<EnvironmentOutlined style={{ color: '#aaa' }} />}
                size="large"
              />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item 
          label={<Text strong>备用名称 (SAN)</Text>}
          tooltip="可选的附加域名，用于通配符或多个子域"
        >
          <Form.List name="sans">
            {(fields, { add, remove }) => (
              <div style={{ marginTop: 8 }}>
                {fields.map((field, index) => (
                  <Row key={field.key} gutter={8} align="middle" style={{ marginBottom: 8 }}>
                    <Col flex="auto">
                      <Form.Item
                        {...field}
                        noStyle
                        rules={[{
                          validator: (_, value) => {
                            if (!value || value.trim() === '') {
                              return Promise.reject(new Error('请输入域名'));
                            }
                            return Promise.resolve();
                          },
                        }]}
                      >
                        <Input
                          placeholder={`域名 ${index + 1}`} 
                          prefix={<EnvironmentOutlined style={{ color: '#aaa' }} />}
                        />
                      </Form.Item>
                    </Col>
                    <Col flex="none">
                      <Button
                        type="text"
                        danger
                        icon={<MinusCircleOutlined />}
                        onClick={() => remove(field.name)}
                      />
                    </Col>
                  </Row>
                ))}
                <Form.Item>
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    block
                    icon={<PlusOutlined />}
                  >
                    添加域名
                  </Button>
                </Form.Item>
              </div>
            )}
          </Form.List>
        </Form.Item>

        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="keyAlgorithm"
              label={<Text strong>密钥算法</Text>}
              rules={[{ required: true, message: '请选择密钥算法' }]}
            >
              <Radio.Group onChange={handleAlgorithmChange} style={{ width: '100%' }}>
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <Radio.Button value="RSA" style={{ width: '100%', textAlign: 'center' }}>
                      <div style={{ padding: '8px 0' }}>
                        <KeyOutlined style={{ marginRight: 6 }} />
                        RSA
                      </div>
                    </Radio.Button>
                  </Col>
                  <Col span={12}>
                    <Radio.Button value="ECC" style={{ width: '100%', textAlign: 'center' }}>
                      <div style={{ padding: '8px 0' }}>
                        <KeyOutlined style={{ marginRight: 6 }} />
                        ECC
                      </div>
                    </Radio.Button>
                  </Col>
                </Row>
              </Radio.Group>
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="keySize"
              label={<Text strong>密钥长度</Text>}
              rules={[{ required: true, message: '请选择密钥长度' }]}
            >
              <Select size="large">
                {keyAlgorithm === 'RSA' ? (
                  <>
                    <Option value={2048}>2048 位 (推荐)</Option>
                    <Option value={4096}>4096 位 (高安全)</Option>
                  </>
                ) : (
                  <>
                    <Option value={256}>P256 (推荐)</Option>
                    <Option value={384}>P384 (高安全)</Option>
                  </>
                )}
              </Select>
            </Form.Item>
          </Col>
        </Row>
      
        <Divider orientation="left" orientationMargin="0" style={{ marginTop: 16, marginBottom: 24 }}>
          <EnvironmentOutlined style={{ marginRight: 8 }} />
          <Text strong>组织信息</Text>
        </Divider>
      
        <Row gutter={16}>
          <Col span={8}>
            <Form.Item
              name="countryCode"
              label="国家/地区"
              rules={[{ required: true, message: '请输入国家/地区代码' }]}
            >
              <Input 
                placeholder="如 CN、US" 
                maxLength={2}
                style={{ textTransform: 'uppercase' }}
              />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              name="province"
              label="省份/州"
              rules={[{ required: true, message: '请输入省份' }]}
            >
              <Input placeholder="如 Beijing" />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              name="locality"
              label="城市"
              rules={[{ required: true, message: '请输入城市' }]}
            >
              <Input placeholder="如 Haidian" />
            </Form.Item>
          </Col>
        </Row>
        
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="corpName"
              label="公司/组织名称"
            >
              <Input 
                placeholder="如公司全称" 
                prefix={<BankOutlined style={{ color: '#aaa' }} />}
              />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="department"
              label="部门"
            >
              <Input 
                placeholder="如 IT 部门" 
                prefix={<BankOutlined style={{ color: '#aaa' }} />}
              />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </Modal>
  );
};

export default GenerateCSRForm;
