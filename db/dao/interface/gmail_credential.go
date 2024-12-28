package _interface

import "wxcloudrun-golang/db/model"

type GmailCredentialInterface interface {
	GetAll() ([]model.GmailCredential, error)
	GetCredential(email string) (string, error)
	UpdateToken(email string, token string) error
}
