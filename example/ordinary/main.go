package ordinary

import (
	"github.com/go-redis/redis/v7"
)

const itemKey = "item_cache"

func ItemCache(db *redis.Client, itemID string) (item string, err error) {
	item, err = db.HGet(itemKey, itemID).Result()
	if err == redis.Nil {
		// call api
		item = "information"

		err = db.HSet(itemKey, itemID, item).Err()
	}
	return
}
