package services

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"kingultron99.com/RoleCall/utils"
)

const (
	SystemNoticeMessage = iota
	NewVersionMessage
)

type BroadcastMessageMeta struct {
	Title string
	Color discord.Color
}

var BroadCastMessageType = map[int]BroadcastMessageMeta{
	SystemNoticeMessage: {Title: "System Notice", Color: utils.Blue},
	NewVersionMessage:   {Title: "New Version", Color: utils.Purple},
}

type BroadCastMessage struct {
	Type    int    `json:"type"`
	Content string `json:"content"`
}

func appendResult(
	mu *sync.Mutex,
	results *[]BroadcastResult,
	result BroadcastResult,
) {
	mu.Lock()
	defer mu.Unlock()

	*results = append(*results, result)
}
