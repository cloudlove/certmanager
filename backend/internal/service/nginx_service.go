package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"certmanager-backend/internal/model"
	"certmanager-backend/internal/repository"
	"certmanager-backend/pkg/crypto"
)

// ClusterVO 集群视图对象
type ClusterVO struct {
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	NodeCount   int      `json:"node_count"`
	OnlineCount int      `json:"online_count"`
	Nodes       []NodeVO `json:"nodes,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// NodeVO 节点视图对象
type NodeVO struct {
	ID            uint   `json:"id"`
	ClusterID     uint   `json:"cluster_id"`
	IP            string `json:"ip"`
	Port          string `json:"port"`
	Status        string `json:"status"`
	LastHeartbeat string `json:"last_heartbeat"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// DeployResult 部署结果
type DeployResult struct {
	NodeID  uint   `json:"node_id"`
	IP      string `json:"ip"`
	Port    string `json:"port"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NginxService Nginx 业务逻辑层
type NginxService struct {
	nginxRepo *repository.NginxRepository
	certRepo  *repository.CertificateRepository
	aesKey    string
}

// NewNginxService 创建 NginxService 实例
func NewNginxService(nginxRepo *repository.NginxRepository, certRepo *repository.CertificateRepository, aesKey string) *NginxService {
	return &NginxService{
		nginxRepo: nginxRepo,
		certRepo:  certRepo,
		aesKey:    aesKey,
	}
}

// toClusterVO 将 NginxCluster 转换为 ClusterVO
func (s *NginxService) toClusterVO(c *model.NginxCluster, nodes []model.NginxNode) *ClusterVO {
	nodeVOs := make([]NodeVO, 0, len(nodes))
	onlineCount := 0
	for _, node := range nodes {
		if node.Status == "online" {
			onlineCount++
		}
		nodeVOs = append(nodeVOs, *s.toNodeVO(&node))
	}

	return &ClusterVO{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		NodeCount:   len(nodes),
		OnlineCount: onlineCount,
		Nodes:       nodeVOs,
		CreatedAt:   c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toClusterVOLite 将 NginxCluster 转换为 ClusterVO（不包含节点详情）
func (s *NginxService) toClusterVOLite(c *model.NginxCluster, nodeCount, onlineCount int) *ClusterVO {
	return &ClusterVO{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		NodeCount:   nodeCount,
		OnlineCount: onlineCount,
		CreatedAt:   c.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toNodeVO 将 NginxNode 转换为 NodeVO
func (s *NginxService) toNodeVO(n *model.NginxNode) *NodeVO {
	lastHeartbeat := ""
	if n.LastHeartbeat != nil {
		lastHeartbeat = n.LastHeartbeat.Format("2006-01-02 15:04:05")
	}
	return &NodeVO{
		ID:            n.ID,
		ClusterID:     n.ClusterID,
		IP:            n.IP,
		Port:          n.Port,
		Status:        n.Status,
		LastHeartbeat: lastHeartbeat,
		CreatedAt:     n.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     n.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// CreateCluster 创建集群
func (s *NginxService) CreateCluster(name, description string) (*ClusterVO, error) {
	if name == "" {
		return nil, errors.New("cluster name is required")
	}

	cluster := &model.NginxCluster{
		Name:        name,
		Description: description,
	}

	if err := s.nginxRepo.CreateCluster(cluster); err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	return s.toClusterVOLite(cluster, 0, 0), nil
}

// GetCluster 获取集群详情
func (s *NginxService) GetCluster(id uint) (*ClusterVO, error) {
	cluster, err := s.nginxRepo.GetClusterByID(id)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	nodes, err := s.nginxRepo.GetNodesByClusterID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	return s.toClusterVO(cluster, nodes), nil
}

// ListClusters 分页获取集群列表
func (s *NginxService) ListClusters(page, pageSize int) ([]*ClusterVO, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	clusters, total, err := s.nginxRepo.ListClusters(page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list clusters: %w", err)
	}

	vos := make([]*ClusterVO, 0, len(clusters))
	for i := range clusters {
		nodes, _ := s.nginxRepo.GetNodesByClusterID(clusters[i].ID)
		onlineCount := 0
		for _, node := range nodes {
			if node.Status == "online" {
				onlineCount++
			}
		}
		vos = append(vos, s.toClusterVOLite(&clusters[i], len(nodes), onlineCount))
	}

	return vos, total, nil
}

// DeleteCluster 删除集群
func (s *NginxService) DeleteCluster(id uint) error {
	if _, err := s.nginxRepo.GetClusterByID(id); err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}
	return s.nginxRepo.DeleteCluster(id)
}

// AddNode 添加节点
func (s *NginxService) AddNode(clusterID uint, ip, port string) (*NodeVO, error) {
	if ip == "" || port == "" {
		return nil, errors.New("ip and port are required")
	}

	// 检查集群是否存在
	if _, err := s.nginxRepo.GetClusterByID(clusterID); err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	node := &model.NginxNode{
		ClusterID: clusterID,
		IP:        ip,
		Port:      port,
		Status:    "offline",
	}

	if err := s.nginxRepo.AddNode(node); err != nil {
		return nil, fmt.Errorf("failed to add node: %w", err)
	}

	return s.toNodeVO(node), nil
}

// RemoveNode 移除节点
func (s *NginxService) RemoveNode(nodeID uint) error {
	if _, err := s.nginxRepo.GetNodeByID(nodeID); err != nil {
		return fmt.Errorf("node not found: %w", err)
	}
	return s.nginxRepo.RemoveNode(nodeID)
}

// DeployToCluster 部署证书到集群
func (s *NginxService) DeployToCluster(clusterID uint, certID uint) ([]DeployResult, error) {
	// 获取集群信息
	cluster, err := s.nginxRepo.GetClusterByID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster not found: %w", err)
	}

	// 获取证书信息
	cert, err := s.certRepo.GetByID(certID)
	if err != nil {
		return nil, fmt.Errorf("certificate not found: %w", err)
	}

	// 获取私钥
	if cert.PrivateKeyEncrypted == "" {
		return nil, errors.New("certificate private key not found")
	}

	privateKeyPEM, err := crypto.Decrypt(cert.PrivateKeyEncrypted, s.aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// 获取集群节点
	nodes, err := s.nginxRepo.GetNodesByClusterID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster nodes: %w", err)
	}

	if len(nodes) == 0 {
		return nil, errors.New("cluster has no nodes")
	}

	// 滚动部署
	results := make([]DeployResult, 0, len(nodes))
	for _, node := range nodes {
		result := s.deployToNode(&node, cert.CertPEM, privateKeyPEM)
		results = append(results, result)

		// 一个节点失败则停止
		if !result.Success {
			break
		}
	}

	_ = cluster // 避免未使用变量警告
	return results, nil
}

// deployToNode 部署证书到单个节点
func (s *NginxService) deployToNode(node *model.NginxNode, certPEM, keyPEM string) DeployResult {
	result := DeployResult{
		NodeID: node.ID,
		IP:     node.IP,
		Port:   node.Port,
	}

	// 构建 Agent URL
	agentURL := fmt.Sprintf("http://%s:%s", node.IP, node.Port)

	// 构建部署请求
	deployReq := map[string]string{
		"cert_pem":  certPEM,
		"key_pem":   keyPEM,
		"cert_path": "/etc/nginx/ssl/cert.pem",
		"key_path":  "/etc/nginx/ssl/key.pem",
	}

	jsonData, err := json.Marshal(deployReq)
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("failed to marshal deploy request: %v", err)
		return result
	}

	// 发送部署请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(agentURL+"/deploy", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("failed to connect to agent: %v", err)
		return result
	}
	defer resp.Body.Close()

	// 解析响应
	var deployResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&deployResp); err != nil {
		result.Success = false
		result.Message = fmt.Sprintf("failed to decode deploy response: %v", err)
		return result
	}

	result.Success = deployResp.Success
	result.Message = deployResp.Message

	return result
}

// ReceiveHeartbeat 接收节点心跳
func (s *NginxService) ReceiveHeartbeat(ip, port, status string) error {
	// 查找节点
	node, err := s.nginxRepo.GetNodeByIPAndPort(ip, port)
	if err != nil {
		return fmt.Errorf("node not found: %w", err)
	}

	// 更新心跳和状态
	if err := s.nginxRepo.UpdateNodeHeartbeat(node.ID); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	// 如果状态变化，更新状态
	if status != "" && status != node.Status {
		if err := s.nginxRepo.UpdateNodeStatus(node.ID, status); err != nil {
			return fmt.Errorf("failed to update node status: %w", err)
		}
	}

	return nil
}
