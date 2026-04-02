package model

// CSRRecord CSR 记录模型
type CSRRecord struct {
	BaseModel
	CommonName          string `gorm:"size:255;not null" json:"common_name"`
	SAN                 string `gorm:"size:1000" json:"san"` // 逗号分隔的域名列表
	KeyAlgorithm        string `gorm:"size:50;not null" json:"key_algorithm"`
	KeySize             int    `json:"key_size"`
	CSRPEM              string `gorm:"type:text" json:"-"`
	PrivateKeyEncrypted string `gorm:"type:text" json:"-"`
	Status              string `gorm:"size:50;not null" json:"status"`
	// 阿里云 CreateCsr 参数
	CountryCode string `gorm:"size:10" json:"country_code"`
	Province    string `gorm:"size:100" json:"province"`
	Locality    string `gorm:"size:100" json:"locality"`
	CorpName    string `gorm:"size:255" json:"corp_name"`
	Department  string `gorm:"size:100" json:"department"`
}

// TableName 指定表名
func (CSRRecord) TableName() string {
	return "csr_records"
}
