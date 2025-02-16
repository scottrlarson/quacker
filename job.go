// job.go

package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title string    `xml:"title"`
		Link  string    `xml:"link"`
		Items []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title    string `xml:"title"`
	Link     string `xml:"link"`
	PubDate  string `xml:"pubDate"`
	ImageURL string `xml:"enclosure" attr:"url"`
}

func job() {
	ensureSetup()

	emailLogFile, _ := os.OpenFile("/var/log/quacker/quacker.email.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer emailLogFile.Close()
	emailLogger := log.New(emailLogFile, "EMAIL: ", log.LstdFlags)

	jobLogFile, _ := os.OpenFile("/var/log/quacker/job.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer jobLogFile.Close()
	jobLogger := log.New(jobLogFile, "JOB: ", log.LstdFlags)

	sites, _ := rdb.Keys(ctx, "user_sites:*:*").Result()
	for _, siteKey := range sites {
		parts := strings.Split(siteKey, ":")
		if len(parts) < 3 {
			continue
		}
		owner, domain := parts[1], parts[2]
		subs, _ := rdb.SMembers(ctx, "subs:"+owner+":"+domain).Result()

		if len(subs) == 0 {
			jobLogger.Printf("No subscribers for domain: %s\n", domain)
			continue
		}

		posts := fetchRSS(domain)
		for _, post := range posts {
			jobLogger.Printf("Found new post: %s\n", post.Link)
			for _, subscriber := range subs {
				sentKey := "sent:" + owner + ":" + domain + ":" + subscriber
				if rdb.SIsMember(ctx, sentKey, post.Link).Val() {
					jobLogger.Printf("Skipping duplicate email for %s (post: %s)\n", subscriber, post.Link)
					continue
				}

				jobLogger.Printf("Checking subscriber: %s\n", subscriber)
				if !validateEmailMX(subscriber) {
					rdb.SRem(ctx, "subs:"+owner+":"+domain, subscriber)
					emailLogger.Printf("Removed invalid email: %s for site %s\n", subscriber, domain)
					continue
				}

				unsubscribeLink := fmt.Sprintf("https://%s/unsubscribe?email=%s&domain=%s&owner=%s", domain, url.QueryEscape(subscriber), domain, owner)
				emailHTML := generateEmailHTML(domain, post.Link, unsubscribeLink, post.ImageURL)
				emailLogger.Printf("Preparing to send email to: %s\n", subscriber)
				if err := sendEmail(subscriber, domain, post.Link, emailHTML, owner); err == nil {
					rdb.SAdd(ctx, sentKey, post.Link)
					rdb.Expire(ctx, sentKey, 120*time.Hour)
					jobLogger.Printf("Marked email as sent for %s (post: %s)\n", subscriber, post.Link)
					emailLogger.Printf("Successfully sent email to: %s\n", subscriber)
				} else {
					emailLogger.Printf("Error sending email to %s: %v\n", subscriber, err)
				}
			}
		}
	}
	cleanSent()
}


func validateEmailMX(email string) bool {
	domain := strings.Split(email, "@")[1]
	mxRecords, err := net.LookupMX(domain)
	return err == nil && len(mxRecords) > 0
}

func fetchRSS(domain string) []RSSItem {
	resp, err := http.Get("https://" + domain + "/rss")
	if err != nil || resp.StatusCode != 200 {
		log.Printf("Failed to fetch RSS feed for domain: %s\n", domain)
		return nil
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return extractLinksFromRSS(body)
}

func extractLinksFromRSS(body []byte) []RSSItem {
	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		log.Println("Failed to parse RSS feed")
		return nil
	}
	var posts []RSSItem
	cutoff := time.Now().Add(-96 * time.Hour)
	for _, item := range rss.Channel.Items {
		pubDate, _ := time.Parse(time.RFC1123Z, item.PubDate)
		if pubDate.After(cutoff) {
			posts = append(posts, item)
		}
	}
	if len(posts) == 0 {
		log.Println("No new posts found in RSS feed")
	}
	return posts
}

func cleanSent() {
	keys, err := rdb.Keys(ctx, "sent:*").Result()
	if err != nil {
		log.Printf("Failed to fetch sent keys: %v", err)
		return
	}
	for _, key := range keys {
		ttl, err := rdb.TTL(ctx, key).Result()
		if err != nil {
			log.Printf("Error fetching TTL for key %s: %v", key, err)
			continue
		}
		// If TTL is <= 0, the key is either expired or has no expiration set
		if ttl <= 0 {
			rdb.Del(ctx, key)
		}
	}
}

