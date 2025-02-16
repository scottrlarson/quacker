// email.go

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func generateEmailHTML(domain, post, unsubscribeLink, imageURL string) string {
	// Generate the HTML content for the email
	emailBody := `<html>
		<body>`

	// If an image URL is provided, include it above the post link
	if imageURL != "" {
		emailBody += fmt.Sprintf(`<img src="%s" alt="Post Image" style="max-width:400px; width:100%%; height:auto; display:block; margin:0 auto;"/><br>`, imageURL)
	}

	emailBody += fmt.Sprintf(
		`<p><a href="%s">%s</a></p>
		<p style="font-size:small;">You are receiving this email because you subscribed to updates from %s. If you no longer wish to receive these emails, you can <a href="%s">unsubscribe here</a>.</p>
		</body>
		</html>`,
			post, post, domain, unsubscribeLink,
	)

	return emailBody
}

func sendEmail(to, domain, post, emailHTML, owner string) error {
	if to == "" || domain == "" || post == "" || emailHTML == "" || owner == "" {
		return errors.New("invalid email parameters")
	}

	// Fetch Mailgun API key and host from Redis
	configJSON, err := rdb.Get(ctx, "config").Result()
	if err != nil {
		return errors.New("Failed to retrieve Mailgun configuration")
	}

	var config Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return errors.New("Failed to parse Mailgun configuration")
	}

	// Fetch the reply-to email for the site
	replyTo, err := rdb.Get(ctx, "user_sites:"+owner+":"+domain).Result()
	if err != nil || replyTo == "" {
		fmt.Println("Failed to retrieve reply-to email for domain", domain, "- Using default")
		replyTo = "no-reply@" + config.MailgunHost // Default fallback
	}

	// Set the from address to no-reply from the configured Mailgun domain
	fromAddress := "no-reply@" + config.MailgunHost

	// Prepare Mailgun API request
	mailgunAPI := "https://api.mailgun.net/v3/" + config.MailgunHost + "/messages"
	form := url.Values{}
	form.Add("from", fromAddress)
	form.Add("to", to)
	form.Add("subject", "New post from " + domain)
	form.Add("html", emailHTML)
	form.Add("h:Reply-To", replyTo)

	req, err := http.NewRequest("POST", mailgunAPI, strings.NewReader(form.Encode()))
	if err != nil {
		return errors.New("Failed to create Mailgun request")
	}
	req.SetBasicAuth("api", config.MailgunAPIKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send email: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Mailgun API error: %s", resp.Status)
	}

	fmt.Printf("Email successfully sent to %s with from %s and reply-to %s\n", to, fromAddress, replyTo)
	return nil
}
