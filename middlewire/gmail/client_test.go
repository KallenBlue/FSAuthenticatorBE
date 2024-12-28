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
	"os"
	"regexp"
	"strings"
	"testing"
)

func getClientT(config *oauth2.Config) *http.Client {
	// 从 token.json 文件加载 token
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveTokenT(tokFile, tok)
	}

	// 使用 TokenSource 自动处理令牌刷新
	tokenSource := config.TokenSource(context.Background(), tok)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalf("无法刷新令牌: %v", err)
	}

	// 如果令牌已更新，保存到文件
	if newToken.AccessToken != tok.AccessToken {
		saveTokenT(tokFile, newToken)
	}

	return oauth2.NewClient(context.Background(), tokenSource)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("访问以下链接并授权应用:\n%v\n", authURL)

	// 将授权码直接写入此处
	authCode := "4/0AanRRrsMmaxC_XNnHNYAlc_T01mIENfukp2hlTRUNxAhCfMrUAPeLIfZc4ximZ4Cdq3lIQ"

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("无法检索令牌: %v", err)
	}
	return tok
}

func saveTokenT(path string, token *oauth2.Token) {
	fmt.Printf("保存 token 到文件: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("无法缓存 oauth 令牌: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func TestGetMail(t *testing.T) {
	ctx := context.Background()

	// Load credentials.json
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}

	// Parse OAuth2 configuration
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClientT(config)

	// Create Gmail service
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	// Get the list of messages
	user := "me"
	r, err := srv.Users.Messages.List(user).MaxResults(10).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve messages: %v", err)
	}

	// Filter messages
	fmt.Println("Filtering emails...")
	for _, m := range r.Messages {
		// Get full email details
		msg, err := srv.Users.Messages.Get(user, m.Id).Format("full").Do()
		if err != nil {
			log.Fatalf("Unable to retrieve message details: %v", err)
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
				} else {
					fmt.Println("Verification code not found in subject.")
				}
			}
		}
	}
}
