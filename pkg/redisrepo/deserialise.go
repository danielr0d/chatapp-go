package redisrepo

import (
	"chatapp/model"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
)

type Document struct {
	ID      string `json:"_id"`
	Payload []byte `json:"payload"`
	Total   int64  `json:"total"`
}

func Deserialise(res interface{}) []Document {
	switch v := res.(type) {
	case []interface{}:
		if len(v) > 1 {
			total := len(v) - 1
			var docs = make([]Document, 0, total/2)

			for i := 1; i <= total; i = i + 2 {
				arrOfValues := v[i+1].([]interface{})
				value := arrOfValues[len(arrOfValues)-1].(string)

				doc := Document{
					ID:      v[i].(string),
					Payload: []byte(value),
					Total:   v[0].(int64),
				}
				docs = append(docs, doc)
			}
			return docs
		}
	default:
		log.Print("different response type otherthan []interface{}. type: %T", res)
		return nil
	}
	return nil
}

func DeserialiseChat(docs []Document) []model.Chat {
	chats := []model.Chat{}
	for _, doc := range docs {
		var c model.Chat
		json.Unmarshal(doc.Payload, &c)

		c.ID = doc.ID
		chats = append(chats, c)
	}

	return chats
}

func DeserialiseContactList(contacts []redis.Z) []model.ContactList {
	contatList := make([]model.ContactList, 0, len(contacts))

	for _, contact := range contacts {
		contatList = append(contatList, model.ContactList{
			Username:     contact.Member.(string),
			LastActivity: int64(contact.Score),
		})
	}
	return contatList
}
