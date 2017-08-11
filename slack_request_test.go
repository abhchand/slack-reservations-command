package main

import(
    "testing"
)

func TestFormattedSubcommand(t *testing.T) {

  sr := SlackRequest{
    Token:          "foo",
    TeamId:         "foo",
    TeamDomain:     "foo",
    EnterpriseId:   "foo",
    EnterpriseName: "foo",
    ChannelId:      "foo",
    ChannelName:    "foo",
    UserId:         "foo",
    UserName:       "jimmy-carter",
    Command:        "/reservations",
    Text:           " SOME SUBCOMMAND   ",
    ResponseUrl:    "foo",
  }

  expected := "some subcommand"
  actual := sr.FormattedSubcommand()

  if actual != "some subcommand" {
    t.Error("expected", expected, "got", actual)
  }

}
