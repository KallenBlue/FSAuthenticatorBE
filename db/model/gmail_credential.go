package model

type GmailCredential struct {
	Email      string `gorm:"column:email" json:"email"`
	AuthCode   string `gorm:"column:auth_code" json:"auth_code"`
	Credential string `gorm:"column:credential" json:"credential"`
	Token      string `gorm:"column:token" json:"token"`
}
