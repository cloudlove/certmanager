import React from 'react';
import {
  Drawer,
  Descriptions,
  Tag,
  Typography,
  Button,
  Space,
  message,
} from 'antd';
import { DownloadOutlined, KeyOutlined } from '@ant-design/icons';
import type { CSRRecord } from '@/types';
import { downloadCSR, downloadPrivateKey } from '@/api/csr';

const { Paragraph } = Typography;

interface CSRDetailProps {
  open: boolean;
  onClose: () => void;
  record: CSRRecord | null;
}

const CSRDetail: React.FC<CSRDetailProps> = ({ open, onClose, record }) => {
  if (!record) return null;

  const handleDownloadCSR = async () => {
    try {
      const blob = await downloadCSR(record.id);
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${record.commonName}.csr.pem`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      message.success('CSR 下载成功');
    } catch {
      message.error('下载失败');
    }
  };

  const handleDownloadPrivateKey = async () => {
    try {
      const blob = await downloadPrivateKey(record.id);
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `${record.commonName}.key.pem`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      message.success('私钥下载成功');
    } catch {
      message.error('下载失败');
    }
  };

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      active: { color: 'success', text: '有效' },
      revoked: { color: 'error', text: '已吊销' },
    };
    const config = statusMap[status] || { color: 'default', text: status };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const getKeyTypeTag = (keyAlgorithm: string) => {
    const color = keyAlgorithm === 'RSA' ? 'blue' : 'green';
    return <Tag color={color}>{keyAlgorithm}</Tag>;
  };

  return (
    <Drawer
      title="CSR 详情"
      open={open}
      onClose={onClose}
      width={700}
      footer={
        <Space>
          <Button
            type="primary"
            icon={<DownloadOutlined />}
            onClick={handleDownloadCSR}
          >
            下载 CSR
          </Button>
          <Button icon={<KeyOutlined />} onClick={handleDownloadPrivateKey}>
            下载私钥
          </Button>
        </Space>
      }
    >
      <Descriptions column={2} bordered size="small">
        <Descriptions.Item label="Common Name" span={2}>
          {record.commonName}
        </Descriptions.Item>
        <Descriptions.Item label="SAN 域名" span={2}>
          <Space size={[0, 8]} wrap>
            {record.san?.map((s, index) => (
              <Tag key={index}>{s}</Tag>
            ))}
          </Space>
        </Descriptions.Item>
        <Descriptions.Item label="密钥算法">
          {getKeyTypeTag(record.keyAlgorithm)}
        </Descriptions.Item>
        <Descriptions.Item label="密钥长度">{record.keySize}</Descriptions.Item>
        <Descriptions.Item label="状态">{getStatusTag(record.status)}</Descriptions.Item>
        <Descriptions.Item label="创建时间">
          {new Date(record.createdAt).toLocaleString()}
        </Descriptions.Item>
      
        <Descriptions.Item label="国家/地区">{record.countryCode || '-'}</Descriptions.Item>
        <Descriptions.Item label="省份">{record.province || '-'}</Descriptions.Item>
        <Descriptions.Item label="城市">{record.locality || '-'}</Descriptions.Item>
        <Descriptions.Item label="单位名称">{record.corpName || '-'}</Descriptions.Item>
        <Descriptions.Item label="部门">{record.department || '-'}</Descriptions.Item>
      </Descriptions>

      <div style={{ marginTop: 24 }}>
        <Typography.Title level={5}>CSR PEM 内容</Typography.Title>
        <Paragraph
          copyable
          style={{
            backgroundColor: '#f5f5f5',
            padding: 16,
            borderRadius: 4,
            fontFamily: 'monospace',
            fontSize: 12,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-all',
            maxHeight: 300,
            overflow: 'auto',
          }}
        >
          {record.csrPem}
        </Paragraph>
      </div>
    </Drawer>
  );
};

export default CSRDetail;
