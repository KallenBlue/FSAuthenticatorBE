package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"regexp"
	"strings"
	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"
)

var GmailServerMap map[string]*gmail.Service

func Init() {
	GmailServerMap = make(map[string]*gmail.Service)
	credentialList, err := dao.GCImp.GetAll()
	if err != nil {
		panic(err)
	}
	for _, credential := range credentialList {
		gmailService, err := getGmailService(credential)
		if err != nil {
			panic(err)
		}
		GmailServerMap[credential.Email] = gmailService
	}
}

func GetEmailCode(email string) (string, error) {
	srv, ok := GmailServerMap[email]
	if !ok {
		fmt.Printf("Email %s not found in the credential list.\n", email)
		return "", fmt.Errorf("Email %s not found in the credential list.\n", email)
	}
	// Get the list of messages
	user := "me"
	r, err := srv.Users.Messages.List(user).MaxResults(10).Do()
	if err != nil {
		fmt.Printf("An error occurred while retrieving messages: %v\n", err)
		return "", err
	}
	// Filter messages
	fmt.Println("Filtering emails...")
	for _, m := range r.Messages {
		// Get full email details
		msg, err := srv.Users.Messages.Get(user, m.Id).Format("full").Do()
		if err != nil {
			fmt.Printf("An error occurred while retrieving a message: %v\n", err)
			return "", err
		}

		// Get the subject
		var subject string
		for _, header := range msg.Payload.Headers {
			if header.Name == "Subject" {
				subject = header.Value
				break
			}
		}

		// Check if the subject matches the criteria
		if len(subject) > 0 && strings.HasPrefix(subject, "Your ChatGPT code is") {
			// Check if the snippet matches the criteria
			if len(msg.Snippet) > 0 && strings.Contains(msg.Snippet, "We noticed a suspicious log-in on your account. If that was you, enter this code:") {
				// Extract the verification code from the subject
				re := regexp.MustCompile(`Your ChatGPT code is (\d{6})`)
				matches := re.FindStringSubmatch(subject)
				if len(matches) > 1 {
					fmt.Printf("Found verification code: %s\n", matches[1])
					return matches[1], nil
				} else {
					fmt.Println("Verification code not found in subject.")
				}
			}
		}
	}
	return "", nil
}

func getGmailService(credential model.GmailCredential) (*gmail.Service, error) {
	config, err := google.ConfigFromJSON([]byte(credential.Credential), gmail.GmailReadonlyScope)
	if err != nil {
		return nil, err
	}
	client := getClient(config, credential)
	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func getClient(config *oauth2.Config, credential model.GmailCredential) *http.Client {
	tok := &oauth2.Token{}
	// 从 token.json 文件加载 token
	err := json.Unmarshal([]byte(credential.Token), tok)
	if err != nil {
		tok = getNewToken(config, credential.AuthCode)
		saveToken(tok, credential.Email)
	}

	// 使用 TokenSource 自动处理令牌刷新
	tokenSource := config.TokenSource(context.Background(), tok)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalf("无法刷新令牌: %v", err)
	}

	// 如果令牌已更新，保存到文件
	if newToken.AccessToken != tok.AccessToken {
		saveToken(newToken, credential.Email)
	}

	return oauth2.NewClient(context.Background(), tokenSource)
}

func saveToken(tok *oauth2.Token, email string) {
	bytes, err := json.Marshal(tok)
	if err != nil {
		log.Fatalf("无法序列化令牌: %v", err)
	}
	err = dao.GCImp.UpdateToken(email, string(bytes))
	if err != nil {
		log.Fatalf("无法更新令牌: %v", err)
	}
}

func getNewToken(config *oauth2.Config, authCode string) *oauth2.Token {

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("无法检索令牌: %v", err)
	}
	return tok
}
