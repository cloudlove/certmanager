import { useState, useEffect, useCallback } from 'react';
import {
  Card,
  Typography,
  Button,
  Space,
  Table,
  Tag,
  Badge,
  Switch,
  Popconfirm,
  message,
  Tooltip,
  Tabs,
  Drawer,
  Form,
  Input,
  Select,
  InputNumber,
  Checkbox,
  Row,
  Col,
  // DatePicker, // 预留时间范围筛选功能
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SendOutlined,
  BellOutlined,
  FileTextOutlined,
} from '@ant-design/icons';
// import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import type { NotificationRule, NotificationLog } from '@/api/notification';
import {
  getRules,
  createRule,
  updateRule,
  deleteRule,
  toggleRule,
  testRule,
  getNotificationLogs,
  type CreateRuleRequest,
  type UpdateRuleRequest,
} from '@/api/notification';

const { Title } = Typography;
const { TabPane } = Tabs;
const { Option } = Select;
// const { RangePicker } = DatePicker; // 预留时间范围筛选功能

// 事件类型映射
const eventTypeMap: Record<string, { text: string; color: string }> = {
  cert_expiry: { text: '证书到期', color: 'orange' },
  deploy_success: { text: '部署成功', color: 'green' },
  deploy_failed: { text: '部署失败', color: 'red' },
};

// 通知渠道映射
const channelMap: Record<string, { text: string; color: string }> = {
  email: { text: '邮件', color: 'blue' },
  dingtalk: { text: '钉钉', color: 'cyan' },
  wecom: { text: '企微', color: 'green' },
  webhook: { text: 'Webhook', color: 'purple' },
};

// 发送状态映射
const statusMap: Record<string, { text: string; status: 'success' | 'error' }> = {
  success: { text: '成功', status: 'success' },
  failed: { text: '失败', status: 'error' },
};

const NotificationPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState('rules');
  
  // 规则相关状态
  const [rulesLoading, setRulesLoading] = useState(false);
  const [rulesData, setRulesData] = useState<NotificationRule[]>([]);
  const [rulesPagination, setRulesPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [editingRule, setEditingRule] = useState<NotificationRule | null>(null);
  const [form] = Form.useForm();
  const [submitLoading, setSubmitLoading] = useState(false);
  const [testLoading, setTestLoading] = useState<Record<string, boolean>>({});

  // 日志相关状态
  const [logsLoading, setLogsLoading] = useState(false);
  const [logsData, setLogsData] = useState<NotificationLog[]>([]);
  const [logsPagination, setLogsPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [eventTypeFilter, setEventTypeFilter] = useState<string>('');
  // const [dateRange, setDateRange] = useState<[Dayjs | null, Dayjs | null] | null>(null); // 预留时间范围筛选

  // 获取规则列表
  const fetchRules = useCallback(async (page = 1, pageSize = 10) => {
    setRulesLoading(true);
    try {
      const res = await getRules({ page, pageSize });
      setRulesData(res.list);
      setRulesPagination({
        current: page,
        pageSize,
        total: res.total,
      });
    } catch (error) {
      message.error('获取通知规则失败');
    } finally {
      setRulesLoading(false);
    }
  }, []);

  // 获取日志列表
  const fetchLogs = useCallback(async (page = 1, pageSize = 10) => {
    setLogsLoading(true);
    try {
      const res = await getNotificationLogs({ 
        page, 
        pageSize, 
        eventType: eventTypeFilter || undefined 
      });
      setLogsData(res.list);
      setLogsPagination({
        current: page,
        pageSize,
        total: res.total,
      });
    } catch (error) {
      message.error('获取通知日志失败');
    } finally {
      setLogsLoading(false);
    }
  }, [eventTypeFilter]);

  useEffect(() => {
    fetchRules();
  }, [fetchRules]);

  useEffect(() => {
    if (activeTab === 'logs') {
      fetchLogs();
    }
  }, [activeTab, fetchLogs]);

  // 处理规则分页变化
  const handleRulesTableChange = (pagination: { current?: number; pageSize?: number }) => {
    const { current = 1, pageSize = 10 } = pagination;
    fetchRules(current, pageSize);
  };

  // 处理日志分页变化
  const handleLogsTableChange = (pagination: { current?: number; pageSize?: number }) => {
    const { current = 1, pageSize = 10 } = pagination;
    fetchLogs(current, pageSize);
  };

  // 打开创建抽屉
  const handleCreate = () => {
    setEditingRule(null);
    form.resetFields();
    form.setFieldsValue({
      enabled: true,
      thresholdDays: 30,
      channels: [],
      recipients: [''],
    });
    setDrawerVisible(true);
  };

  // 打开编辑抽屉
  const handleEdit = (record: NotificationRule) => {
    setEditingRule(record);
    form.setFieldsValue({
      name: record.name,
      eventType: record.eventType,
      thresholdDays: record.thresholdDays,
      channels: record.channels,
      recipients: record.recipients.length > 0 ? record.recipients : [''],
      enabled: record.enabled,
    });
    setDrawerVisible(true);
  };

  // 提交表单
  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitLoading(true);

      // 过滤掉空的接收人
      const recipients = values.recipients.filter((r: string) => r.trim() !== '');
      
      if (recipients.length === 0) {
        message.error('请至少填写一个接收人');
        setSubmitLoading(false);
        return;
      }

      const data = {
        ...values,
        recipients,
      };

      if (editingRule) {
        await updateRule(editingRule.id, data as UpdateRuleRequest);
        message.success('更新规则成功');
      } else {
        await createRule(data as CreateRuleRequest);
        message.success('创建规则成功');
      }

      setDrawerVisible(false);
      fetchRules(rulesPagination.current, rulesPagination.pageSize);
    } catch (error) {
      // 表单验证失败或请求失败
    } finally {
      setSubmitLoading(false);
    }
  };

  // 删除规则
  const handleDelete = async (id: string) => {
    try {
      await deleteRule(id);
      message.success('删除规则成功');
      fetchRules(rulesPagination.current, rulesPagination.pageSize);
    } catch (error) {
      message.error('删除规则失败');
    }
  };

  // 切换规则状态
  const handleToggle = async (id: string, enabled: boolean) => {
    try {
      await toggleRule(id, enabled);
      message.success(enabled ? '已启用' : '已禁用');
      fetchRules(rulesPagination.current, rulesPagination.pageSize);
    } catch (error) {
      message.error('操作失败');
    }
  };

  // 测试规则
  const handleTest = async (id: string) => {
    setTestLoading({ ...testLoading, [id]: true });
    try {
      await testRule(id);
      message.success('测试通知已发送');
    } catch (error) {
      message.error('测试通知发送失败');
    } finally {
      setTestLoading({ ...testLoading, [id]: false });
    }
  };

  // 规则表格列
  const rulesColumns: ColumnsType<NotificationRule> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '事件类型',
      dataIndex: 'eventType',
      key: 'eventType',
      render: (type: string) => {
        const config = eventTypeMap[type] || { text: type, color: 'default' };
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: '阈值天数',
      dataIndex: 'thresholdDays',
      key: 'thresholdDays',
      render: (days: number, record: NotificationRule) => {
        if (record.eventType !== 'cert_expiry') {
          return '-';
        }
        return days ? `${days} 天` : '-';
      },
    },
    {
      title: '通知渠道',
      dataIndex: 'channels',
      key: 'channels',
      render: (channels: string[]) => (
        <Space size="small" wrap>
          {channels.map((channel) => {
            const config = channelMap[channel] || { text: channel, color: 'default' };
            return <Tag key={channel} color={config.color}>{config.text}</Tag>;
          })}
        </Space>
      ),
    },
    {
      title: '接收人',
      dataIndex: 'recipients',
      key: 'recipients',
      render: (recipients: string[]) => (
        <Tooltip title={recipients.join(', ')}>
          <span style={{ maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', display: 'inline-block' }}>
            {recipients.join(', ')}
          </span>
        </Tooltip>
      ),
    },
    {
      title: '启用',
      dataIndex: 'enabled',
      key: 'enabled',
      render: (enabled: boolean, record: NotificationRule) => (
        <Switch
          checked={enabled}
          onChange={(checked) => handleToggle(record.id, checked)}
        />
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_, record: NotificationRule) => (
        <Space size="small">
          <Button
            type="text"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Button
            type="text"
            icon={<SendOutlined />}
            loading={testLoading[record.id]}
            onClick={() => handleTest(record.id)}
          >
            测试
          </Button>
          <Popconfirm
            title="确定删除此规则吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="text" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  // 日志表格列
  const logsColumns: ColumnsType<NotificationLog> = [
    {
      title: '事件类型',
      dataIndex: 'eventType',
      key: 'eventType',
      render: (type: string) => {
        const config = eventTypeMap[type] || { text: type, color: 'default' };
        return <Tag color={config.color}>{config.text}</Tag>;
      },
    },
    {
      title: '内容摘要',
      dataIndex: 'content',
      key: 'content',
      render: (content: string) => (
        <Tooltip title={content}>
          <span style={{ maxWidth: 400, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', display: 'inline-block' }}>
            {content}
          </span>
        </Tooltip>
      ),
    },
    {
      title: '发送状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const config = statusMap[status] || { text: status, status: 'default' as const };
        return <Badge status={config.status} text={config.text} />;
      },
    },
    {
      title: '发送时间',
      dataIndex: 'sentAt',
      key: 'sentAt',
      render: (sentAt: string) => sentAt ? dayjs(sentAt).format('YYYY-MM-DD HH:mm:ss') : '-',
    },
  ];

  // 监听事件类型筛选变化
  useEffect(() => {
    if (activeTab === 'logs') {
      fetchLogs(1, logsPagination.pageSize);
    }
  }, [eventTypeFilter]);

  return (
    <Card>
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <BellOutlined />
              通知规则
            </span>
          }
          key="rules"
        >
          <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Title level={4} style={{ margin: 0 }}>通知规则</Title>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
              创建规则
            </Button>
          </div>
          <Table
            columns={rulesColumns}
            dataSource={rulesData}
            rowKey="id"
            loading={rulesLoading}
            pagination={{
              ...rulesPagination,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条`,
            }}
            onChange={handleRulesTableChange}
          />
        </TabPane>
        <TabPane
          tab={
            <span>
              <FileTextOutlined />
              通知日志
            </span>
          }
          key="logs"
        >
          <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Title level={4} style={{ margin: 0 }}>通知日志</Title>
            <Space>
              <Select
                placeholder="事件类型"
                allowClear
                style={{ width: 150 }}
                value={eventTypeFilter || undefined}
                onChange={(value) => setEventTypeFilter(value || '')}
              >
                <Option value="cert_expiry">证书到期</Option>
                <Option value="deploy_success">部署成功</Option>
                <Option value="deploy_failed">部署失败</Option>
              </Select>
            </Space>
          </div>
          <Table
            columns={logsColumns}
            dataSource={logsData}
            rowKey="id"
            loading={logsLoading}
            pagination={{
              ...logsPagination,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条`,
            }}
            onChange={handleLogsTableChange}
          />
        </TabPane>
      </Tabs>

      {/* 创建/编辑规则抽屉 */}
      <Drawer
        title={editingRule ? '编辑规则' : '创建规则'}
        width={520}
        open={drawerVisible}
        onClose={() => setDrawerVisible(false)}
        footer={
          <Space style={{ float: 'right' }}>
            <Button onClick={() => setDrawerVisible(false)}>取消</Button>
            <Button type="primary" loading={submitLoading} onClick={handleSubmit}>
              {editingRule ? '更新' : '创建'}
            </Button>
          </Space>
        }
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            enabled: true,
            thresholdDays: 30,
            channels: [],
            recipients: [''],
          }}
        >
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: '请输入规则名称' }]}
          >
            <Input placeholder="例如：证书到期提醒" />
          </Form.Item>

          <Form.Item
            label="事件类型"
            name="eventType"
            rules={[{ required: true, message: '请选择事件类型' }]}
          >
            <Select placeholder="选择事件类型">
              <Option value="cert_expiry">证书到期</Option>
              <Option value="deploy_success">部署成功</Option>
              <Option value="deploy_failed">部署失败</Option>
            </Select>
          </Form.Item>

          <Form.Item
            noStyle
            shouldUpdate={(prevValues, currentValues) => prevValues.eventType !== currentValues.eventType}
          >
            {({ getFieldValue }) => {
              const eventType = getFieldValue('eventType');
              if (eventType !== 'cert_expiry') {
                return null;
              }
              return (
                <Form.Item
                  label="阈值天数"
                  name="thresholdDays"
                  rules={[{ required: true, message: '请输入阈值天数' }]}
                >
                  <InputNumber min={1} max={365} style={{ width: '100%' }} placeholder="证书到期前多少天发送通知" />
                </Form.Item>
              );
            }}
          </Form.Item>

          <Form.Item
            label="通知渠道"
            name="channels"
            rules={[{ required: true, message: '请至少选择一个通知渠道' }]}
          >
            <Checkbox.Group style={{ width: '100%' }}>
              <Row>
                <Col span={12}>
                  <Checkbox value="email">邮件</Checkbox>
                </Col>
                <Col span={12}>
                  <Checkbox value="dingtalk">钉钉机器人</Checkbox>
                </Col>
                <Col span={12}>
                  <Checkbox value="wecom">企业微信机器人</Checkbox>
                </Col>
                <Col span={12}>
                  <Checkbox value="webhook">自定义 Webhook</Checkbox>
                </Col>
              </Row>
            </Checkbox.Group>
          </Form.Item>

          <Form.List name="recipients">
            {(fields, { add, remove }) => (
              <>
                <Form.Item label="接收人" required>
                  {fields.map((field) => (
                    <Space key={field.key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                      <Form.Item
                        {...field}
                        validateTrigger={['onChange', 'onBlur']}
                        rules={[
                          {
                            required: true,
                            whitespace: true,
                            message: '请输入接收人',
                          },
                        ]}
                        noStyle
                      >
                        <Input placeholder="邮箱 / 手机号 / Webhook URL" style={{ width: 320 }} />
                      </Form.Item>
                      {fields.length > 1 && (
                        <Button type="link" onClick={() => remove(field.name)}>
                          删除
                        </Button>
                      )}
                    </Space>
                  ))}
                  <Button type="dashed" onClick={() => add()} block>
                    添加接收人
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Form.Item
            label="启用"
            name="enabled"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>
        </Form>
      </Drawer>
    </Card>
  );
};

export default NotificationPage;
