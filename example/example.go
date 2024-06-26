package example

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/antonyho/redismock"
)

var _ = example

func example() {
	var ctx = context.TODO()

	// get redis.Client and mock
	db, mock := redismock.NewClientMock()

	//the order of commands expected and executed must be the same
	//this is the default value
	mock.MatchExpectationsInOrder(true)

	//simple matching

	//hget command return error
	mock.ExpectHGet("key", "field").SetErr(errors.New("error"))
	//db.HGet(ctx, "key", "field").Err() == errors.New("error")

	//hget command return value
	mock.ExpectHGet("key", "field").SetVal("test value")
	//db.HGet(ctx, "key", "field").Val() == "test value"

	//hget command return redis.Nil
	mock.ExpectHGet("key", "field").RedisNil()
	//db.HGet(ctx, "key", "field").Err() == redis.Nil

	//hget command... do not set return
	mock.ExpectHGet("key", "field")
	//db.HGet(ctx, "key", "field").Err() != nil

	//------------

	//clean up all expectations
	//reset expected redis command
	mock.ClearExpect()

	//regular, uncertain value

	db.HSet(ctx, "key", "field", time.Now().Unix())
	mock.Regexp().ExpectHSet("key", "field", `^[0-9]+$`).SetVal(1)

	//------------
	mock.ClearExpect()

	//custom match, regular expression can not meet the requirements
	mock.CustomMatch(func(expected, actual []interface{}) error {
		//expected == cmd.Args()
		return nil
	}).ExpectGet("key").SetVal("value")

	//--------

	//all the expected redis commands have been matched
	//otherwise return an error
	if err := mock.ExpectationsWereMet(); err != nil {
		//error
		panic(err)
	}

	//--------

	//unexpected redis commands have been called
	//returns true and the unexpected calls
	if unexpected, calls := mock.UnexpectedCallsWereMade(); unexpected {
		log.Printf("Unexpected calls: %+v", calls)
	}

	//---------

	//any order
	//this is useful if your redis commands are executed concurrently
	mock.MatchExpectationsInOrder(false)

	//1-2-3
	mock.ExpectGet("key").SetVal("value")
	mock.ExpectSet("set-key", "set-value", 1).SetVal("OK")
	mock.ExpectHGet("hash-get", "hash-field").SetVal("hash-value")

	//3-1-2
	_ = db.HGet(ctx, "hash-get", "hash-field")
	_ = db.Get(ctx, "key")
	_ = db.Set(ctx, "set-key", "set-value", 1)

	//--------------

	//pipeline, pipeline is not a redis command, is a collection of commands
	mock.ExpectGet("key").SetVal("value")
	mock.ExpectSet("key", "value", 1).SetVal("OK")

	pipe := db.Pipeline()
	_ = pipe.Get(ctx, "key")
	_ = pipe.Set(ctx, "key", "value", 1)
	_, _ = pipe.Exec(ctx)

	//---------------------
	mock.ClearExpect()

	//TxPipeline
	mock.ExpectTxPipeline()
	mock.ExpectGet("key").SetVal("value")
	mock.ExpectSet("key", "value", 1).SetVal("OK")
	mock.ExpectTxPipelineExec()

	pipe = db.TxPipeline()
	_ = pipe.Get(ctx, "key")
	_ = pipe.Set(ctx, "key", "value", 1)
	_, _ = pipe.Exec(ctx)

	//Watch
	mock.ExpectWatch("key1", "key2").SetErr(errors.New("watch error"))
	mock.ExpectGet("key1").SetVal("1")
	mock.ExpectSet("key2", "2", 1*time.Second).SetVal("OK")

	//err := db.Watch(ctx, func(tx *redis.Tx) error {
	//	tx.Get(ctx, "key1")
	//	tx.Set(ctx, "key2", "2", 1 * time.Second)
	//	return nil
	//}, "key1", "key2")
	//reflect.DeepEqual(err, errors.New("watch error"))
}
