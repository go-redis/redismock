package ordinary

import (
	"errors"
	"github.com/go-redis/redismock/v7"
	"testing"
)

func TestItemCacheFail(t *testing.T) {
	var (
		itemID = "7d373ca58e"
		setErr = errors.New("bomb")
	)
	db, mock := redismock.NewClientMock()

	mock.ExpectHGet(itemKey, itemID).RedisNil()
	mock.Regexp().ExpectHSet(itemKey, itemID, `^[a-z]+$`).SetErr(setErr)

	_, err := ItemCache(db, itemID)
	if err != setErr {
		t.Error("expectation error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestItemCacheSuccess(t *testing.T) {
	var (
		itemID = "7d373ca58e"
	)
	db, mock := redismock.NewClientMock()

	mock.ExpectHGet(itemKey, itemID).RedisNil()
	mock.Regexp().ExpectHSet(itemKey, itemID, `^[a-z]+$`).SetVal(1)

	item, err := ItemCache(db, itemID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if item != "information" {
		t.Errorf("unexpected item: %s", item)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}

	//----------

	//clean up all expectations
	//reset expected redis command
	mock.ClearExpect()
	mock.ExpectHGet(itemKey, itemID).SetVal("news")

	item, err = ItemCache(db, itemID)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}
	if item != "news" {
		t.Errorf("unexpected item: %s", item)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
