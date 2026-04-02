package model

// NginxCluster Nginx 集群模型
type NginxCluster struct {
	BaseModel
	Name        string `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description string `gorm:"type:text" json:"description"`
}

// TableName 指定表名
func (NginxCluster) TableName() string {
	return "nginx_clusters"
}
