package gosak

import (
	"testing"
)

func TestListSlackChannels(t *testing.T) {
	ListSlackChannels()
}

func TestPostSlackMessage(t *testing.T) {
	PostSlackMessage("slack-bot-test", "hello from gosak")
}
