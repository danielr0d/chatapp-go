package redisrepo

import (
	"fmt"
	"time"
)

func userSetKey() string {
	return "users"
}

func sessionKey(client string) string {
	return "session#" + client
}

func chatKey() string {
	return fmt.Sprintln("chat#%d", time.Now().UnixMilli())
}

func chatIndex() string {
	return "id#chats"
}

func contactListZKey(username string) string {
	return "contacts: " + username
}
