// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.```

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"regexp"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"strconv"
	"strings"
	u "ticketservice/internal/utils"
	t "ticketservice/internal/ticketinterfaces"
	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Command struct {
	Token       string `json:"token"`
	Command     string `json:"command"`
	Text        string `json:"text"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	ResponseURL string `json:"response_url"`
}


var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
const slackSigningSecret = ""

type SlackTicketService struct {
	slackClient *slack.Client
	channelAsTicket bool
}

func CreateService() t.BaseTicketService{
	var service SlackTicketService
	return &service
}

func (s *SlackTicketService) Init() error {
	apiToken := os.Getenv("SLACK_API_TOKEN")
	if apiToken == "" {
		u.LogPrint(4,"SLACK_API_TOKEN environment variable not set")
	}
	slackSigningSecret := os.Getenv("SLACK_SIGNING_SECRET")
	if slackSigningSecret == "" {
		u.LogPrint(4,"SLACK_SIGNING_SECRET environment variable not set")
	}
	// Create a new Slack client with your API token
	s.slackClient = slack.New(apiToken)

	// Use the Slack client in your code
	_, err := s.slackClient.AuthTest()
	if err != nil {
		log.Fatalf("Error authenticating with Slack: %s", err)
	}
	log.Println("Successfully authenticated with Slack!")
	// Let's see if the environment wants to use channel as ticket
	// or thread as ticket
	cAsT := os.Getenv("SLACK_CHANNEL_AS_TICKET")
	defaultValue := true
	if cAsT != "" {
		var err error
		defaultValue, err = strconv.ParseBool(cAsT)
		if err != nil {
			u.LogPrint(3,"Error parsing SLACK_CHANNEL_AS_TICKET as bool: %v\n", err)
		}
	}
	s.channelAsTicket = defaultValue
	u.LogPrint(1,"CHANNEL_AS_TICKET is set to "+strconv.FormatBool(s.channelAsTicket))
	return nil
}

func (s *SlackTicketService) createNewChannel(channelName string) (*slack.Channel, error){
	// Check if channel already exists
	channels, _, err := s.slackClient.GetConversations(&slack.GetConversationsParameters{
		ExcludeArchived: true,
	})
	if err != nil {
		return nil, err
	}
	// One could argue we could store this result in memory or some form of memorystore.
	// But I'm not sure the length here would get to a performance impact. Happy to adjust
	// in the future
	for _, channel := range channels {
		if channel.Name == channelName {
			u.LogPrint(1,"Channel "+channel.Name+" already exists")
			return &channel, nil
		}
	}
	// Create channel if it doesn't exist
	channel, err := s.slackClient.CreateConversation(slack.CreateConversationParams{
		ChannelName: channelName,
	})
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (s *SlackTicketService) createChannelAsTicket(ticket *t.Ticket, row t.RecommendationQueryResult) (string, error) {
	lastSlashIndex := strings.LastIndex(row.TargetResource, "/")
	secondToLast := strings.LastIndex(row.TargetResource[:lastSlashIndex], "/")
	// This could be moved to BQ Query. But ehh
	ticket.CreationDate = time.Now()
	ticket.LastUpdateDate = time.Now()
	ticket.LastPingDate = time.Now()
	ticket.SnoozeDate = time.Now().AddDate(0,0,7)
	ticket.Subject = fmt.Sprintf("%s-%s",
			row.Recommender_subtype,
			nonAlphanumericRegex.ReplaceAllString(
				row.TargetResource[secondToLast+1:],
				""))
	ticket.RecommenderID = row.Recommender_name
	channelName := fmt.Sprintf("rec-%s-%s",ticket.TargetContact,ticket.Subject)
	channelName = strings.ReplaceAll(channelName, " ", "")
	// According to this document the string length can be a max of 80
	// https://api.slack.com/methods/conversations.create
	sliceLength := 80
	stringLength := len(channelName) - 1
	if stringLength  < sliceLength {
		sliceLength = stringLength
	}
	channelName = strings.ToLower(channelName[0:sliceLength])
	u.LogPrint(1,"Creating Channel: "+channelName)
	channel, err := s.createNewChannel(channelName)
	if err != nil {
		u.LogPrint(3,"Error creating channel")
		return "", err
	}

	ticket.IssueKey = channel.ID
	_, err = s.slackClient.InviteUsersToConversation(channel.ID, ticket.Assignee...)
	if err != nil {
		// If user is already in channel we should continue
		if err.Error() != "already_in_channel" {
			u.LogPrint(3,"Failed to invite users to channel:")
			return channel.ID, err
		}
		u.LogPrint(1,"User(s) were already in channel")
	}

	// Ping Channel with details of the Recommendation
	s.UpdateTicket(ticket, row)
	u.LogPrint(1,"Created Channel: "+channelName+"   with ID: "+channel.ID)
	return channel.ID, nil
}

func (s *SlackTicketService) createThreadAsTicket(ticket *t.Ticket, row t.RecommendationQueryResult) (string, error) {
	lastSlashIndex := strings.LastIndex(row.TargetResource, "/")
	secondToLast := strings.LastIndex(row.TargetResource[:lastSlashIndex], "/")
	// This could be moved to BQ Query. But ehh
	ticket.CreationDate = time.Now()
	ticket.LastUpdateDate = time.Now()
	ticket.LastPingDate = time.Now()
	ticket.SnoozeDate = time.Now().AddDate(0,0,7)
	ticket.Subject = fmt.Sprintf("%s-%s-%s",
			row.Project_name,
			nonAlphanumericRegex.ReplaceAllString(
				row.TargetResource[secondToLast+1:],
				""),
			row.Recommender_subtype)
	ticket.RecommenderID = row.Recommender_name
	channelName := strings.ToLower(ticket.TargetContact)

	// Replace multiple characters using regex to conform to Slack channel name restrictions
	re := regexp.MustCompile(`[\s@#._/:\\*?"<>|]+`)
	channelName = re.ReplaceAllString(channelName, "-")

	u.LogPrint(1, "Creating Channel: "+channelName)
	channel, err := s.createNewChannel(channelName)
	if err != nil {
		u.LogPrint(3, "Error creating channel")
		return "", err
	}
	// Invite users to the channel
	_, err = s.slackClient.InviteUsersToConversation(channel.ID, ticket.Assignee...)
	if err != nil {
		// If user is already in channel we should continue
		if err.Error() != "already_in_channel" {
			u.LogPrint(3,"Failed to invite users to channel:")
			return channel.ID, err
		}
		u.LogPrint(1,"User(s) were already in channel")
	}

	// Send message to the created channel to create "ticket/thread"
	messageOptions := slack.MsgOptionText(ticket.Subject, false)
	_ ,timestamp, err := s.slackClient.PostMessage(channel.ID, messageOptions)
	if err != nil {
		u.LogPrint(3, "Failed to send message to channel")
		return channel.ID, err
	}

	ticket.IssueKey = channel.ID + "-" + timestamp

	s.UpdateTicket(ticket, row)
	u.LogPrint(1, "Created Ticket in Channel: "+channelName+" with ID: "+timestamp)
	return ticket.IssueKey, nil
}

func (s *SlackTicketService) CreateTicket(ticket *t.Ticket, row t.RecommendationQueryResult) (string, error) {
	// One could argue that we should set the function on startup
	// Would save an IF statement. But meh for now
	if s.channelAsTicket{
		return s.createChannelAsTicket(ticket, row)
	}else {
		return s.createThreadAsTicket(ticket, row)
	}
}

// TODO (Ghaun): Update this to take in channel as ticket vs not.
func (s *SlackTicketService) UpdateTicket(ticket *t.Ticket, row t.RecommendationQueryResult) error {
	jsonData, err := json.MarshalIndent(ticket, "", "    ")
	if err != nil {
		return err
	}
	//Convert to code block
	message := fmt.Sprintf("```%s```\n Cost Savings:%v in %v \n Description: %v", 
		string(jsonData),
		row.Impact_cost_unit,
		row.Impact_currency_code,
		row.Description)
	if !s.channelAsTicket {
		// This will return an array. [0] will be channel id [1] will be timestamp
		channelTimestamp := strings.Split(ticket.IssueKey, "-")
		threadMessageOptions := slack.MsgOptionText(message, false)
		_, _, _, err = s.slackClient.SendMessage(channelTimestamp[0], slack.MsgOptionTS(channelTimestamp[1]), threadMessageOptions)
		if err != nil {
			u.LogPrint(3, "Failed to respond in thread")
			return err
		}
	}
	_, _, err = s.slackClient.PostMessage(
		ticket.IssueKey,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		return err
	}
	return nil
}

// CloseTicket is a function that closes an existing channel in Slack based on the given IssueKey.
func (s *SlackTicketService) CloseTicket(key string) error {
	// Use the ArchiveConversation method provided by the Slack API to close the channel with the given IssueKey.
	err := s.slackClient.ArchiveConversation(key)
	if err != nil {
		// If there's an error while closing the channel, return the error.
		return err
	}
	// If the channel was successfully closed, return nil.
	return nil
}

// Incomplete
func (s *SlackTicketService) GetTicket(issueKey string) (t.Ticket, error) {
	conversationInfo, err := s.slackClient.GetConversationInfo(
		&slack.GetConversationInfoInput{
			ChannelID:     issueKey,
			IncludeLocale: false,
		})
	if err != nil {
		return t.Ticket{}, err
	}
	ticket := t.Ticket{
		IssueKey: conversationInfo.ID,
		// Need to determinet the best way to get the ticket information back from slack
		// Will need to do this once testing begings
	}
	return ticket, nil
}

type Message struct {
	Token       string   `json:"token"`
	TeamID      string   `json:"team_id"`
	APIAppID    string   `json:"api_app_id"`
	Event       Event    `json:"event"`
	Text        string   `json:"text"`
	Type        string   `json:"type"`
	AuthedUsers []string `json:"authed_users"`
}

type Event struct {
	Type           string          `json:"type"`
	User           string          `json:"user"`
	Text           string          `json:"text"`
	Ts             string          `json:"ts"`
	Channel        string          `json:"channel"`
	EventTimestamp json.RawMessage `json:"event_ts"`
}

func (s *SlackTicketService) HandleWebhookAction(c echo.Context) error {
    // Read the request body
    defer c.Request().Body.Close()
    body, err := ioutil.ReadAll(c.Request().Body)
    if err != nil {
		return err
    }
    // Verify the request signature
    if !verifyRequestSignature(c.Request().Header, body) {
		return fmt.Errorf("Failed to Verify Request Signature")
    }
    // Parse the event payload
    var event slackevents.EventsAPIEvent
    err = json.Unmarshal(body, &event)
    if err != nil {
        return err
    }
	u.LogPrint(1, "Testing")
    switch event.Type {
    case slackevents.URLVerification:
        var r *slackevents.ChallengeResponse
        err := json.Unmarshal(body, &r)
        if err != nil {
            return err
        }
        return c.JSON(http.StatusOK, r)

    case slackevents.CallbackEvent:
        // Handle other Slack events here
        // ...

    default:
        return c.String(http.StatusInternalServerError, fmt.Sprintf("Unexpected event type: %s", event.Type))
    }
	return nil
}

func verifyRequestSignature(header http.Header, body []byte) bool {
    // Extract the signature and timestamp from the header
    signature := header.Get("X-Slack-Signature")
    timestamp := header.Get("X-Slack-Request-Timestamp")

	u.LogPrint(1, "Signature: %v    Timestamp: %v", signature, timestamp)
    // Ensure the timestamp is not too old
    timestampInt, err := strconv.Atoi(timestamp)
    if err != nil {
        return false
    }
    age := time.Now().Unix() - int64(timestampInt)
    if age > 300 {
        return false
    }
	// Encode the request body as a URL-encoded string
	resultingObject, err := u.ParseJSONToMap(string(body))
	if err != nil {
		u.LogPrint(3, "Body: %v", string(body))
		u.LogPrint(3, "Failed to Parse JSON to Map: %v", err)
		return false
	}
	u.LogPrint(1, "object: %v", len(resultingObject))
	params := make(url.Values)
	for key, value := range resultingObject {
		params.Set(key, fmt.Sprintf("%v", value))
	}
	encoded := params.Encode()
    // Concatenate the timestamp and request body
    sigBaseString := fmt.Sprintf("v0:%s:%s", timestamp, string(encoded))
	u.LogPrint(1, "BaseString: %v", sigBaseString)
    // Hash the base string with the Slack signing secret
    signatureHash := hmac.New(sha256.New, []byte(slackSigningSecret))
    signatureHash.Write([]byte(sigBaseString))
    expectedSignature := fmt.Sprintf("v0=%s", hex.EncodeToString(signatureHash.Sum(nil)))
	u.LogPrint(1, "expected Sig: %v", expectedSignature)

    // Compare the expected signature to the actual signature
	equal := hmac.Equal([]byte(signature), []byte(expectedSignature))
    return equal
}
