package dao

import (
	"wxcloudrun-golang/db"
	interf "wxcloudrun-golang/db/dao/interface"
	"wxcloudrun-golang/db/model"
)

type GmailCredentialImp struct{}

func (g GmailCredentialImp) UpdateToken(email string, token string) error {
	cli := db.Get()
	err := cli.Table(gmailCredentialsTableName).Where("email = ?", email).Update("token", token).Error
	return err
}

func (g GmailCredentialImp) GetAll() ([]model.GmailCredential, error) {
	var err error
	var counters []model.GmailCredential
	cli := db.Get()
	err = cli.Table(gmailCredentialsTableName).Find(&counters).Error
	return counters, err
}

const gmailCredentialsTableName = "gmail_credential"

func (g GmailCredentialImp) GetCredential(email string) (string, error) {
	var err error
	counter := model.GmailCredential{}
	cli := db.Get()
	err = cli.Table(gmailCredentialsTableName).Where("email = ?", email).First(&counter).Error
	return counter.Credential, err
}

var GCImp interf.GmailCredentialInterface = &GmailCredentialImp{}
