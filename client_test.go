package redismock

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RedisMock", func() {
	var (
		client     *redis.Client
		clientMock ClientMock
		disorder   func() map[string]interface{}
	)

	BeforeEach(func() {
		client, clientMock = NewClientMock()
		disorder = func() map[string]interface{} {
			d := make(map[string]interface{})
			for i := 0; i < 16; i++ {
				k, v := fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i)
				d[k] = v
			}
			return d
		}
	})

	AfterEach(func() {
		Expect(client.Close()).NotTo(HaveOccurred())
		Expect(clientMock.ExpectationsWereMet()).NotTo(HaveOccurred())
	})

	Describe("tx pipeline", func() {
		var pipe redis.Pipeliner

		BeforeEach(func() {
			clientMock.ExpectTxPipeline()
			clientMock.ExpectGet("key1").SetVal("pipeline get")
			clientMock.ExpectHGet("hash_key", "hash_field").SetVal("pipeline hash get")
			clientMock.ExpectSet("set_key", "set value", 1*time.Minute).SetVal("OK")
			clientMock.ExpectTxPipelineExec()

			pipe = client.TxPipeline()
		})

		It("tx pipeline order", func() {
			get := pipe.Get(ctx, "key1")
			hashGet := pipe.HGet(ctx, "hash_key", "hash_field")
			set := pipe.Set(ctx, "set_key", "set value", 1*time.Minute)

			_, err := pipe.Exec(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal("pipeline get"))

			Expect(hashGet.Err()).NotTo(HaveOccurred())
			Expect(hashGet.Val()).To(Equal("pipeline hash get"))

			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))
		})

		It("tx pipeline not order", func() {
			clientMock.MatchExpectationsInOrder(false)

			hashGet := pipe.HGet(ctx, "hash_key", "hash_field")
			set := pipe.Set(ctx, "set_key", "set value", 1*time.Minute)
			get := pipe.Get(ctx, "key1")

			_, err := pipe.Exec(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal("pipeline get"))

			Expect(hashGet.Err()).NotTo(HaveOccurred())
			Expect(hashGet.Val()).To(Equal("pipeline hash get"))

			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))
		})
	})

	Describe("pipeline", func() {
		var pipe redis.Pipeliner

		BeforeEach(func() {
			clientMock.ExpectGet("key1").SetVal("pipeline get")
			clientMock.ExpectHGet("hash_key", "hash_field").SetVal("pipeline hash get")
			clientMock.ExpectSet("set_key", "set value", 1*time.Minute).SetVal("OK")

			pipe = client.Pipeline()
		})

		It("pipeline order", func() {
			clientMock.MatchExpectationsInOrder(true)

			get := pipe.Get(ctx, "key1")
			hashGet := pipe.HGet(ctx, "hash_key", "hash_field")
			set := pipe.Set(ctx, "set_key", "set value", 1*time.Minute)

			_, err := pipe.Exec(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal("pipeline get"))

			Expect(hashGet.Err()).NotTo(HaveOccurred())
			Expect(hashGet.Val()).To(Equal("pipeline hash get"))

			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))
		})

		It("pipeline not order", func() {
			clientMock.MatchExpectationsInOrder(false)

			hashGet := pipe.HGet(ctx, "hash_key", "hash_field")
			set := pipe.Set(ctx, "set_key", "set value", 1*time.Minute)
			get := pipe.Get(ctx, "key1")

			_, err := pipe.Exec(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal("pipeline get"))

			Expect(hashGet.Err()).NotTo(HaveOccurred())
			Expect(hashGet.Val()).To(Equal("pipeline hash get"))

			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))
		})
	})

	Describe("watch", func() {
		BeforeEach(func() {
			clientMock.ExpectWatch("key1", "key2")
			clientMock.ExpectGet("key1").SetVal("1")
			clientMock.ExpectSet("key2", "2", 1*time.Second).SetVal("OK")
		})

		It("watch error", func() {
			clientMock.MatchExpectationsInOrder(false)
			txf := func(tx *redis.Tx) error {
				_ = tx.Get(ctx, "key1")
				_ = tx.Set(ctx, "key2", "2", 1*time.Second)
				return errors.New("watch tx error")
			}

			err := client.Watch(ctx, txf, "key1", "key2")
			Expect(err).To(Equal(errors.New("watch tx error")))

			clientMock.ExpectWatch("key3", "key4").SetErr(errors.New("watch error"))
			txf = func(tx *redis.Tx) error {
				return nil
			}

			err = client.Watch(ctx, txf, "key3", "key4")
			Expect(err).To(Equal(errors.New("watch error")))
		})

		It("watch in order", func() {
			clientMock.MatchExpectationsInOrder(true)
			txf := func(tx *redis.Tx) error {
				val, err := tx.Get(ctx, "key1").Int64()
				if err != nil {
					return err
				}
				Expect(val).To(Equal(int64(1)))
				err = tx.Set(ctx, "key2", "2", 1*time.Second).Err()
				if err != nil {
					return err
				}
				return nil
			}

			err := client.Watch(ctx, txf, "key1", "key2")
			Expect(err).NotTo(HaveOccurred())
		})

		It("watch out of order", func() {
			clientMock.MatchExpectationsInOrder(false)
			txf := func(tx *redis.Tx) error {
				err := tx.Set(ctx, "key2", "2", 1*time.Second).Err()
				if err != nil {
					return err
				}
				val, err := tx.Get(ctx, "key1").Int64()
				if err != nil {
					return err
				}
				Expect(val).To(Equal(int64(1)))
				return nil
			}

			err := client.Watch(ctx, txf, "key1", "key2")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("work order", func() {

		BeforeEach(func() {
			clientMock.ExpectGet("key").RedisNil()
			clientMock.ExpectSet("key", "1", 1*time.Second).SetVal("OK")
			clientMock.ExpectGet("key").SetVal("1")
			clientMock.ExpectGetSet("key", "0").SetVal("1")
		})

		It("ordinary", func() {
			get := client.Get(ctx, "key")
			Expect(get.Err()).To(Equal(redis.Nil))
			Expect(get.Val()).To(Equal(""))

			set := client.Set(ctx, "key", "1", 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))

			get = client.Get(ctx, "key")
			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal("1"))

			getSet := client.GetSet(ctx, "key", "0")
			Expect(getSet.Err()).NotTo(HaveOccurred())
			Expect(getSet.Val()).To(Equal("1"))
		})

		It("surplus", func() {
			_ = client.Get(ctx, "key")

			set := client.Set(ctx, "key", "1", 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))

			Expect(clientMock.ExpectationsWereMet()).To(HaveOccurred())

			_ = client.Get(ctx, "key")
			Expect(clientMock.ExpectationsWereMet()).To(HaveOccurred())

			_ = client.GetSet(ctx, "key", "0")
		})

		It("not enough", func() {
			_ = client.Get(ctx, "key")
			_ = client.Set(ctx, "key", "1", 1*time.Second)
			_ = client.Get(ctx, "key")
			_ = client.GetSet(ctx, "key", "0")
			Expect(clientMock.ExpectationsWereMet()).NotTo(HaveOccurred())

			get := client.HGet(ctx, "key", "field")
			Expect(get.Err()).To(HaveOccurred())
			Expect(get.Val()).To(Equal(""))
		})
	})

	Describe("work not order", func() {

		BeforeEach(func() {
			clientMock.MatchExpectationsInOrder(false)

			clientMock.ExpectSet("key", "1", 1*time.Second).SetVal("OK")
			clientMock.ExpectGet("key").SetVal("1")
			clientMock.ExpectGetSet("key", "0").SetVal("1")
		})

		It("ordinary", func() {
			get := client.Get(ctx, "key")
			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal("1"))

			set := client.Set(ctx, "key", "1", 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))

			getSet := client.GetSet(ctx, "key", "0")
			Expect(getSet.Err()).NotTo(HaveOccurred())
			Expect(getSet.Val()).To(Equal("1"))
		})
	})

	Describe("work other match", func() {

		It("regexp match", func() {
			clientMock.Regexp().ExpectSet("key", `^order_id_[0-9]{10}$`, 1*time.Second).SetVal("OK")
			clientMock.Regexp().ExpectSet("key2", `^order_id_[0-9]{4}\-[0-9]{2}\-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}.+$`, 1*time.Second).SetVal("OK")

			set := client.Set(ctx, "key", fmt.Sprintf("order_id_%d", time.Now().Unix()), 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))

			// no regexp
			set = client.Set(ctx, "key2", fmt.Sprintf("order_id_%s", time.Now().Format(time.UnixDate)), 1*time.Second)
			Expect(set.Err()).To(HaveOccurred())
			Expect(set.Val()).To(Equal(""))

			set = client.Set(ctx, "key2", fmt.Sprintf("order_id_%s", time.Now().Format(time.RFC3339)), 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))
		})

		It("custom match", func() {
			clientMock.CustomMatch(func(expected, actual []interface{}) error {
				return errors.New("mismatch")
			}).ExpectGet("key").SetVal("OK")

			get := client.Get(ctx, "key")
			Expect(get.Err()).To(Equal(errors.New("mismatch")))
			Expect(get.Val()).To(Equal(""))

			set := client.Incr(ctx, "key")
			Expect(set.Err()).To(HaveOccurred())
			Expect(set.Err()).NotTo(Equal(errors.New("mismatch")))
			Expect(set.Val()).To(Equal(int64(0)))

			// no match, no pass
			Expect(clientMock.ExpectationsWereMet()).To(HaveOccurred())

			// let AfterEach pass
			clientMock.ClearExpect()
		})

	})

	Describe("work error", func() {

		It("set error", func() {
			clientMock.ExpectGet("key").SetErr(errors.New("set error"))

			get := client.Get(ctx, "key")
			Expect(get.Err()).To(Equal(errors.New("set error")))
			Expect(get.Val()).To(Equal(""))
		})

		It("not set", func() {
			clientMock.ExpectGet("key")

			get := client.Get(ctx, "key")
			Expect(get.Err()).To(HaveOccurred())
			Expect(get.Val()).To(Equal(""))
		})

		It("set zero", func() {
			clientMock.ExpectGet("key").SetVal("")

			get := client.Get(ctx, "key")
			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal(""))
		})

	})

	Describe("expect", func() {

		It("Command", func() {
			commandsInfo := []*redis.CommandInfo{
				{
					Name:        "data",
					Arity:       3,
					Flags:       []string{"get", "set"},
					ACLFlags:    nil,
					FirstKeyPos: 1,
					LastKeyPos:  1,
					StepCount:   1,
					ReadOnly:    true,
				},
				{
					Name:        "buff",
					Arity:       2,
					Flags:       []string{"read"},
					ACLFlags:    nil,
					FirstKeyPos: 1,
					LastKeyPos:  -1,
					StepCount:   1,
					ReadOnly:    true,
				},
			}
			clientMock.ExpectCommand().SetVal(commandsInfo)

			commands, err := client.Command(ctx).Result()
			Expect(err).NotTo(HaveOccurred())

			cmd := commands["data"]
			Expect(cmd.Name).To(Equal("data"))
			Expect(cmd.Arity).To(Equal(int8(3)))
			Expect(cmd.Flags).To(Equal([]string{"get", "set"}))
			Expect(cmd.FirstKeyPos).To(Equal(int8(1)))
			Expect(cmd.LastKeyPos).To(Equal(int8(1)))
			Expect(cmd.StepCount).To(Equal(int8(1)))

			cmd = commands["buff"]
			Expect(cmd.Name).To(Equal("buff"))
			Expect(cmd.Arity).To(Equal(int8(2)))
			Expect(cmd.Flags).To(Equal([]string{"read"}))
			Expect(cmd.FirstKeyPos).To(Equal(int8(1)))
			Expect(cmd.LastKeyPos).To(Equal(int8(-1)))
			Expect(cmd.StepCount).To(Equal(int8(1)))
		})

		It("ClientGetName", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectClientGetName()
			}, func() *redis.StringCmd {
				return client.ClientGetName(ctx)
			})
		})

		It("Echo", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectEcho("mock")
			}, func() *redis.StringCmd {
				return client.Echo(ctx, "mock")
			})
		})

		It("Ping", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectPing()
			}, func() *redis.StatusCmd {
				return client.Ping(ctx)
			})
		})

		It("Quit", func() {
			//not implemented
		})

		It("Del", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectDel()
			}, func() *redis.IntCmd {
				return client.Del(ctx)
			})
		})

		It("Unlink", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectUnlink()
			}, func() *redis.IntCmd {
				return client.Unlink(ctx)
			})
		})

		It("Dump", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectDump("key")
			}, func() *redis.StringCmd {
				return client.Dump(ctx, "key")
			})
		})

		It("Exists", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectExists()
			}, func() *redis.IntCmd {
				return client.Exists(ctx)
			})
		})

		It("Expire", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectExpire("key", 1*time.Second)
			}, func() *redis.BoolCmd {
				return client.Expire(ctx, "key", 1*time.Second)
			})
		})

		It("ExpireAt", func() {
			now := time.Now()
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectExpireAt("key", now.Add(20*time.Minute))
			}, func() *redis.BoolCmd {
				return client.ExpireAt(ctx, "key", now.Add(20*time.Minute))
			})
		})

		It("Keys", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectKeys("key")
			}, func() *redis.StringSliceCmd {
				return client.Keys(ctx, "key")
			})
		})

		It("Migrate", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectMigrate("host", "port", "key", 1, 1*time.Hour)
			}, func() *redis.StatusCmd {
				return client.Migrate(ctx, "host", "port", "key", 1, 1*time.Hour)
			})
		})

		It("Move", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectMove("key", 1)
			}, func() *redis.BoolCmd {
				return client.Move(ctx, "key", 1)
			})
		})

		It("ObjectRefCount", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectObjectRefCount("key")
			}, func() *redis.IntCmd {
				return client.ObjectRefCount(ctx, "key")
			})
		})

		It("ObjectEncoding", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectObjectEncoding("key")
			}, func() *redis.StringCmd {
				return client.ObjectEncoding(ctx, "key")
			})
		})

		It("ObjectIdleTime", func() {
			operationDurationCmd(clientMock, func() *ExpectedDuration {
				return clientMock.ExpectObjectIdleTime("key")
			}, func() *redis.DurationCmd {
				return client.ObjectIdleTime(ctx, "key")
			})
		})

		It("Persist", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectPersist("key")
			}, func() *redis.BoolCmd {
				return client.Persist(ctx, "key")
			})
		})

		It("PExpire", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectPExpire("key", 1*time.Minute)
			}, func() *redis.BoolCmd {
				return client.PExpire(ctx, "key", 1*time.Minute)
			})
		})

		It("PExpireAt", func() {
			now := time.Now()
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectPExpireAt("key", now.Add(10*time.Minute))
			}, func() *redis.BoolCmd {
				return client.PExpireAt(ctx, "key", now.Add(10*time.Minute))
			})
		})

		It("PTTL", func() {
			operationDurationCmd(clientMock, func() *ExpectedDuration {
				return clientMock.ExpectPTTL("key")
			}, func() *redis.DurationCmd {
				return client.PTTL(ctx, "key")
			})
		})

		It("RandomKey", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectRandomKey()
			}, func() *redis.StringCmd {
				return client.RandomKey(ctx)
			})
		})

		It("Rename", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectRename("key", "new_key")
			}, func() *redis.StatusCmd {
				return client.Rename(ctx, "key", "new_key")
			})
		})

		It("RenameNX", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectRenameNX("key", "new_key")
			}, func() *redis.BoolCmd {
				return client.RenameNX(ctx, "key", "new_key")
			})
		})

		It("Restore", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectRestore("key", 1*time.Minute, "value")
			}, func() *redis.StatusCmd {
				return client.Restore(ctx, "key", 1*time.Minute, "value")
			})
		})

		It("RestoreReplace", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectRestoreReplace("key", 1*time.Minute, "value")
			}, func() *redis.StatusCmd {
				return client.RestoreReplace(ctx, "key", 1*time.Minute, "value")
			})
		})

		It("Sort", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectSort("key", &redis.Sort{
					Offset: 0,
					Count:  2,
					Order:  "ASC",
				})
			}, func() *redis.StringSliceCmd {
				return client.Sort(ctx, "key", &redis.Sort{
					Offset: 0,
					Count:  2,
					Order:  "ASC",
				})
			})
		})

		It("SortStore", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectSortStore("key", "store", &redis.Sort{
					Offset: 0,
					Count:  2,
					Order:  "ASC",
				})
			}, func() *redis.IntCmd {
				return client.SortStore(ctx, "key", "store", &redis.Sort{
					Offset: 0,
					Count:  2,
					Order:  "ASC",
				})
			})
		})

		It("SortInterfaces", func() {
			operationSliceCmd(clientMock, func() *ExpectedSlice {
				return clientMock.ExpectSortInterfaces("key", &redis.Sort{
					Get: []string{"object_*"},
				})
			}, func() *redis.SliceCmd {
				return client.SortInterfaces(ctx, "key", &redis.Sort{
					Get: []string{"object_*"},
				})
			})
		})

		It("Touch", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectTouch()
			}, func() *redis.IntCmd {
				return client.Touch(ctx)
			})
		})

		It("TTL", func() {
			operationDurationCmd(clientMock, func() *ExpectedDuration {
				return clientMock.ExpectTTL("key")
			}, func() *redis.DurationCmd {
				return client.TTL(ctx, "key")
			})
		})

		It("Type", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectType("key")
			}, func() *redis.StatusCmd {
				return client.Type(ctx, "key")
			})
		})

		It("Append", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectAppend("key", "value")
			}, func() *redis.IntCmd {
				return client.Append(ctx, "key", "value")
			})
		})

		It("Decr", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectDecr("key")
			}, func() *redis.IntCmd {
				return client.Decr(ctx, "key")
			})
		})

		It("DecrBy", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectDecrBy("key", 1)
			}, func() *redis.IntCmd {
				return client.DecrBy(ctx, "key", 1)
			})
		})

		It("Get", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectGet("key")
			}, func() *redis.StringCmd {
				return client.Get(ctx, "key")
			})
		})

		It("GetRange", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectGetRange("key", 1, 10)
			}, func() *redis.StringCmd {
				return client.GetRange(ctx, "key", 1, 10)
			})
		})

		It("GetSet", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectGetSet("key", 1)
			}, func() *redis.StringCmd {
				return client.GetSet(ctx, "key", 1)
			})
		})

		It("Incr", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectIncr("key")
			}, func() *redis.IntCmd {
				return client.Incr(ctx, "key")
			})
		})

		It("IncrBy", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectIncrBy("key", 1)
			}, func() *redis.IntCmd {
				return client.IncrBy(ctx, "key", 1)
			})
		})

		It("IncrByFloat", func() {
			operationFloatCmd(clientMock, func() *ExpectedFloat {
				return clientMock.ExpectIncrByFloat("key", 1)
			}, func() *redis.FloatCmd {
				return client.IncrByFloat(ctx, "key", 1)
			})
		})

		It("MGet", func() {
			operationSliceCmd(clientMock, func() *ExpectedSlice {
				return clientMock.ExpectMGet()
			}, func() *redis.SliceCmd {
				return client.MGet(ctx)
			})
		})

		It("MSet", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectMSet()
			}, func() *redis.StatusCmd {
				return client.MSet(ctx)
			})
		})

		It("MSet Map", func() {
			clientMock.ExpectMSet(disorder()).SetVal("OK")
			res, err := client.MSet(ctx, disorder()).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("OK"))

			clientMock.ExpectMSet("key1", "value1", "key2", "value2").SetVal("OK")
			res, err = client.MSet(ctx, "key2", "value2", "key1", "value1").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("OK"))
		})

		It("MSetNX", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectMSetNX()
			}, func() *redis.BoolCmd {
				return client.MSetNX(ctx)
			})
		})

		It("MSetNX Map", func() {
			clientMock.ExpectMSetNX(disorder()).SetVal(true)
			res, err := client.MSetNX(ctx, disorder()).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeTrue())

			clientMock.ExpectMSetNX("key1", "value1", "key2", "value2").SetVal(true)
			res, err = client.MSetNX(ctx, "key2", "value2", "key1", "value1").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeTrue())
		})

		It("Set", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectSet("key", "value", 1*time.Minute)
			}, func() *redis.StatusCmd {
				return client.Set(ctx, "key", "value", 1*time.Minute)
			})
		})

		It("SetArgs", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectSetArgs("key", "value", redis.SetArgs{
					Get: true,
				})
			}, func() *redis.StatusCmd {
				return client.SetArgs(ctx, "key", "value", redis.SetArgs{
					Get: true,
				})
			})
		})

		It("SetEX", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectSetEX("key", "value", 1*time.Minute)
			}, func() *redis.StatusCmd {
				return client.SetEX(ctx, "key", "value", 1*time.Minute)
			})
		})

		It("SetNX", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectSetNX("key", "value", 1*time.Minute)
			}, func() *redis.BoolCmd {
				return client.SetNX(ctx, "key", "value", 1*time.Minute)
			})
		})

		It("SetXX", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectSetXX("key", "value", 1*time.Minute)
			}, func() *redis.BoolCmd {
				return client.SetXX(ctx, "key", "value", 1*time.Minute)
			})
		})

		It("SetRange", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectSetRange("key", 1, "value")
			}, func() *redis.IntCmd {
				return client.SetRange(ctx, "key", 1, "value")
			})
		})

		It("StrLen", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectStrLen("key")
			}, func() *redis.IntCmd {
				return client.StrLen(ctx, "key")
			})
		})

		It("GetBit", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectGetBit("key", 1)
			}, func() *redis.IntCmd {
				return client.GetBit(ctx, "key", 1)
			})
		})

		It("SetBit", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectSetBit("key", 1, 2)
			}, func() *redis.IntCmd {
				return client.SetBit(ctx, "key", 1, 2)
			})
		})

		It("BitCount", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectBitCount("key", &redis.BitCount{
					Start: 1,
					End:   2,
				})
			}, func() *redis.IntCmd {
				return client.BitCount(ctx, "key", &redis.BitCount{
					Start: 1,
					End:   2,
				})
			})
		})

		It("BitOpAnd", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectBitOpAnd("dest", "key1", "key2", "key3")
			}, func() *redis.IntCmd {
				return client.BitOpAnd(ctx, "dest", "key1", "key2", "key3")
			})
		})

		It("BitOpOr", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectBitOpOr("dest", "key1", "key2", "key3")
			}, func() *redis.IntCmd {
				return client.BitOpOr(ctx, "dest", "key1", "key2", "key3")
			})
		})

		It("BitOpXor", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectBitOpXor("dest", "key1", "key2", "key3")
			}, func() *redis.IntCmd {
				return client.BitOpXor(ctx, "dest", "key1", "key2", "key3")
			})
		})

		It("BitOpNot", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectBitOpNot("dest", "key")
			}, func() *redis.IntCmd {
				return client.BitOpNot(ctx, "dest", "key")
			})
		})

		It("BitPos", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectBitPos("key", 1, 2, 3)
			}, func() *redis.IntCmd {
				return client.BitPos(ctx, "key", 1, 2, 3)
			})
		})

		It("BitField", func() {
			operationIntSliceCmd(clientMock, func() *ExpectedIntSlice {
				return clientMock.ExpectBitField("key", "INCRBY", "i5", 100, 1, "GET", "u4", 0)
			}, func() *redis.IntSliceCmd {
				return client.BitField(ctx, "key", "INCRBY", "i5", 100, 1, "GET", "u4", 0)
			})
		})

		It("Scan", func() {
			operationScanCmd(clientMock, func() *ExpectedScan {
				return clientMock.ExpectScan(1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.Scan(ctx, 1, "match", 2)
			})
		})

		It("SScan", func() {
			operationScanCmd(clientMock, func() *ExpectedScan {
				return clientMock.ExpectSScan("key", 1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.SScan(ctx, "key", 1, "match", 2)
			})
		})

		It("HScan", func() {
			operationScanCmd(clientMock, func() *ExpectedScan {
				return clientMock.ExpectHScan("key", 1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.HScan(ctx, "key", 1, "match", 2)
			})
		})

		It("ZScan", func() {
			operationScanCmd(clientMock, func() *ExpectedScan {
				return clientMock.ExpectZScan("key", 1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.ZScan(ctx, "key", 1, "match", 2)
			})
		})

		It("HDel", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectHDel("key", "field1", "field2")
			}, func() *redis.IntCmd {
				return client.HDel(ctx, "key", "field1", "field2")
			})
		})

		It("HExists", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectHExists("key", "field")
			}, func() *redis.BoolCmd {
				return client.HExists(ctx, "key", "field")
			})
		})

		It("HGet", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectHGet("key", "field")
			}, func() *redis.StringCmd {
				return client.HGet(ctx, "key", "field")
			})
		})

		It("HGetAll", func() {
			operationStringStringMapCmd(clientMock, func() *ExpectedStringStringMap {
				return clientMock.ExpectHGetAll("key")
			}, func() *redis.StringStringMapCmd {
				return client.HGetAll(ctx, "key")
			})
		})

		It("HIncrBy", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectHIncrBy("key", "field", 1)
			}, func() *redis.IntCmd {
				return client.HIncrBy(ctx, "key", "field", 1)
			})
		})

		It("HIncrByFloat", func() {
			operationFloatCmd(clientMock, func() *ExpectedFloat {
				return clientMock.ExpectHIncrByFloat("key", "field", 1.1)
			}, func() *redis.FloatCmd {
				return client.HIncrByFloat(ctx, "key", "field", 1.1)
			})
		})

		It("HKeys", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectHKeys("key")
			}, func() *redis.StringSliceCmd {
				return client.HKeys(ctx, "key")
			})
		})

		It("HLen", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectHLen("key")
			}, func() *redis.IntCmd {
				return client.HLen(ctx, "key")
			})
		})

		It("HMGet", func() {
			operationSliceCmd(clientMock, func() *ExpectedSlice {
				return clientMock.ExpectHMGet("key", "field1", "field2")
			}, func() *redis.SliceCmd {
				return client.HMGet(ctx, "key", "field1", "field2")
			})
		})

		It("HSet", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectHSet("key", "field1", "value1", "field2", "value2")
			}, func() *redis.IntCmd {
				return client.HSet(ctx, "key", "field1", "value1", "field2", "value2")
			})
		})

		It("HSet Map", func() {
			clientMock.ExpectHSet("key", disorder()).SetVal(1)
			res, err := client.HSet(ctx, "key", disorder()).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(int64(1)))

			clientMock.ExpectHSet("key", "key1", "value1", "key2", "value2").SetVal(1)
			res, err = client.HSet(ctx, "key", "key2", "value2", "key1", "value1").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(int64(1)))
		})

		It("HMSet", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectHMSet("key", "field1", "value1", "field2", "value2")
			}, func() *redis.BoolCmd {
				return client.HMSet(ctx, "key", "field1", "value1", "field2", "value2")
			})
		})

		It("HMSet Map", func() {
			clientMock.ExpectHMSet("key", disorder()).SetVal(true)
			res, err := client.HMSet(ctx, "key", disorder()).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeTrue())

			clientMock.ExpectHMSet("key", "key1", "value1", "key2", "value2").SetVal(true)
			res, err = client.HMSet(ctx, "key", "key2", "value2", "key1", "value1").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeTrue())
		})

		It("HSetNX", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectHSetNX("key", "field", "value")
			}, func() *redis.BoolCmd {
				return client.HSetNX(ctx, "key", "field", "value")
			})
		})

		It("HVals", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectHVals("key")
			}, func() *redis.StringSliceCmd {
				return client.HVals(ctx, "key")
			})
		})

		It("BLPop", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectBLPop(1*time.Second, "key1", "key2")
			}, func() *redis.StringSliceCmd {
				return client.BLPop(ctx, 1*time.Second, "key1", "key2")
			})
		})

		It("BRPop", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectBRPop(1*time.Second, "key1", "key2")
			}, func() *redis.StringSliceCmd {
				return client.BRPop(ctx, 1*time.Second, "key1", "key2")
			})
		})

		It("BRPopLPush", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectBRPopLPush("list1", "list2", 1*time.Minute)
			}, func() *redis.StringCmd {
				return client.BRPopLPush(ctx, "list1", "list2", 1*time.Minute)
			})
		})

		It("LIndex", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectLIndex("key", 1)
			}, func() *redis.StringCmd {
				return client.LIndex(ctx, "key", 1)
			})
		})

		It("LInsert", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectLInsert("list", "BEFORE", "World", "There")
			}, func() *redis.IntCmd {
				return client.LInsert(ctx, "list", "BEFORE", "World", "There")
			})
		})

		It("LInsertBefore", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectLInsertBefore("key", "pivot", "value")
			}, func() *redis.IntCmd {
				return client.LInsertBefore(ctx, "key", "pivot", "value")
			})
		})

		It("LInsertAfter", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectLInsertAfter("key", "pivot", "value")
			}, func() *redis.IntCmd {
				return client.LInsertAfter(ctx, "key", "pivot", "value")
			})
		})

		It("LLen", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectLLen("key")
			}, func() *redis.IntCmd {
				return client.LLen(ctx, "key")
			})
		})

		It("LPop", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectLPop("key")
			}, func() *redis.StringCmd {
				return client.LPop(ctx, "key")
			})
		})

		It("LPos", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectLPos("list", "b", redis.LPosArgs{Rank: 2})
			}, func() *redis.IntCmd {
				return client.LPos(ctx, "list", "b", redis.LPosArgs{Rank: 2})
			})
		})

		It("LPosCount", func() {
			operationIntSliceCmd(clientMock, func() *ExpectedIntSlice {
				return clientMock.ExpectLPosCount("list", "b", 2, redis.LPosArgs{Rank: 2})
			}, func() *redis.IntSliceCmd {
				return client.LPosCount(ctx, "list", "b", 2, redis.LPosArgs{Rank: 2})
			})
		})

		It("LPush", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectLPush("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.LPush(ctx, "key", "value1", "value2")
			})
		})

		It("LPushX", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectLPushX("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.LPushX(ctx, "key", "value1", "value2")
			})
		})

		It("LRange", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectLRange("key", 1, 2)
			}, func() *redis.StringSliceCmd {
				return client.LRange(ctx, "key", 1, 2)
			})
		})

		It("LRem", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectLRem("key", 2, "value")
			}, func() *redis.IntCmd {
				return client.LRem(ctx, "key", 2, "value")
			})
		})

		It("LSet", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectLSet("key", 1, "value")
			}, func() *redis.StatusCmd {
				return client.LSet(ctx, "key", 1, "value")
			})
		})

		It("LTrim", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectLTrim("key", 1, 2)
			}, func() *redis.StatusCmd {
				return client.LTrim(ctx, "key", 1, 2)
			})
		})

		It("RPop", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectRPop("key")
			}, func() *redis.StringCmd {
				return client.RPop(ctx, "key")
			})
		})

		It("RPopLPush", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectRPopLPush("key", "list")
			}, func() *redis.StringCmd {
				return client.RPopLPush(ctx, "key", "list")
			})
		})

		It("RPush", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectRPush("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.RPush(ctx, "key", "value1", "value2")
			})
		})

		It("RPushX", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectRPushX("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.RPushX(ctx, "key", "value1", "value2")
			})
		})

		It("SAdd", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectSAdd("key", "add")
			}, func() *redis.IntCmd {
				return client.SAdd(ctx, "key", "add")
			})
		})

		It("SCard", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectSCard("key")
			}, func() *redis.IntCmd {
				return client.SCard(ctx, "key")
			})
		})

		It("SDiff", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectSDiff("set1", "set2")
			}, func() *redis.StringSliceCmd {
				return client.SDiff(ctx, "set1", "set2")
			})
		})

		It("SDiffStore", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectSDiffStore("set", "set1", "set2")
			}, func() *redis.IntCmd {
				return client.SDiffStore(ctx, "set", "set1", "set2")
			})
		})

		It("SInter", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectSInter()
			}, func() *redis.StringSliceCmd {
				return client.SInter(ctx)
			})
		})

		It("SInterStore", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectSInterStore("set", "set1", "set2")
			}, func() *redis.IntCmd {
				return client.SInterStore(ctx, "set", "set1", "set2")
			})
		})

		It("SIsMember", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectSIsMember("key", "one")
			}, func() *redis.BoolCmd {
				return client.SIsMember(ctx, "key", "one")
			})
		})

		It("SMembers", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectSMembers("key")
			}, func() *redis.StringSliceCmd {
				return client.SMembers(ctx, "key")
			})
		})

		It("SMembersMap", func() {
			operationStringStructMapCmd(clientMock, func() *ExpectedStringStructMap {
				return clientMock.ExpectSMembersMap("key")
			}, func() *redis.StringStructMapCmd {
				return client.SMembersMap(ctx, "key")
			})
		})

		It("SMove", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectSMove("set1", "set2", "two")
			}, func() *redis.BoolCmd {
				return client.SMove(ctx, "set1", "set2", "two")
			})
		})

		It("SPop", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectSPop("key")
			}, func() *redis.StringCmd {
				return client.SPop(ctx, "key")
			})
		})

		It("SPopN", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectSPopN("key", 1)
			}, func() *redis.StringSliceCmd {
				return client.SPopN(ctx, "key", 1)
			})
		})

		It("SRandMember", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectSRandMember("key")
			}, func() *redis.StringCmd {
				return client.SRandMember(ctx, "key")
			})
		})

		It("SRandMemberN", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectSRandMemberN("key", 1)
			}, func() *redis.StringSliceCmd {
				return client.SRandMemberN(ctx, "key", 1)
			})
		})

		It("SRem", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectSRem("set", "one")
			}, func() *redis.IntCmd {
				return client.SRem(ctx, "set", "one")
			})
		})

		It("SUnion", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectSUnion()
			}, func() *redis.StringSliceCmd {
				return client.SUnion(ctx)
			})
		})

		It("SUnionStore", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectSUnionStore("set", "set1", "set2")
			}, func() *redis.IntCmd {
				return client.SUnionStore(ctx, "set", "set1", "set2")
			})
		})

		It("XAdd", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectXAdd(&redis.XAddArgs{
					Stream: "stream",
					ID:     "1-0",
					Values: map[string]interface{}{"uno": "un"},
				})
			}, func() *redis.StringCmd {
				return client.XAdd(ctx, &redis.XAddArgs{
					Stream: "stream",
					ID:     "1-0",
					Values: map[string]interface{}{"uno": "un"},
				})
			})
		})

		It("XDel", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectXDel("stream", "1-0", "2-0", "3-0")
			}, func() *redis.IntCmd {
				return client.XDel(ctx, "stream", "1-0", "2-0", "3-0")
			})
		})

		It("XLen", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectXLen("stream")
			}, func() *redis.IntCmd {
				return client.XLen(ctx, "stream")
			})
		})

		It("XRange", func() {
			operationXMessageSliceCmd(clientMock, func() *ExpectedXMessageSlice {
				return clientMock.ExpectXRange("stream", "-", "+")
			}, func() *redis.XMessageSliceCmd {
				return client.XRange(ctx, "stream", "-", "+")
			})
		})

		It("XRangeN", func() {
			operationXMessageSliceCmd(clientMock, func() *ExpectedXMessageSlice {
				return clientMock.ExpectXRangeN("stream", "-", "+", 2)
			}, func() *redis.XMessageSliceCmd {
				return client.XRangeN(ctx, "stream", "-", "+", 2)
			})
		})

		It("XRevRange", func() {
			operationXMessageSliceCmd(clientMock, func() *ExpectedXMessageSlice {
				return clientMock.ExpectXRevRange("stream", "+", "-")
			}, func() *redis.XMessageSliceCmd {
				return client.XRevRange(ctx, "stream", "+", "-")
			})
		})

		It("XRevRangeN", func() {
			operationXMessageSliceCmd(clientMock, func() *ExpectedXMessageSlice {
				return clientMock.ExpectXRevRangeN("stream", "+", "-", 2)
			}, func() *redis.XMessageSliceCmd {
				return client.XRevRangeN(ctx, "stream", "+", "-", 2)
			})
		})

		It("XRead", func() {
			operationXStreamSliceCmd(clientMock, func() *ExpectedXStreamSlice {
				return clientMock.ExpectXRead(&redis.XReadArgs{
					Streams: []string{"stream", "0"},
					Count:   2,
					Block:   100 * time.Millisecond,
				})
			}, func() *redis.XStreamSliceCmd {
				return client.XRead(ctx, &redis.XReadArgs{
					Streams: []string{"stream", "0"},
					Count:   2,
					Block:   100 * time.Millisecond,
				})
			})
		})

		It("XReadStreams", func() {
			operationXStreamSliceCmd(clientMock, func() *ExpectedXStreamSlice {
				return clientMock.ExpectXReadStreams()
			}, func() *redis.XStreamSliceCmd {
				return client.XReadStreams(ctx)
			})
		})

		It("XGroupCreate", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectXGroupCreate("stream", "group", "0")
			}, func() *redis.StatusCmd {
				return client.XGroupCreate(ctx, "stream", "group", "0")
			})
		})

		It("XGroupCreateMkStream", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectXGroupCreateMkStream("stream", "group", "0")
			}, func() *redis.StatusCmd {
				return client.XGroupCreateMkStream(ctx, "stream", "group", "0")
			})
		})

		It("XGroupSetID", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectXGroupSetID("stream", "group", "0")
			}, func() *redis.StatusCmd {
				return client.XGroupSetID(ctx, "stream", "group", "0")
			})
		})

		It("XGroupDestroy", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectXGroupDestroy("stream", "group")
			}, func() *redis.IntCmd {
				return client.XGroupDestroy(ctx, "stream", "group")
			})
		})

		It("XGroupDelConsumer", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectXGroupDelConsumer("stream", "group", "consumer")
			}, func() *redis.IntCmd {
				return client.XGroupDelConsumer(ctx, "stream", "group", "consumer")
			})
		})

		It("XReadGroup", func() {
			operationXStreamSliceCmd(clientMock, func() *ExpectedXStreamSlice {
				return clientMock.ExpectXReadGroup(&redis.XReadGroupArgs{
					Group:    "group",
					Consumer: "consumer",
					Streams:  []string{"stream", ">"},
				})
			}, func() *redis.XStreamSliceCmd {
				return client.XReadGroup(ctx, &redis.XReadGroupArgs{
					Group:    "group",
					Consumer: "consumer",
					Streams:  []string{"stream", ">"},
				})
			})
		})

		It("XAck", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectXAck("stream", "group", "1-0", "2-0", "4-0")
			}, func() *redis.IntCmd {
				return client.XAck(ctx, "stream", "group", "1-0", "2-0", "4-0")
			})
		})

		It("XPending", func() {
			operationXPendingCmd(clientMock, func() *ExpectedXPending {
				return clientMock.ExpectXPending("stream", "group")
			}, func() *redis.XPendingCmd {
				return client.XPending(ctx, "stream", "group")
			})
		})

		It("XPendingExt", func() {
			operationXPendingExtCmd(clientMock, func() *ExpectedXPendingExt {
				return clientMock.ExpectXPendingExt(&redis.XPendingExtArgs{
					Stream:   "stream",
					Group:    "group",
					Start:    "-",
					End:      "+",
					Count:    10,
					Consumer: "consumer",
				})
			}, func() *redis.XPendingExtCmd {
				return client.XPendingExt(ctx, &redis.XPendingExtArgs{
					Stream:   "stream",
					Group:    "group",
					Start:    "-",
					End:      "+",
					Count:    10,
					Consumer: "consumer",
				})
			})
		})

		It("XClaim", func() {
			operationXMessageSliceCmd(clientMock, func() *ExpectedXMessageSlice {
				return clientMock.ExpectXClaim(&redis.XClaimArgs{
					Stream:   "stream",
					Group:    "group",
					Consumer: "consumer",
					Messages: []string{"1-0", "2-0", "3-0"},
				})
			}, func() *redis.XMessageSliceCmd {
				return client.XClaim(ctx, &redis.XClaimArgs{
					Stream:   "stream",
					Group:    "group",
					Consumer: "consumer",
					Messages: []string{"1-0", "2-0", "3-0"},
				})
			})
		})

		It("XClaimJustID", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectXClaimJustID(&redis.XClaimArgs{
					Stream:   "stream",
					Group:    "group",
					Consumer: "consumer",
					Messages: []string{"1-0", "2-0", "3-0"},
				})
			}, func() *redis.StringSliceCmd {
				return client.XClaimJustID(ctx, &redis.XClaimArgs{
					Stream:   "stream",
					Group:    "group",
					Consumer: "consumer",
					Messages: []string{"1-0", "2-0", "3-0"},
				})
			})
		})

		It("XTrim", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectXTrim("stream", 0)
			}, func() *redis.IntCmd {
				return client.XTrim(ctx, "stream", 0)
			})
		})

		It("XTrimApprox", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectXTrimApprox("stream", 0)
			}, func() *redis.IntCmd {
				return client.XTrimApprox(ctx, "stream", 0)
			})
		})

		It("XInfoGroups", func() {
			operationXInfoGroupsCmd(clientMock, func() *ExpectedXInfoGroups {
				return clientMock.ExpectXInfoGroups("key")
			}, func() *redis.XInfoGroupsCmd {
				return client.XInfoGroups(ctx, "key")
			})
		})

		It("XInfoStream", func() {
			operationXInfoStreamCmd(clientMock, func() *ExpectedXInfoStream {
				return clientMock.ExpectXInfoStream("key")
			}, func() *redis.XInfoStreamCmd {
				return client.XInfoStream(ctx, "key")
			})
		})

		It("BZPopMax", func() {
			operationZWithKeyCmd(clientMock, func() *ExpectedZWithKey {
				return clientMock.ExpectBZPopMax(0, "zset1", "zset2")
			}, func() *redis.ZWithKeyCmd {
				return client.BZPopMax(ctx, 0, "zset1", "zset2")
			})
		})

		It("BZPopMin", func() {
			operationZWithKeyCmd(clientMock, func() *ExpectedZWithKey {
				return clientMock.ExpectBZPopMin(0, "zset1", "zset2")
			}, func() *redis.ZWithKeyCmd {
				return client.BZPopMin(ctx, 0, "zset1", "zset2")
			})
		})

		It("ZAdd", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZAdd("zset", &redis.Z{
					Member: "a",
					Score:  1,
				})
			}, func() *redis.IntCmd {
				return client.ZAdd(ctx, "zset", &redis.Z{
					Member: "a",
					Score:  1,
				})
			})
		})

		It("ZAddNX", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZAddNX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.IntCmd {
				return client.ZAddNX(ctx, "zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZAddXX", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZAddXX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.IntCmd {
				return client.ZAddXX(ctx, "zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZAddCh", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZAddCh("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.IntCmd {
				return client.ZAddCh(ctx, "zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZAddNXCh", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZAddNXCh("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.IntCmd {
				return client.ZAddNXCh(ctx, "zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZAddXXCh", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZAddXXCh("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.IntCmd {
				return client.ZAddXXCh(ctx, "zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZIncr", func() {
			operationFloatCmd(clientMock, func() *ExpectedFloat {
				return clientMock.ExpectZIncr("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.FloatCmd {
				return client.ZIncr(ctx, "zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZIncrNX", func() {
			operationFloatCmd(clientMock, func() *ExpectedFloat {
				return clientMock.ExpectZIncrNX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.FloatCmd {
				return client.ZIncrNX(ctx, "zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZIncrXX", func() {
			operationFloatCmd(clientMock, func() *ExpectedFloat {
				return clientMock.ExpectZIncrXX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.FloatCmd {
				return client.ZIncrXX(ctx, "zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZCard", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZCard("key")
			}, func() *redis.IntCmd {
				return client.ZCard(ctx, "key")
			})
		})

		It("ZCount", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZCount("zset", "-inf", "+inf")
			}, func() *redis.IntCmd {
				return client.ZCount(ctx, "zset", "-inf", "+inf")
			})
		})

		It("ZLexCount", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZLexCount("zset", "-", "+")
			}, func() *redis.IntCmd {
				return client.ZLexCount(ctx, "zset", "-", "+")
			})
		})

		It("ZIncrBy", func() {
			operationFloatCmd(clientMock, func() *ExpectedFloat {
				return clientMock.ExpectZIncrBy("zset", 2, "one")
			}, func() *redis.FloatCmd {
				return client.ZIncrBy(ctx, "zset", 2, "one")
			})
		})

		It("ZInterStore", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZInterStore("out", &redis.ZStore{
					Keys:    []string{"zset1", "zset2"},
					Weights: []float64{2, 3},
				})
			}, func() *redis.IntCmd {
				return client.ZInterStore(ctx, "out", &redis.ZStore{
					Keys:    []string{"zset1", "zset2"},
					Weights: []float64{2, 3},
				})
			})
		})

		It("ZPopMax", func() {
			operationZSliceCmd(clientMock, func() *ExpectedZSlice {
				return clientMock.ExpectZPopMax("key")
			}, func() *redis.ZSliceCmd {
				return client.ZPopMax(ctx, "key")
			})
		})

		It("ZPopMin", func() {
			operationZSliceCmd(clientMock, func() *ExpectedZSlice {
				return clientMock.ExpectZPopMin("key")
			}, func() *redis.ZSliceCmd {
				return client.ZPopMin(ctx, "key")
			})
		})

		It("ZRange", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectZRange("zset", 0, -1)
			}, func() *redis.StringSliceCmd {
				return client.ZRange(ctx, "zset", 0, -1)
			})
		})

		It("ZRangeWithScores", func() {
			operationZSliceCmd(clientMock, func() *ExpectedZSlice {
				return clientMock.ExpectZRangeWithScores("zset", 0, -1)
			}, func() *redis.ZSliceCmd {
				return client.ZRangeWithScores(ctx, "zset", 0, -1)
			})
		})

		It("ZRangeByScore", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectZRangeByScore("zset", &redis.ZRangeBy{
					Min: "-inf",
					Max: "+inf",
				})
			}, func() *redis.StringSliceCmd {
				return client.ZRangeByScore(ctx, "zset", &redis.ZRangeBy{
					Min: "-inf",
					Max: "+inf",
				})
			})
		})

		It("ZRangeByLex", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectZRangeByLex("zset", &redis.ZRangeBy{
					Min: "-",
					Max: "+",
				})
			}, func() *redis.StringSliceCmd {
				return client.ZRangeByLex(ctx, "zset", &redis.ZRangeBy{
					Min: "-",
					Max: "+",
				})
			})
		})

		It("ZRangeByScoreWithScores", func() {
			operationZSliceCmd(clientMock, func() *ExpectedZSlice {
				return clientMock.ExpectZRangeByScoreWithScores("zset", &redis.ZRangeBy{
					Min: "-inf",
					Max: "+inf",
				})
			}, func() *redis.ZSliceCmd {
				return client.ZRangeByScoreWithScores(ctx, "zset", &redis.ZRangeBy{
					Min: "-inf",
					Max: "+inf",
				})
			})
		})

		It("ZRank", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZRank("zset", "three")
			}, func() *redis.IntCmd {
				return client.ZRank(ctx, "zset", "three")
			})
		})

		It("ZRem", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZRem("zset", "two")
			}, func() *redis.IntCmd {
				return client.ZRem(ctx, "zset", "two")
			})
		})

		It("ZRemRangeByRank", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZRemRangeByRank("key", 1, 2)
			}, func() *redis.IntCmd {
				return client.ZRemRangeByRank(ctx, "key", 1, 2)
			})
		})

		It("ZRemRangeByScore", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZRemRangeByScore("zset", "-inf", "(2")
			}, func() *redis.IntCmd {
				return client.ZRemRangeByScore(ctx, "zset", "-inf", "(2")
			})
		})

		It("ZRemRangeByLex", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZRemRangeByLex("zset", "[alpha", "[omega")
			}, func() *redis.IntCmd {
				return client.ZRemRangeByLex(ctx, "zset", "[alpha", "[omega")
			})
		})

		It("ZRevRange", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectZRevRange("zset", 0, -1)
			}, func() *redis.StringSliceCmd {
				return client.ZRevRange(ctx, "zset", 0, -1)
			})
		})

		It("ZRevRangeWithScores", func() {
			operationZSliceCmd(clientMock, func() *ExpectedZSlice {
				return clientMock.ExpectZRevRangeWithScores("zset", 0, -1)
			}, func() *redis.ZSliceCmd {
				return client.ZRevRangeWithScores(ctx, "zset", 0, -1)
			})
		})

		It("ZRevRangeByScore", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectZRevRangeByScore("zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			}, func() *redis.StringSliceCmd {
				return client.ZRevRangeByScore(ctx, "zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			})
		})

		It("ZRevRangeByLex", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectZRevRangeByLex("zset", &redis.ZRangeBy{Max: "+", Min: "-"})
			}, func() *redis.StringSliceCmd {
				return client.ZRevRangeByLex(ctx, "zset", &redis.ZRangeBy{Max: "+", Min: "-"})
			})
		})

		It("ZRevRangeByScoreWithScores", func() {
			operationZSliceCmd(clientMock, func() *ExpectedZSlice {
				return clientMock.ExpectZRevRangeByScoreWithScores("zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			}, func() *redis.ZSliceCmd {
				return client.ZRevRangeByScoreWithScores(ctx, "zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			})
		})

		It("ZRevRank", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZRevRank("key", "member")
			}, func() *redis.IntCmd {
				return client.ZRevRank(ctx, "key", "member")
			})
		})

		It("ZScore", func() {
			operationFloatCmd(clientMock, func() *ExpectedFloat {
				return clientMock.ExpectZScore("key", "member")
			}, func() *redis.FloatCmd {
				return client.ZScore(ctx, "key", "member")
			})
		})

		It("ZUnionWithScores", func() {
			operationZSliceCmd(clientMock, func() *ExpectedZSlice {
				return clientMock.ExpectZUnionWithScores(redis.ZStore{
					Keys:    []string{"zset1", "zset2"},
					Weights: []float64{2, 3},
				})
			}, func() *redis.ZSliceCmd {
				return client.ZUnionWithScores(ctx, redis.ZStore{
					Keys:    []string{"zset1", "zset2"},
					Weights: []float64{2, 3},
				})
			})
		})

		It("ZUnionStore", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectZUnionStore("out", &redis.ZStore{
					Keys:    []string{"zset1", "zset2"},
					Weights: []float64{2, 3},
				})
			}, func() *redis.IntCmd {
				return client.ZUnionStore(ctx, "out", &redis.ZStore{
					Keys:    []string{"zset1", "zset2"},
					Weights: []float64{2, 3},
				})
			})
		})

		It("PFAdd", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectPFAdd("hll1", "1", "2", "3", "4", "5")
			}, func() *redis.IntCmd {
				return client.PFAdd(ctx, "hll1", "1", "2", "3", "4", "5")
			})
		})

		It("PFCount", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectPFCount("hll1", "hll2")
			}, func() *redis.IntCmd {
				return client.PFCount(ctx, "hll1", "hll2")
			})
		})

		It("PFMerge", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectPFMerge("hllMerged", "hll1", "hll2")
			}, func() *redis.StatusCmd {
				return client.PFMerge(ctx, "hllMerged", "hll1", "hll2")
			})
		})

		It("BgRewriteAOF", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectBgRewriteAOF()
			}, func() *redis.StatusCmd {
				return client.BgRewriteAOF(ctx)
			})
		})

		It("BgSave", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectBgSave()
			}, func() *redis.StatusCmd {
				return client.BgSave(ctx)
			})
		})

		It("ClientKill", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClientKill("1.1.1.1:1111")
			}, func() *redis.StatusCmd {
				return client.ClientKill(ctx, "1.1.1.1:1111")
			})
		})

		It("ClientKillByFilter", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectClientKillByFilter("11.11.11.11:1234")
			}, func() *redis.IntCmd {
				return client.ClientKillByFilter(ctx, "11.11.11.11:1234")
			})
		})

		It("ClientList", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectClientList()
			}, func() *redis.StringCmd {
				return client.ClientList(ctx)
			})
		})

		It("ClientPause", func() {
			operationBoolCmd(clientMock, func() *ExpectedBool {
				return clientMock.ExpectClientPause(1 * time.Minute)
			}, func() *redis.BoolCmd {
				return client.ClientPause(ctx, 1*time.Minute)
			})
		})

		It("ClientID", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectClientID()
			}, func() *redis.IntCmd {
				return client.ClientID(ctx)
			})
		})

		It("ConfigGet", func() {
			operationSliceCmd(clientMock, func() *ExpectedSlice {
				return clientMock.ExpectConfigGet("*")
			}, func() *redis.SliceCmd {
				return client.ConfigGet(ctx, "*")
			})
		})

		It("ConfigResetStat", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectConfigResetStat()
			}, func() *redis.StatusCmd {
				return client.ConfigResetStat(ctx)
			})
		})

		It("ConfigSet", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectConfigSet("maxmemory", "1024M")
			}, func() *redis.StatusCmd {
				return client.ConfigSet(ctx, "maxmemory", "1024M")
			})
		})

		It("ConfigRewrite", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectConfigRewrite()
			}, func() *redis.StatusCmd {
				return client.ConfigRewrite(ctx)
			})
		})

		It("DBSize", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectDBSize()
			}, func() *redis.IntCmd {
				return client.DBSize(ctx)
			})
		})

		It("FlushAll", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectFlushAll()
			}, func() *redis.StatusCmd {
				return client.FlushAll(ctx)
			})
		})

		It("FlushAllAsync", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectFlushAllAsync()
			}, func() *redis.StatusCmd {
				return client.FlushAllAsync(ctx)
			})
		})

		It("FlushDB", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectFlushDB()
			}, func() *redis.StatusCmd {
				return client.FlushDB(ctx)
			})
		})

		It("FlushDBAsync", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectFlushDBAsync()
			}, func() *redis.StatusCmd {
				return client.FlushDBAsync(ctx)
			})
		})

		It("Info", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectInfo()
			}, func() *redis.StringCmd {
				return client.Info(ctx)
			})
		})

		It("LastSave", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectLastSave()
			}, func() *redis.IntCmd {
				return client.LastSave(ctx)
			})
		})

		It("Save", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectSave()
			}, func() *redis.StatusCmd {
				return client.Save(ctx)
			})
		})

		It("Shutdown", func() {
			//no test
		})

		It("ShutdownSave", func() {
			//no test
		})

		It("ShutdownNoSave", func() {
			//no test
		})

		It("SlaveOf", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectSlaveOf("localhost", "8888")
			}, func() *redis.StatusCmd {
				return client.SlaveOf(ctx, "localhost", "8888")
			})
		})

		It("Time", func() {
			operationTimeCmd(clientMock, func() *ExpectedTime {
				return clientMock.ExpectTime()
			}, func() *redis.TimeCmd {
				return client.Time(ctx)
			})
		})

		It("DebugObject", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectDebugObject("foo")
			}, func() *redis.StringCmd {
				return client.DebugObject(ctx, "foo")
			})
		})

		It("ReadOnly", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectReadOnly()
			}, func() *redis.StatusCmd {
				return client.ReadOnly(ctx)
			})
		})

		It("ReadWrite", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectReadWrite()
			}, func() *redis.StatusCmd {
				return client.ReadWrite(ctx)
			})
		})

		It("MemoryUsage", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectMemoryUsage("foo")
			}, func() *redis.IntCmd {
				return client.MemoryUsage(ctx, "foo")
			})
		})

		It("Eval", func() {
			operationCmdCmd(clientMock, func() *ExpectedCmd {
				return clientMock.ExpectEval("return {KEYS[1],ARGV[1]}", []string{"key"}, "hello")
			}, func() *redis.Cmd {
				return client.Eval(ctx, "return {KEYS[1],ARGV[1]}", []string{"key"}, "hello")
			})
		})

		It("EvalSha", func() {
			operationCmdCmd(clientMock, func() *ExpectedCmd {
				return clientMock.ExpectEvalSha("sha", []string{"key1", "key2"}, "args1", "args2")
			}, func() *redis.Cmd {
				return client.EvalSha(ctx, "sha", []string{"key1", "key2"}, "args1", "args2")
			})
		})

		It("ScriptExists", func() {
			operationBoolSliceCmd(clientMock, func() *ExpectedBoolSlice {
				return clientMock.ExpectScriptExists()
			}, func() *redis.BoolSliceCmd {
				return client.ScriptExists(ctx)
			})
		})

		It("ScriptFlush", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectScriptFlush()
			}, func() *redis.StatusCmd {
				return client.ScriptFlush(ctx)
			})
		})

		It("ScriptKill", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectScriptKill()
			}, func() *redis.StatusCmd {
				return client.ScriptKill(ctx)
			})
		})

		It("ScriptLoad", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectScriptLoad("script")
			}, func() *redis.StringCmd {
				return client.ScriptLoad(ctx, "script")
			})
		})

		It("Publish", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectPublish("channel", "message")
			}, func() *redis.IntCmd {
				return client.Publish(ctx, "channel", "message")
			})
		})

		It("PubSubChannels", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectPubSubChannels("pattern")
			}, func() *redis.StringSliceCmd {
				return client.PubSubChannels(ctx, "pattern")
			})
		})

		It("PubSubNumSub", func() {
			operationStringIntMapCmd(clientMock, func() *ExpectedStringIntMap {
				return clientMock.ExpectPubSubNumSub()
			}, func() *redis.StringIntMapCmd {
				return client.PubSubNumSub(ctx)
			})
		})

		It("PubSubNumPat", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectPubSubNumPat()
			}, func() *redis.IntCmd {
				return client.PubSubNumPat(ctx)
			})
		})

		It("ClusterSlots", func() {
			operationClusterSlotsCmd(clientMock, func() *ExpectedClusterSlots {
				return clientMock.ExpectClusterSlots()
			}, func() *redis.ClusterSlotsCmd {
				return client.ClusterSlots(ctx)
			})
		})

		It("ClusterNodes", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectClusterNodes()
			}, func() *redis.StringCmd {
				return client.ClusterNodes(ctx)
			})
		})

		It("ClusterMeet", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterMeet("1.1.1.1", "1")
			}, func() *redis.StatusCmd {
				return client.ClusterMeet(ctx, "1.1.1.1", "1")
			})
		})

		It("ClusterForget", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterForget("id")
			}, func() *redis.StatusCmd {
				return client.ClusterForget(ctx, "id")
			})
		})

		It("ClusterReplicate", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterReplicate("id")
			}, func() *redis.StatusCmd {
				return client.ClusterReplicate(ctx, "id")
			})
		})

		It("ClusterResetSoft", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterResetSoft()
			}, func() *redis.StatusCmd {
				return client.ClusterResetSoft(ctx)
			})
		})

		It("ClusterResetHard", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterResetHard()
			}, func() *redis.StatusCmd {
				return client.ClusterResetHard(ctx)
			})
		})

		It("ClusterInfo", func() {
			operationStringCmd(clientMock, func() *ExpectedString {
				return clientMock.ExpectClusterInfo()
			}, func() *redis.StringCmd {
				return client.ClusterInfo(ctx)
			})
		})

		It("ClusterKeySlot", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectClusterKeySlot("key")
			}, func() *redis.IntCmd {
				return client.ClusterKeySlot(ctx, "key")
			})
		})

		It("ClusterGetKeysInSlot", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectClusterGetKeysInSlot(1, 2)
			}, func() *redis.StringSliceCmd {
				return client.ClusterGetKeysInSlot(ctx, 1, 2)
			})
		})

		It("ClusterCountFailureReports", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectClusterCountFailureReports("id")
			}, func() *redis.IntCmd {
				return client.ClusterCountFailureReports(ctx, "id")
			})
		})

		It("ClusterCountKeysInSlot", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectClusterCountKeysInSlot(1)
			}, func() *redis.IntCmd {
				return client.ClusterCountKeysInSlot(ctx, 1)
			})
		})

		It("ClusterDelSlots", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterDelSlots()
			}, func() *redis.StatusCmd {
				return client.ClusterDelSlots(ctx)
			})
		})

		It("ClusterDelSlotsRange", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterDelSlotsRange(1, 2)
			}, func() *redis.StatusCmd {
				return client.ClusterDelSlotsRange(ctx, 1, 2)
			})
		})

		It("ClusterSaveConfig", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterSaveConfig()
			}, func() *redis.StatusCmd {
				return client.ClusterSaveConfig(ctx)
			})
		})

		It("ClusterSlaves", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectClusterSlaves("id")
			}, func() *redis.StringSliceCmd {
				return client.ClusterSlaves(ctx, "id")
			})
		})

		It("ClusterFailover", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterFailover()
			}, func() *redis.StatusCmd {
				return client.ClusterFailover(ctx)
			})
		})

		It("ClusterAddSlots", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterAddSlots()
			}, func() *redis.StatusCmd {
				return client.ClusterAddSlots(ctx)
			})
		})

		It("ClusterAddSlotsRange", func() {
			operationStatusCmd(clientMock, func() *ExpectedStatus {
				return clientMock.ExpectClusterAddSlotsRange(1, 2)
			}, func() *redis.StatusCmd {
				return client.ClusterAddSlotsRange(ctx, 1, 2)
			})
		})

		It("GeoAdd", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectGeoAdd("Sicily",
					&redis.GeoLocation{Longitude: 13.361389, Latitude: 38.115556, Name: "Palermo"},
					&redis.GeoLocation{Longitude: 15.087269, Latitude: 37.502669, Name: "Tokyo"},
				)
			}, func() *redis.IntCmd {
				return client.GeoAdd(ctx, "Sicily",
					&redis.GeoLocation{Longitude: 13.361389, Latitude: 38.115556, Name: "Palermo"},
					&redis.GeoLocation{Longitude: 15.087269, Latitude: 37.502669, Name: "Tokyo"})
			})
		})

		It("GeoPos", func() {
			operationGeoPosCmd(clientMock, func() *ExpectedGeoPos {
				return clientMock.ExpectGeoPos("Sicily", "Palermo", "Catania", "NonExisting")
			}, func() *redis.GeoPosCmd {
				return client.GeoPos(ctx, "Sicily", "Palermo", "Catania", "NonExisting")
			})
		})

		It("GeoRadius", func() {
			operationGeoLocationCmd(clientMock, func() *ExpectedGeoLocation {
				return clientMock.ExpectGeoRadius("Sicily", 15, 37, &redis.GeoRadiusQuery{
					Radius:      200,
					Unit:        "km",
					WithGeoHash: true,
					WithCoord:   true,
					WithDist:    true,
					Count:       2,
					Sort:        "ASC",
				})
			}, func() *redis.GeoLocationCmd {
				return client.GeoRadius(ctx, "Sicily", 15, 37, &redis.GeoRadiusQuery{
					Radius:      200,
					Unit:        "km",
					WithGeoHash: true,
					WithCoord:   true,
					WithDist:    true,
					Count:       2,
					Sort:        "ASC",
				})
			})
		})

		It("GeoRadiusStore", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectGeoRadiusStore("Sicily", 15, 37, &redis.GeoRadiusQuery{
					Radius: 200,
					Store:  "result",
				})
			}, func() *redis.IntCmd {
				return client.GeoRadiusStore(ctx, "Sicily", 15, 37, &redis.GeoRadiusQuery{
					Radius: 200,
					Store:  "result",
				})
			})
		})

		It("GeoRadiusByMember", func() {
			operationGeoLocationCmd(clientMock, func() *ExpectedGeoLocation {
				return clientMock.ExpectGeoRadiusByMember("Sicily", "Catania", &redis.GeoRadiusQuery{
					Radius:      200,
					Unit:        "km",
					WithGeoHash: true,
					WithCoord:   true,
					WithDist:    true,
					Count:       2,
					Sort:        "ASC",
				})
			}, func() *redis.GeoLocationCmd {
				return client.GeoRadiusByMember(ctx, "Sicily", "Catania", &redis.GeoRadiusQuery{
					Radius:      200,
					Unit:        "km",
					WithGeoHash: true,
					WithCoord:   true,
					WithDist:    true,
					Count:       2,
					Sort:        "ASC",
				})
			})
		})

		It("GeoRadiusByMemberStore", func() {
			operationIntCmd(clientMock, func() *ExpectedInt {
				return clientMock.ExpectGeoRadiusByMemberStore("key", "member", &redis.GeoRadiusQuery{
					Radius:      1,
					Unit:        "unit",
					WithCoord:   true,
					WithDist:    true,
					WithGeoHash: true,
					Count:       10,
					Sort:        "desc",
					Store:       "data",
					StoreDist:   "dist",
				})
			}, func() *redis.IntCmd {
				return client.GeoRadiusByMemberStore(ctx, "key", "member", &redis.GeoRadiusQuery{
					Radius:      1,
					Unit:        "unit",
					WithCoord:   true,
					WithDist:    true,
					WithGeoHash: true,
					Count:       10,
					Sort:        "desc",
					Store:       "data",
					StoreDist:   "dist",
				})
			})
		})

		It("GeoDist", func() {
			operationFloatCmd(clientMock, func() *ExpectedFloat {
				return clientMock.ExpectGeoDist("Sicily", "Palermo", "Catania", "km")
			}, func() *redis.FloatCmd {
				return client.GeoDist(ctx, "Sicily", "Palermo", "Catania", "km")
			})
		})

		It("GeoHash", func() {
			operationStringSliceCmd(clientMock, func() *ExpectedStringSlice {
				return clientMock.ExpectGeoHash("Sicily", "Palermo", "Catania")
			}, func() *redis.StringSliceCmd {
				return client.GeoHash(ctx, "Sicily", "Palermo", "Catania")
			})
		})
	})
})
