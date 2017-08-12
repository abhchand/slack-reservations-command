package main

import(
    "strings"
)

type SlackRequest struct {
    Token           string  `json:"token"`
    TeamId          string  `json:"team_id"`
    TeamDomain      string  `json:"team_domain"`
    EnterpriseId    string  `json:"enterprise_id"`
    EnterpriseName  string  `json:"enterprise_name"`
    ChannelId       string  `json:"channel_id"`
    ChannelName     string  `json:"channel_name"`
    UserId          string  `json:"user_id"`
    UserName        string  `json:"user_name"`
    Command         string  `json:"command"`
    Text            string  `json:"text"`
    ResponseUrl     string  `json:"response_url"`
    TriggerId       string  `json:"trigger_id"`
}

func (sr SlackRequest) FormattedSubcommand() string {

    return strings.ToLower(strings.Trim(sr.Text, " "))

}
