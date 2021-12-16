package redismock

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RedisMock", func() {
	var (
		client      *redis.ClusterClient
		clusterMock ClusterClientMock
		disorder    func() map[string]interface{}
	)

	BeforeEach(func() {
		client, clusterMock = NewClusterMock()
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
		Expect(clusterMock.ExpectationsWereMet()).NotTo(HaveOccurred())
	})

	//Describe("tx pipeline", func() {
	//	var pipe redis.Pipeliner
	//
	//	BeforeEach(func() {
	//		clusterMock.ExpectTxPipeline()
	//		clusterMock.ExpectGet("key1").SetVal("pipeline get")
	//		clusterMock.ExpectHGet("hash_key", "hash_field").SetVal("pipeline hash get")
	//		clusterMock.ExpectSet("set_key", "set value", 1*time.Minute).SetVal("OK")
	//		clusterMock.ExpectTxPipelineExec()
	//
	//		pipe = client.TxPipeline()
	//	})
	//
	//	It("tx pipeline order", func() {
	//		get := pipe.Get(ctx, "key1")
	//		hashGet := pipe.HGet(ctx, "hash_key", "hash_field")
	//		set := pipe.Set(ctx, "set_key", "set value", 1*time.Minute)
	//
	//		_, err := pipe.Exec(ctx)
	//		Expect(err).NotTo(HaveOccurred())
	//
	//		Expect(get.Err()).NotTo(HaveOccurred())
	//		Expect(get.Val()).To(Equal("pipeline get"))
	//
	//		Expect(hashGet.Err()).NotTo(HaveOccurred())
	//		Expect(hashGet.Val()).To(Equal("pipeline hash get"))
	//
	//		Expect(set.Err()).NotTo(HaveOccurred())
	//		Expect(set.Val()).To(Equal("OK"))
	//	})
	//
	//	It("tx pipeline not order", func() {
	//		clusterMock.MatchExpectationsInOrder(false)
	//
	//		hashGet := pipe.HGet(ctx, "hash_key", "hash_field")
	//		set := pipe.Set(ctx, "set_key", "set value", 1*time.Minute)
	//		get := pipe.Get(ctx, "key1")
	//
	//		_, err := pipe.Exec(ctx)
	//		Expect(err).NotTo(HaveOccurred())
	//
	//		Expect(get.Err()).NotTo(HaveOccurred())
	//		Expect(get.Val()).To(Equal("pipeline get"))
	//
	//		Expect(hashGet.Err()).NotTo(HaveOccurred())
	//		Expect(hashGet.Val()).To(Equal("pipeline hash get"))
	//
	//		Expect(set.Err()).NotTo(HaveOccurred())
	//		Expect(set.Val()).To(Equal("OK"))
	//	})
	//})

	//Describe("pipeline", func() {
	//	var pipe redis.Pipeliner
	//
	//	BeforeEach(func() {
	//		clusterMock.ExpectGet("key1").SetVal("pipeline get")
	//		clusterMock.ExpectHGet("hash_key", "hash_field").SetVal("pipeline hash get")
	//		clusterMock.ExpectSet("set_key", "set value", 1*time.Minute).SetVal("OK")
	//
	//		pipe = client.Pipeline()
	//	})
	//
	//	It("pipeline order", func() {
	//		clusterMock.MatchExpectationsInOrder(true)
	//
	//		get := pipe.Get(ctx, "key1")
	//		hashGet := pipe.HGet(ctx, "hash_key", "hash_field")
	//		set := pipe.Set(ctx, "set_key", "set value", 1*time.Minute)
	//
	//		_, err := pipe.Exec(ctx)
	//		Expect(err).NotTo(HaveOccurred())
	//
	//		Expect(get.Err()).NotTo(HaveOccurred())
	//		Expect(get.Val()).To(Equal("pipeline get"))
	//
	//		Expect(hashGet.Err()).NotTo(HaveOccurred())
	//		Expect(hashGet.Val()).To(Equal("pipeline hash get"))
	//
	//		Expect(set.Err()).NotTo(HaveOccurred())
	//		Expect(set.Val()).To(Equal("OK"))
	//	})
	//
	//	It("pipeline not order", func() {
	//		clusterMock.MatchExpectationsInOrder(false)
	//
	//		hashGet := pipe.HGet(ctx, "hash_key", "hash_field")
	//		set := pipe.Set(ctx, "set_key", "set value", 1*time.Minute)
	//		get := pipe.Get(ctx, "key1")
	//
	//		_, err := pipe.Exec(ctx)
	//		Expect(err).NotTo(HaveOccurred())
	//
	//		Expect(get.Err()).NotTo(HaveOccurred())
	//		Expect(get.Val()).To(Equal("pipeline get"))
	//
	//		Expect(hashGet.Err()).NotTo(HaveOccurred())
	//		Expect(hashGet.Val()).To(Equal("pipeline hash get"))
	//
	//		Expect(set.Err()).NotTo(HaveOccurred())
	//		Expect(set.Val()).To(Equal("OK"))
	//	})
	//})

	//Describe("watch", func() {
	//	BeforeEach(func() {
	//		clusterMock.ExpectWatch("key1", "key2")
	//		clusterMock.ExpectGet("key1").SetVal("1")
	//		clusterMock.ExpectSet("key2", "2", 1*time.Second).SetVal("OK")
	//	})
	//
	//	It("watch error", func() {
	//		clusterMock.MatchExpectationsInOrder(false)
	//		txf := func(tx *redis.Tx) error {
	//			_ = tx.Get(ctx, "key1")
	//			_ = tx.Set(ctx, "key2", "2", 1*time.Second)
	//			return errors.New("watch tx error")
	//		}
	//
	//		err := client.Watch(ctx, txf, "key1", "key2")
	//		Expect(err).To(Equal(errors.New("watch tx error")))
	//
	//		clusterMock.ExpectWatch("key3", "key4").SetErr(errors.New("watch error"))
	//		txf = func(tx *redis.Tx) error {
	//			return nil
	//		}
	//
	//		err = client.Watch(ctx, txf, "key3", "key4")
	//		Expect(err).To(Equal(errors.New("watch error")))
	//	})
	//
	//	It("watch in order", func() {
	//		clusterMock.MatchExpectationsInOrder(true)
	//		txf := func(tx *redis.Tx) error {
	//			val, err := tx.Get(ctx, "key1").Int64()
	//			if err != nil {
	//				return err
	//			}
	//			Expect(val).To(Equal(int64(1)))
	//			err = tx.Set(ctx, "key2", "2", 1*time.Second).Err()
	//			if err != nil {
	//				return err
	//			}
	//			return nil
	//		}
	//
	//		err := client.Watch(ctx, txf, "key1", "key2")
	//		Expect(err).NotTo(HaveOccurred())
	//	})
	//
	//	It("watch out of order", func() {
	//		clusterMock.MatchExpectationsInOrder(false)
	//		txf := func(tx *redis.Tx) error {
	//			err := tx.Set(ctx, "key2", "2", 1*time.Second).Err()
	//			if err != nil {
	//				return err
	//			}
	//			val, err := tx.Get(ctx, "key1").Int64()
	//			if err != nil {
	//				return err
	//			}
	//			Expect(val).To(Equal(int64(1)))
	//			return nil
	//		}
	//
	//		err := client.Watch(ctx, txf, "key1", "key2")
	//		Expect(err).NotTo(HaveOccurred())
	//	})
	//})

	Describe("work order", func() {
		BeforeEach(func() {
			clusterMock.ExpectGet("key").RedisNil()
			clusterMock.ExpectSet("key", "1", 1*time.Second).SetVal("OK")
			clusterMock.ExpectGet("key").SetVal("1")
			clusterMock.ExpectGetSet("key", "0").SetVal("1")
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

			Expect(clusterMock.ExpectationsWereMet()).To(HaveOccurred())

			_ = client.Get(ctx, "key")
			Expect(clusterMock.ExpectationsWereMet()).To(HaveOccurred())

			_ = client.GetSet(ctx, "key", "0")
		})

		It("not enough", func() {
			_ = client.Get(ctx, "key")
			_ = client.Set(ctx, "key", "1", 1*time.Second)
			_ = client.Get(ctx, "key")
			_ = client.GetSet(ctx, "key", "0")
			Expect(clusterMock.ExpectationsWereMet()).NotTo(HaveOccurred())

			get := client.HGet(ctx, "key", "field")
			Expect(get.Err()).To(HaveOccurred())
			Expect(get.Val()).To(Equal(""))
		})
	})

	Describe("work not order", func() {

		BeforeEach(func() {
			clusterMock.MatchExpectationsInOrder(false)

			clusterMock.ExpectSet("key", "1", 1*time.Second).SetVal("OK")
			clusterMock.ExpectGet("key").SetVal("1")
			clusterMock.ExpectGetSet("key", "0").SetVal("1")
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
			clusterMock.Regexp().ExpectSet("key", `^order_id_[0-9]{10}$`, 1*time.Second).SetVal("OK")
			clusterMock.Regexp().ExpectSet("key2", `^order_id_[0-9]{4}\-[0-9]{2}\-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}.+$`, 1*time.Second).SetVal("OK")

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
			clusterMock.CustomMatch(func(expected, actual []interface{}) error {
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
			Expect(clusterMock.ExpectationsWereMet()).To(HaveOccurred())

			// let AfterEach pass
			clusterMock.ClearExpect()
		})

	})

	Describe("work error", func() {

		It("set error", func() {
			clusterMock.ExpectGet("key").SetErr(errors.New("set error"))

			get := client.Get(ctx, "key")
			Expect(get.Err()).To(Equal(errors.New("set error")))
			Expect(get.Val()).To(Equal(""))
		})

		It("not set", func() {
			clusterMock.ExpectGet("key")

			get := client.Get(ctx, "key")
			Expect(get.Err()).To(HaveOccurred())
			Expect(get.Val()).To(Equal(""))
		})

		It("set zero", func() {
			clusterMock.ExpectGet("key").SetVal("")

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
			clusterMock.ExpectCommand().SetVal(commandsInfo)

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
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectClientGetName()
			}, func() *redis.StringCmd {
				return client.ClientGetName(ctx)
			})
		})

		It("Echo", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectEcho("mock")
			}, func() *redis.StringCmd {
				return client.Echo(ctx, "mock")
			})
		})

		It("Ping", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectPing()
			}, func() *redis.StatusCmd {
				return client.Ping(ctx)
			})
		})

		It("Quit", func() {
			//not implemented
		})

		It("Del", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectDel()
			}, func() *redis.IntCmd {
				return client.Del(ctx)
			})
		})

		It("Unlink", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectUnlink()
			}, func() *redis.IntCmd {
				return client.Unlink(ctx)
			})
		})

		It("Dump", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectDump("key")
			}, func() *redis.StringCmd {
				return client.Dump(ctx, "key")
			})
		})

		It("Exists", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectExists()
			}, func() *redis.IntCmd {
				return client.Exists(ctx)
			})
		})

		It("Expire", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectExpire("key", 1*time.Second)
			}, func() *redis.BoolCmd {
				return client.Expire(ctx, "key", 1*time.Second)
			})
		})

		It("ExpireAt", func() {
			now := time.Now()
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectExpireAt("key", now.Add(20*time.Minute))
			}, func() *redis.BoolCmd {
				return client.ExpireAt(ctx, "key", now.Add(20*time.Minute))
			})
		})

		It("Keys", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectKeys("key")
			}, func() *redis.StringSliceCmd {
				return client.Keys(ctx, "key")
			})
		})

		It("Migrate", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectMigrate("host", "port", "key", 1, 1*time.Hour)
			}, func() *redis.StatusCmd {
				return client.Migrate(ctx, "host", "port", "key", 1, 1*time.Hour)
			})
		})

		It("Move", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectMove("key", 1)
			}, func() *redis.BoolCmd {
				return client.Move(ctx, "key", 1)
			})
		})

		It("ObjectRefCount", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectObjectRefCount("key")
			}, func() *redis.IntCmd {
				return client.ObjectRefCount(ctx, "key")
			})
		})

		It("ObjectEncoding", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectObjectEncoding("key")
			}, func() *redis.StringCmd {
				return client.ObjectEncoding(ctx, "key")
			})
		})

		It("ObjectIdleTime", func() {
			operationDurationCmd(clusterMock, func() *ExpectedDuration {
				return clusterMock.ExpectObjectIdleTime("key")
			}, func() *redis.DurationCmd {
				return client.ObjectIdleTime(ctx, "key")
			})
		})

		It("Persist", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectPersist("key")
			}, func() *redis.BoolCmd {
				return client.Persist(ctx, "key")
			})
		})

		It("PExpire", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectPExpire("key", 1*time.Minute)
			}, func() *redis.BoolCmd {
				return client.PExpire(ctx, "key", 1*time.Minute)
			})
		})

		It("PExpireAt", func() {
			now := time.Now()
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectPExpireAt("key", now.Add(10*time.Minute))
			}, func() *redis.BoolCmd {
				return client.PExpireAt(ctx, "key", now.Add(10*time.Minute))
			})
		})

		It("PTTL", func() {
			operationDurationCmd(clusterMock, func() *ExpectedDuration {
				return clusterMock.ExpectPTTL("key")
			}, func() *redis.DurationCmd {
				return client.PTTL(ctx, "key")
			})
		})

		It("RandomKey", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectRandomKey()
			}, func() *redis.StringCmd {
				return client.RandomKey(ctx)
			})
		})

		It("Rename", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectRename("key", "new_key")
			}, func() *redis.StatusCmd {
				return client.Rename(ctx, "key", "new_key")
			})
		})

		It("RenameNX", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectRenameNX("key", "new_key")
			}, func() *redis.BoolCmd {
				return client.RenameNX(ctx, "key", "new_key")
			})
		})

		It("Restore", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectRestore("key", 1*time.Minute, "value")
			}, func() *redis.StatusCmd {
				return client.Restore(ctx, "key", 1*time.Minute, "value")
			})
		})

		It("RestoreReplace", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectRestoreReplace("key", 1*time.Minute, "value")
			}, func() *redis.StatusCmd {
				return client.RestoreReplace(ctx, "key", 1*time.Minute, "value")
			})
		})

		It("Sort", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectSort("key", &redis.Sort{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectSortStore("key", "store", &redis.Sort{
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
			operationSliceCmd(clusterMock, func() *ExpectedSlice {
				return clusterMock.ExpectSortInterfaces("key", &redis.Sort{
					Get: []string{"object_*"},
				})
			}, func() *redis.SliceCmd {
				return client.SortInterfaces(ctx, "key", &redis.Sort{
					Get: []string{"object_*"},
				})
			})
		})

		It("Touch", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectTouch()
			}, func() *redis.IntCmd {
				return client.Touch(ctx)
			})
		})

		It("TTL", func() {
			operationDurationCmd(clusterMock, func() *ExpectedDuration {
				return clusterMock.ExpectTTL("key")
			}, func() *redis.DurationCmd {
				return client.TTL(ctx, "key")
			})
		})

		It("Type", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectType("key")
			}, func() *redis.StatusCmd {
				return client.Type(ctx, "key")
			})
		})

		It("Append", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectAppend("key", "value")
			}, func() *redis.IntCmd {
				return client.Append(ctx, "key", "value")
			})
		})

		It("Decr", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectDecr("key")
			}, func() *redis.IntCmd {
				return client.Decr(ctx, "key")
			})
		})

		It("DecrBy", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectDecrBy("key", 1)
			}, func() *redis.IntCmd {
				return client.DecrBy(ctx, "key", 1)
			})
		})

		It("Get", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectGet("key")
			}, func() *redis.StringCmd {
				return client.Get(ctx, "key")
			})
		})

		It("GetRange", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectGetRange("key", 1, 10)
			}, func() *redis.StringCmd {
				return client.GetRange(ctx, "key", 1, 10)
			})
		})

		It("GetSet", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectGetSet("key", 1)
			}, func() *redis.StringCmd {
				return client.GetSet(ctx, "key", 1)
			})
		})

		It("Incr", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectIncr("key")
			}, func() *redis.IntCmd {
				return client.Incr(ctx, "key")
			})
		})

		It("IncrBy", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectIncrBy("key", 1)
			}, func() *redis.IntCmd {
				return client.IncrBy(ctx, "key", 1)
			})
		})

		It("IncrByFloat", func() {
			operationFloatCmd(clusterMock, func() *ExpectedFloat {
				return clusterMock.ExpectIncrByFloat("key", 1)
			}, func() *redis.FloatCmd {
				return client.IncrByFloat(ctx, "key", 1)
			})
		})

		It("MGet", func() {
			operationSliceCmd(clusterMock, func() *ExpectedSlice {
				return clusterMock.ExpectMGet()
			}, func() *redis.SliceCmd {
				return client.MGet(ctx)
			})
		})

		It("MSet", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectMSet()
			}, func() *redis.StatusCmd {
				return client.MSet(ctx)
			})
		})

		It("MSet Map", func() {
			clusterMock.ExpectMSet(disorder()).SetVal("OK")
			res, err := client.MSet(ctx, disorder()).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("OK"))

			clusterMock.ExpectMSet("key1", "value1", "key2", "value2").SetVal("OK")
			res, err = client.MSet(ctx, "key2", "value2", "key1", "value1").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("OK"))
		})

		It("MSetNX", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectMSetNX()
			}, func() *redis.BoolCmd {
				return client.MSetNX(ctx)
			})
		})

		It("MSetNX Map", func() {
			clusterMock.ExpectMSetNX(disorder()).SetVal(true)
			res, err := client.MSetNX(ctx, disorder()).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeTrue())

			clusterMock.ExpectMSetNX("key1", "value1", "key2", "value2").SetVal(true)
			res, err = client.MSetNX(ctx, "key2", "value2", "key1", "value1").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeTrue())
		})

		It("Set", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectSet("key", "value", 1*time.Minute)
			}, func() *redis.StatusCmd {
				return client.Set(ctx, "key", "value", 1*time.Minute)
			})
		})

		It("SetEX", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectSetEX("key", "value", 1*time.Minute)
			}, func() *redis.StatusCmd {
				return client.SetEX(ctx, "key", "value", 1*time.Minute)
			})
		})

		It("SetNX", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectSetNX("key", "value", 1*time.Minute)
			}, func() *redis.BoolCmd {
				return client.SetNX(ctx, "key", "value", 1*time.Minute)
			})
		})

		It("SetXX", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectSetXX("key", "value", 1*time.Minute)
			}, func() *redis.BoolCmd {
				return client.SetXX(ctx, "key", "value", 1*time.Minute)
			})
		})

		It("SetRange", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectSetRange("key", 1, "value")
			}, func() *redis.IntCmd {
				return client.SetRange(ctx, "key", 1, "value")
			})
		})

		It("StrLen", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectStrLen("key")
			}, func() *redis.IntCmd {
				return client.StrLen(ctx, "key")
			})
		})

		It("GetBit", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectGetBit("key", 1)
			}, func() *redis.IntCmd {
				return client.GetBit(ctx, "key", 1)
			})
		})

		It("SetBit", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectSetBit("key", 1, 2)
			}, func() *redis.IntCmd {
				return client.SetBit(ctx, "key", 1, 2)
			})
		})

		It("BitCount", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectBitCount("key", &redis.BitCount{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectBitOpAnd("dest", "key1", "key2", "key3")
			}, func() *redis.IntCmd {
				return client.BitOpAnd(ctx, "dest", "key1", "key2", "key3")
			})
		})

		It("BitOpOr", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectBitOpOr("dest", "key1", "key2", "key3")
			}, func() *redis.IntCmd {
				return client.BitOpOr(ctx, "dest", "key1", "key2", "key3")
			})
		})

		It("BitOpXor", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectBitOpXor("dest", "key1", "key2", "key3")
			}, func() *redis.IntCmd {
				return client.BitOpXor(ctx, "dest", "key1", "key2", "key3")
			})
		})

		It("BitOpNot", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectBitOpNot("dest", "key")
			}, func() *redis.IntCmd {
				return client.BitOpNot(ctx, "dest", "key")
			})
		})

		It("BitPos", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectBitPos("key", 1, 2, 3)
			}, func() *redis.IntCmd {
				return client.BitPos(ctx, "key", 1, 2, 3)
			})
		})

		It("BitField", func() {
			operationIntSliceCmd(clusterMock, func() *ExpectedIntSlice {
				return clusterMock.ExpectBitField("key", "INCRBY", "i5", 100, 1, "GET", "u4", 0)
			}, func() *redis.IntSliceCmd {
				return client.BitField(ctx, "key", "INCRBY", "i5", 100, 1, "GET", "u4", 0)
			})
		})

		It("Scan", func() {
			operationScanCmd(clusterMock, func() *ExpectedScan {
				return clusterMock.ExpectScan(1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.Scan(ctx, 1, "match", 2)
			})
		})

		It("SScan", func() {
			operationScanCmd(clusterMock, func() *ExpectedScan {
				return clusterMock.ExpectSScan("key", 1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.SScan(ctx, "key", 1, "match", 2)
			})
		})

		It("HScan", func() {
			operationScanCmd(clusterMock, func() *ExpectedScan {
				return clusterMock.ExpectHScan("key", 1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.HScan(ctx, "key", 1, "match", 2)
			})
		})

		It("ZScan", func() {
			operationScanCmd(clusterMock, func() *ExpectedScan {
				return clusterMock.ExpectZScan("key", 1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.ZScan(ctx, "key", 1, "match", 2)
			})
		})

		It("HDel", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectHDel("key", "field1", "field2")
			}, func() *redis.IntCmd {
				return client.HDel(ctx, "key", "field1", "field2")
			})
		})

		It("HExists", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectHExists("key", "field")
			}, func() *redis.BoolCmd {
				return client.HExists(ctx, "key", "field")
			})
		})

		It("HGet", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectHGet("key", "field")
			}, func() *redis.StringCmd {
				return client.HGet(ctx, "key", "field")
			})
		})

		It("HGetAll", func() {
			operationStringStringMapCmd(clusterMock, func() *ExpectedStringStringMap {
				return clusterMock.ExpectHGetAll("key")
			}, func() *redis.StringStringMapCmd {
				return client.HGetAll(ctx, "key")
			})
		})

		It("HIncrBy", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectHIncrBy("key", "field", 1)
			}, func() *redis.IntCmd {
				return client.HIncrBy(ctx, "key", "field", 1)
			})
		})

		It("HIncrByFloat", func() {
			operationFloatCmd(clusterMock, func() *ExpectedFloat {
				return clusterMock.ExpectHIncrByFloat("key", "field", 1.1)
			}, func() *redis.FloatCmd {
				return client.HIncrByFloat(ctx, "key", "field", 1.1)
			})
		})

		It("HKeys", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectHKeys("key")
			}, func() *redis.StringSliceCmd {
				return client.HKeys(ctx, "key")
			})
		})

		It("HLen", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectHLen("key")
			}, func() *redis.IntCmd {
				return client.HLen(ctx, "key")
			})
		})

		It("HMGet", func() {
			operationSliceCmd(clusterMock, func() *ExpectedSlice {
				return clusterMock.ExpectHMGet("key", "field1", "field2")
			}, func() *redis.SliceCmd {
				return client.HMGet(ctx, "key", "field1", "field2")
			})
		})

		It("HSet", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectHSet("key", "field1", "value1", "field2", "value2")
			}, func() *redis.IntCmd {
				return client.HSet(ctx, "key", "field1", "value1", "field2", "value2")
			})
		})

		It("HSet Map", func() {
			clusterMock.ExpectHSet("key", disorder()).SetVal(1)
			res, err := client.HSet(ctx, "key", disorder()).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(int64(1)))

			clusterMock.ExpectHSet("key", "key1", "value1", "key2", "value2").SetVal(1)
			res, err = client.HSet(ctx, "key", "key2", "value2", "key1", "value1").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal(int64(1)))
		})

		It("HMSet", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectHMSet("key", "field1", "value1", "field2", "value2")
			}, func() *redis.BoolCmd {
				return client.HMSet(ctx, "key", "field1", "value1", "field2", "value2")
			})
		})

		It("HMSet Map", func() {
			clusterMock.ExpectHMSet("key", disorder()).SetVal(true)
			res, err := client.HMSet(ctx, "key", disorder()).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeTrue())

			clusterMock.ExpectHMSet("key", "key1", "value1", "key2", "value2").SetVal(true)
			res, err = client.HMSet(ctx, "key", "key2", "value2", "key1", "value1").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(BeTrue())
		})

		It("HSetNX", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectHSetNX("key", "field", "value")
			}, func() *redis.BoolCmd {
				return client.HSetNX(ctx, "key", "field", "value")
			})
		})

		It("HVals", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectHVals("key")
			}, func() *redis.StringSliceCmd {
				return client.HVals(ctx, "key")
			})
		})

		It("BLPop", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectBLPop(1*time.Second, "key1", "key2")
			}, func() *redis.StringSliceCmd {
				return client.BLPop(ctx, 1*time.Second, "key1", "key2")
			})
		})

		It("BRPop", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectBRPop(1*time.Second, "key1", "key2")
			}, func() *redis.StringSliceCmd {
				return client.BRPop(ctx, 1*time.Second, "key1", "key2")
			})
		})

		It("BRPopLPush", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectBRPopLPush("list1", "list2", 1*time.Minute)
			}, func() *redis.StringCmd {
				return client.BRPopLPush(ctx, "list1", "list2", 1*time.Minute)
			})
		})

		It("LIndex", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectLIndex("key", 1)
			}, func() *redis.StringCmd {
				return client.LIndex(ctx, "key", 1)
			})
		})

		It("LInsert", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectLInsert("list", "BEFORE", "World", "There")
			}, func() *redis.IntCmd {
				return client.LInsert(ctx, "list", "BEFORE", "World", "There")
			})
		})

		It("LInsertBefore", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectLInsertBefore("key", "pivot", "value")
			}, func() *redis.IntCmd {
				return client.LInsertBefore(ctx, "key", "pivot", "value")
			})
		})

		It("LInsertAfter", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectLInsertAfter("key", "pivot", "value")
			}, func() *redis.IntCmd {
				return client.LInsertAfter(ctx, "key", "pivot", "value")
			})
		})

		It("LLen", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectLLen("key")
			}, func() *redis.IntCmd {
				return client.LLen(ctx, "key")
			})
		})

		It("LPop", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectLPop("key")
			}, func() *redis.StringCmd {
				return client.LPop(ctx, "key")
			})
		})

		It("LPos", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectLPos("list", "b", redis.LPosArgs{Rank: 2})
			}, func() *redis.IntCmd {
				return client.LPos(ctx, "list", "b", redis.LPosArgs{Rank: 2})
			})
		})

		It("LPosCount", func() {
			operationIntSliceCmd(clusterMock, func() *ExpectedIntSlice {
				return clusterMock.ExpectLPosCount("list", "b", 2, redis.LPosArgs{Rank: 2})
			}, func() *redis.IntSliceCmd {
				return client.LPosCount(ctx, "list", "b", 2, redis.LPosArgs{Rank: 2})
			})
		})

		It("LPush", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectLPush("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.LPush(ctx, "key", "value1", "value2")
			})
		})

		It("LPushX", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectLPushX("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.LPushX(ctx, "key", "value1", "value2")
			})
		})

		It("LRange", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectLRange("key", 1, 2)
			}, func() *redis.StringSliceCmd {
				return client.LRange(ctx, "key", 1, 2)
			})
		})

		It("LRem", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectLRem("key", 2, "value")
			}, func() *redis.IntCmd {
				return client.LRem(ctx, "key", 2, "value")
			})
		})

		It("LSet", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectLSet("key", 1, "value")
			}, func() *redis.StatusCmd {
				return client.LSet(ctx, "key", 1, "value")
			})
		})

		It("LTrim", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectLTrim("key", 1, 2)
			}, func() *redis.StatusCmd {
				return client.LTrim(ctx, "key", 1, 2)
			})
		})

		It("RPop", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectRPop("key")
			}, func() *redis.StringCmd {
				return client.RPop(ctx, "key")
			})
		})

		It("RPopLPush", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectRPopLPush("key", "list")
			}, func() *redis.StringCmd {
				return client.RPopLPush(ctx, "key", "list")
			})
		})

		It("RPush", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectRPush("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.RPush(ctx, "key", "value1", "value2")
			})
		})

		It("RPushX", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectRPushX("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.RPushX(ctx, "key", "value1", "value2")
			})
		})

		It("SAdd", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectSAdd("key", "add")
			}, func() *redis.IntCmd {
				return client.SAdd(ctx, "key", "add")
			})
		})

		It("SCard", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectSCard("key")
			}, func() *redis.IntCmd {
				return client.SCard(ctx, "key")
			})
		})

		It("SDiff", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectSDiff("set1", "set2")
			}, func() *redis.StringSliceCmd {
				return client.SDiff(ctx, "set1", "set2")
			})
		})

		It("SDiffStore", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectSDiffStore("set", "set1", "set2")
			}, func() *redis.IntCmd {
				return client.SDiffStore(ctx, "set", "set1", "set2")
			})
		})

		It("SInter", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectSInter()
			}, func() *redis.StringSliceCmd {
				return client.SInter(ctx)
			})
		})

		It("SInterStore", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectSInterStore("set", "set1", "set2")
			}, func() *redis.IntCmd {
				return client.SInterStore(ctx, "set", "set1", "set2")
			})
		})

		It("SIsMember", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectSIsMember("key", "one")
			}, func() *redis.BoolCmd {
				return client.SIsMember(ctx, "key", "one")
			})
		})

		It("SMembers", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectSMembers("key")
			}, func() *redis.StringSliceCmd {
				return client.SMembers(ctx, "key")
			})
		})

		It("SMembersMap", func() {
			operationStringStructMapCmd(clusterMock, func() *ExpectedStringStructMap {
				return clusterMock.ExpectSMembersMap("key")
			}, func() *redis.StringStructMapCmd {
				return client.SMembersMap(ctx, "key")
			})
		})

		It("SMove", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectSMove("set1", "set2", "two")
			}, func() *redis.BoolCmd {
				return client.SMove(ctx, "set1", "set2", "two")
			})
		})

		It("SPop", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectSPop("key")
			}, func() *redis.StringCmd {
				return client.SPop(ctx, "key")
			})
		})

		It("SPopN", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectSPopN("key", 1)
			}, func() *redis.StringSliceCmd {
				return client.SPopN(ctx, "key", 1)
			})
		})

		It("SRandMember", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectSRandMember("key")
			}, func() *redis.StringCmd {
				return client.SRandMember(ctx, "key")
			})
		})

		It("SRandMemberN", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectSRandMemberN("key", 1)
			}, func() *redis.StringSliceCmd {
				return client.SRandMemberN(ctx, "key", 1)
			})
		})

		It("SRem", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectSRem("set", "one")
			}, func() *redis.IntCmd {
				return client.SRem(ctx, "set", "one")
			})
		})

		It("SUnion", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectSUnion()
			}, func() *redis.StringSliceCmd {
				return client.SUnion(ctx)
			})
		})

		It("SUnionStore", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectSUnionStore("set", "set1", "set2")
			}, func() *redis.IntCmd {
				return client.SUnionStore(ctx, "set", "set1", "set2")
			})
		})

		It("XAdd", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectXAdd(&redis.XAddArgs{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectXDel("stream", "1-0", "2-0", "3-0")
			}, func() *redis.IntCmd {
				return client.XDel(ctx, "stream", "1-0", "2-0", "3-0")
			})
		})

		It("XLen", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectXLen("stream")
			}, func() *redis.IntCmd {
				return client.XLen(ctx, "stream")
			})
		})

		It("XRange", func() {
			operationXMessageSliceCmd(clusterMock, func() *ExpectedXMessageSlice {
				return clusterMock.ExpectXRange("stream", "-", "+")
			}, func() *redis.XMessageSliceCmd {
				return client.XRange(ctx, "stream", "-", "+")
			})
		})

		It("XRangeN", func() {
			operationXMessageSliceCmd(clusterMock, func() *ExpectedXMessageSlice {
				return clusterMock.ExpectXRangeN("stream", "-", "+", 2)
			}, func() *redis.XMessageSliceCmd {
				return client.XRangeN(ctx, "stream", "-", "+", 2)
			})
		})

		It("XRevRange", func() {
			operationXMessageSliceCmd(clusterMock, func() *ExpectedXMessageSlice {
				return clusterMock.ExpectXRevRange("stream", "+", "-")
			}, func() *redis.XMessageSliceCmd {
				return client.XRevRange(ctx, "stream", "+", "-")
			})
		})

		It("XRevRangeN", func() {
			operationXMessageSliceCmd(clusterMock, func() *ExpectedXMessageSlice {
				return clusterMock.ExpectXRevRangeN("stream", "+", "-", 2)
			}, func() *redis.XMessageSliceCmd {
				return client.XRevRangeN(ctx, "stream", "+", "-", 2)
			})
		})

		It("XRead", func() {
			operationXStreamSliceCmd(clusterMock, func() *ExpectedXStreamSlice {
				return clusterMock.ExpectXRead(&redis.XReadArgs{
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
			operationXStreamSliceCmd(clusterMock, func() *ExpectedXStreamSlice {
				return clusterMock.ExpectXReadStreams()
			}, func() *redis.XStreamSliceCmd {
				return client.XReadStreams(ctx)
			})
		})

		It("XGroupCreate", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectXGroupCreate("stream", "group", "0")
			}, func() *redis.StatusCmd {
				return client.XGroupCreate(ctx, "stream", "group", "0")
			})
		})

		It("XGroupCreateMkStream", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectXGroupCreateMkStream("stream", "group", "0")
			}, func() *redis.StatusCmd {
				return client.XGroupCreateMkStream(ctx, "stream", "group", "0")
			})
		})

		It("XGroupSetID", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectXGroupSetID("stream", "group", "0")
			}, func() *redis.StatusCmd {
				return client.XGroupSetID(ctx, "stream", "group", "0")
			})
		})

		It("XGroupDestroy", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectXGroupDestroy("stream", "group")
			}, func() *redis.IntCmd {
				return client.XGroupDestroy(ctx, "stream", "group")
			})
		})

		It("XGroupDelConsumer", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectXGroupDelConsumer("stream", "group", "consumer")
			}, func() *redis.IntCmd {
				return client.XGroupDelConsumer(ctx, "stream", "group", "consumer")
			})
		})

		It("XReadGroup", func() {
			operationXStreamSliceCmd(clusterMock, func() *ExpectedXStreamSlice {
				return clusterMock.ExpectXReadGroup(&redis.XReadGroupArgs{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectXAck("stream", "group", "1-0", "2-0", "4-0")
			}, func() *redis.IntCmd {
				return client.XAck(ctx, "stream", "group", "1-0", "2-0", "4-0")
			})
		})

		It("XPending", func() {
			operationXPendingCmd(clusterMock, func() *ExpectedXPending {
				return clusterMock.ExpectXPending("stream", "group")
			}, func() *redis.XPendingCmd {
				return client.XPending(ctx, "stream", "group")
			})
		})

		It("XPendingExt", func() {
			operationXPendingExtCmd(clusterMock, func() *ExpectedXPendingExt {
				return clusterMock.ExpectXPendingExt(&redis.XPendingExtArgs{
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
			operationXMessageSliceCmd(clusterMock, func() *ExpectedXMessageSlice {
				return clusterMock.ExpectXClaim(&redis.XClaimArgs{
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
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectXClaimJustID(&redis.XClaimArgs{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectXTrim("stream", 0)
			}, func() *redis.IntCmd {
				return client.XTrim(ctx, "stream", 0)
			})
		})

		It("XTrimApprox", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectXTrimApprox("stream", 0)
			}, func() *redis.IntCmd {
				return client.XTrimApprox(ctx, "stream", 0)
			})
		})

		It("XInfoGroups", func() {
			operationXInfoGroupsCmd(clusterMock, func() *ExpectedXInfoGroups {
				return clusterMock.ExpectXInfoGroups("key")
			}, func() *redis.XInfoGroupsCmd {
				return client.XInfoGroups(ctx, "key")
			})
		})

		It("XInfoStream", func() {
			operationXInfoStreamCmd(clusterMock, func() *ExpectedXInfoStream {
				return clusterMock.ExpectXInfoStream("key")
			}, func() *redis.XInfoStreamCmd {
				return client.XInfoStream(ctx, "key")
			})
		})

		It("BZPopMax", func() {
			operationZWithKeyCmd(clusterMock, func() *ExpectedZWithKey {
				return clusterMock.ExpectBZPopMax(0, "zset1", "zset2")
			}, func() *redis.ZWithKeyCmd {
				return client.BZPopMax(ctx, 0, "zset1", "zset2")
			})
		})

		It("BZPopMin", func() {
			operationZWithKeyCmd(clusterMock, func() *ExpectedZWithKey {
				return clusterMock.ExpectBZPopMin(0, "zset1", "zset2")
			}, func() *redis.ZWithKeyCmd {
				return client.BZPopMin(ctx, 0, "zset1", "zset2")
			})
		})

		It("ZAdd", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZAdd("zset", &redis.Z{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZAddNX("zset", &redis.Z{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZAddXX("zset", &redis.Z{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZAddCh("zset", &redis.Z{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZAddNXCh("zset", &redis.Z{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZAddXXCh("zset", &redis.Z{
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
			operationFloatCmd(clusterMock, func() *ExpectedFloat {
				return clusterMock.ExpectZIncr("zset", &redis.Z{
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
			operationFloatCmd(clusterMock, func() *ExpectedFloat {
				return clusterMock.ExpectZIncrNX("zset", &redis.Z{
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
			operationFloatCmd(clusterMock, func() *ExpectedFloat {
				return clusterMock.ExpectZIncrXX("zset", &redis.Z{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZCard("key")
			}, func() *redis.IntCmd {
				return client.ZCard(ctx, "key")
			})
		})

		It("ZCount", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZCount("zset", "-inf", "+inf")
			}, func() *redis.IntCmd {
				return client.ZCount(ctx, "zset", "-inf", "+inf")
			})
		})

		It("ZLexCount", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZLexCount("zset", "-", "+")
			}, func() *redis.IntCmd {
				return client.ZLexCount(ctx, "zset", "-", "+")
			})
		})

		It("ZIncrBy", func() {
			operationFloatCmd(clusterMock, func() *ExpectedFloat {
				return clusterMock.ExpectZIncrBy("zset", 2, "one")
			}, func() *redis.FloatCmd {
				return client.ZIncrBy(ctx, "zset", 2, "one")
			})
		})

		It("ZInterStore", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZInterStore("out", &redis.ZStore{
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
			operationZSliceCmd(clusterMock, func() *ExpectedZSlice {
				return clusterMock.ExpectZPopMax("key")
			}, func() *redis.ZSliceCmd {
				return client.ZPopMax(ctx, "key")
			})
		})

		It("ZPopMin", func() {
			operationZSliceCmd(clusterMock, func() *ExpectedZSlice {
				return clusterMock.ExpectZPopMin("key")
			}, func() *redis.ZSliceCmd {
				return client.ZPopMin(ctx, "key")
			})
		})

		It("ZRange", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectZRange("zset", 0, -1)
			}, func() *redis.StringSliceCmd {
				return client.ZRange(ctx, "zset", 0, -1)
			})
		})

		It("ZRangeWithScores", func() {
			operationZSliceCmd(clusterMock, func() *ExpectedZSlice {
				return clusterMock.ExpectZRangeWithScores("zset", 0, -1)
			}, func() *redis.ZSliceCmd {
				return client.ZRangeWithScores(ctx, "zset", 0, -1)
			})
		})

		It("ZRangeByScore", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectZRangeByScore("zset", &redis.ZRangeBy{
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
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectZRangeByLex("zset", &redis.ZRangeBy{
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
			operationZSliceCmd(clusterMock, func() *ExpectedZSlice {
				return clusterMock.ExpectZRangeByScoreWithScores("zset", &redis.ZRangeBy{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZRank("zset", "three")
			}, func() *redis.IntCmd {
				return client.ZRank(ctx, "zset", "three")
			})
		})

		It("ZRem", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZRem("zset", "two")
			}, func() *redis.IntCmd {
				return client.ZRem(ctx, "zset", "two")
			})
		})

		It("ZRemRangeByRank", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZRemRangeByRank("key", 1, 2)
			}, func() *redis.IntCmd {
				return client.ZRemRangeByRank(ctx, "key", 1, 2)
			})
		})

		It("ZRemRangeByScore", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZRemRangeByScore("zset", "-inf", "(2")
			}, func() *redis.IntCmd {
				return client.ZRemRangeByScore(ctx, "zset", "-inf", "(2")
			})
		})

		It("ZRemRangeByLex", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZRemRangeByLex("zset", "[alpha", "[omega")
			}, func() *redis.IntCmd {
				return client.ZRemRangeByLex(ctx, "zset", "[alpha", "[omega")
			})
		})

		It("ZRevRange", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectZRevRange("zset", 0, -1)
			}, func() *redis.StringSliceCmd {
				return client.ZRevRange(ctx, "zset", 0, -1)
			})
		})

		It("ZRevRangeWithScores", func() {
			operationZSliceCmd(clusterMock, func() *ExpectedZSlice {
				return clusterMock.ExpectZRevRangeWithScores("zset", 0, -1)
			}, func() *redis.ZSliceCmd {
				return client.ZRevRangeWithScores(ctx, "zset", 0, -1)
			})
		})

		It("ZRevRangeByScore", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectZRevRangeByScore("zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			}, func() *redis.StringSliceCmd {
				return client.ZRevRangeByScore(ctx, "zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			})
		})

		It("ZRevRangeByLex", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectZRevRangeByLex("zset", &redis.ZRangeBy{Max: "+", Min: "-"})
			}, func() *redis.StringSliceCmd {
				return client.ZRevRangeByLex(ctx, "zset", &redis.ZRangeBy{Max: "+", Min: "-"})
			})
		})

		It("ZRevRangeByScoreWithScores", func() {
			operationZSliceCmd(clusterMock, func() *ExpectedZSlice {
				return clusterMock.ExpectZRevRangeByScoreWithScores("zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			}, func() *redis.ZSliceCmd {
				return client.ZRevRangeByScoreWithScores(ctx, "zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			})
		})

		It("ZRevRank", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZRevRank("key", "member")
			}, func() *redis.IntCmd {
				return client.ZRevRank(ctx, "key", "member")
			})
		})

		It("ZScore", func() {
			operationFloatCmd(clusterMock, func() *ExpectedFloat {
				return clusterMock.ExpectZScore("key", "member")
			}, func() *redis.FloatCmd {
				return client.ZScore(ctx, "key", "member")
			})
		})

		It("ZUnionWithScores", func() {
			operationZSliceCmd(clusterMock, func() *ExpectedZSlice {
				return clusterMock.ExpectZUnionWithScores(redis.ZStore{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectZUnionStore("out", &redis.ZStore{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectPFAdd("hll1", "1", "2", "3", "4", "5")
			}, func() *redis.IntCmd {
				return client.PFAdd(ctx, "hll1", "1", "2", "3", "4", "5")
			})
		})

		It("PFCount", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectPFCount("hll1", "hll2")
			}, func() *redis.IntCmd {
				return client.PFCount(ctx, "hll1", "hll2")
			})
		})

		It("PFMerge", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectPFMerge("hllMerged", "hll1", "hll2")
			}, func() *redis.StatusCmd {
				return client.PFMerge(ctx, "hllMerged", "hll1", "hll2")
			})
		})

		It("BgRewriteAOF", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectBgRewriteAOF()
			}, func() *redis.StatusCmd {
				return client.BgRewriteAOF(ctx)
			})
		})

		It("BgSave", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectBgSave()
			}, func() *redis.StatusCmd {
				return client.BgSave(ctx)
			})
		})

		It("ClientKill", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClientKill("1.1.1.1:1111")
			}, func() *redis.StatusCmd {
				return client.ClientKill(ctx, "1.1.1.1:1111")
			})
		})

		It("ClientKillByFilter", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectClientKillByFilter("11.11.11.11:1234")
			}, func() *redis.IntCmd {
				return client.ClientKillByFilter(ctx, "11.11.11.11:1234")
			})
		})

		It("ClientList", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectClientList()
			}, func() *redis.StringCmd {
				return client.ClientList(ctx)
			})
		})

		It("ClientPause", func() {
			operationBoolCmd(clusterMock, func() *ExpectedBool {
				return clusterMock.ExpectClientPause(1 * time.Minute)
			}, func() *redis.BoolCmd {
				return client.ClientPause(ctx, 1*time.Minute)
			})
		})

		It("ClientID", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectClientID()
			}, func() *redis.IntCmd {
				return client.ClientID(ctx)
			})
		})

		It("ConfigGet", func() {
			operationSliceCmd(clusterMock, func() *ExpectedSlice {
				return clusterMock.ExpectConfigGet("*")
			}, func() *redis.SliceCmd {
				return client.ConfigGet(ctx, "*")
			})
		})

		It("ConfigResetStat", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectConfigResetStat()
			}, func() *redis.StatusCmd {
				return client.ConfigResetStat(ctx)
			})
		})

		It("ConfigSet", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectConfigSet("maxmemory", "1024M")
			}, func() *redis.StatusCmd {
				return client.ConfigSet(ctx, "maxmemory", "1024M")
			})
		})

		It("ConfigRewrite", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectConfigRewrite()
			}, func() *redis.StatusCmd {
				return client.ConfigRewrite(ctx)
			})
		})

		//It("DBSize", func() {
		//	operationIntCmd(clusterMock, func() *ExpectedInt {
		//		return clusterMock.ExpectDBSize()
		//	}, func() *redis.IntCmd {
		//		return client.DBSize(ctx)
		//	})
		//})

		It("FlushAll", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectFlushAll()
			}, func() *redis.StatusCmd {
				return client.FlushAll(ctx)
			})
		})

		It("FlushAllAsync", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectFlushAllAsync()
			}, func() *redis.StatusCmd {
				return client.FlushAllAsync(ctx)
			})
		})

		It("FlushDB", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectFlushDB()
			}, func() *redis.StatusCmd {
				return client.FlushDB(ctx)
			})
		})

		It("FlushDBAsync", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectFlushDBAsync()
			}, func() *redis.StatusCmd {
				return client.FlushDBAsync(ctx)
			})
		})

		It("Info", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectInfo()
			}, func() *redis.StringCmd {
				return client.Info(ctx)
			})
		})

		It("LastSave", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectLastSave()
			}, func() *redis.IntCmd {
				return client.LastSave(ctx)
			})
		})

		It("Save", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectSave()
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
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectSlaveOf("localhost", "8888")
			}, func() *redis.StatusCmd {
				return client.SlaveOf(ctx, "localhost", "8888")
			})
		})

		It("Time", func() {
			operationTimeCmd(clusterMock, func() *ExpectedTime {
				return clusterMock.ExpectTime()
			}, func() *redis.TimeCmd {
				return client.Time(ctx)
			})
		})

		It("DebugObject", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectDebugObject("foo")
			}, func() *redis.StringCmd {
				return client.DebugObject(ctx, "foo")
			})
		})

		It("ReadOnly", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectReadOnly()
			}, func() *redis.StatusCmd {
				return client.ReadOnly(ctx)
			})
		})

		It("ReadWrite", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectReadWrite()
			}, func() *redis.StatusCmd {
				return client.ReadWrite(ctx)
			})
		})

		It("MemoryUsage", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectMemoryUsage("foo")
			}, func() *redis.IntCmd {
				return client.MemoryUsage(ctx, "foo")
			})
		})

		It("Eval", func() {
			operationCmdCmd(clusterMock, func() *ExpectedCmd {
				return clusterMock.ExpectEval("return {KEYS[1],ARGV[1]}", []string{"key"}, "hello")
			}, func() *redis.Cmd {
				return client.Eval(ctx, "return {KEYS[1],ARGV[1]}", []string{"key"}, "hello")
			})
		})

		It("EvalSha", func() {
			operationCmdCmd(clusterMock, func() *ExpectedCmd {
				return clusterMock.ExpectEvalSha("sha", []string{"key1", "key2"}, "args1", "args2")
			}, func() *redis.Cmd {
				return client.EvalSha(ctx, "sha", []string{"key1", "key2"}, "args1", "args2")
			})
		})

		It("ScriptExists", func() {
			operationBoolSliceCmd(clusterMock, func() *ExpectedBoolSlice {
				return clusterMock.ExpectScriptExists()
			}, func() *redis.BoolSliceCmd {
				return client.ScriptExists(ctx)
			})
		})

		It("ScriptFlush", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectScriptFlush()
			}, func() *redis.StatusCmd {
				return client.ScriptFlush(ctx)
			})
		})

		It("ScriptKill", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectScriptKill()
			}, func() *redis.StatusCmd {
				return client.ScriptKill(ctx)
			})
		})

		It("ScriptLoad", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectScriptLoad("script")
			}, func() *redis.StringCmd {
				return client.ScriptLoad(ctx, "script")
			})
		})

		It("Publish", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectPublish("channel", "message")
			}, func() *redis.IntCmd {
				return client.Publish(ctx, "channel", "message")
			})
		})

		It("PubSubChannels", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectPubSubChannels("pattern")
			}, func() *redis.StringSliceCmd {
				return client.PubSubChannels(ctx, "pattern")
			})
		})

		It("PubSubNumSub", func() {
			operationStringIntMapCmd(clusterMock, func() *ExpectedStringIntMap {
				return clusterMock.ExpectPubSubNumSub()
			}, func() *redis.StringIntMapCmd {
				return client.PubSubNumSub(ctx)
			})
		})

		It("PubSubNumPat", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectPubSubNumPat()
			}, func() *redis.IntCmd {
				return client.PubSubNumPat(ctx)
			})
		})

		It("ClusterSlots", func() {
			operationClusterSlotsCmd(clusterMock, func() *ExpectedClusterSlots {
				return clusterMock.ExpectClusterSlots()
			}, func() *redis.ClusterSlotsCmd {
				return client.ClusterSlots(ctx)
			})
		})

		It("ClusterNodes", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectClusterNodes()
			}, func() *redis.StringCmd {
				return client.ClusterNodes(ctx)
			})
		})

		It("ClusterMeet", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterMeet("1.1.1.1", "1")
			}, func() *redis.StatusCmd {
				return client.ClusterMeet(ctx, "1.1.1.1", "1")
			})
		})

		It("ClusterForget", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterForget("id")
			}, func() *redis.StatusCmd {
				return client.ClusterForget(ctx, "id")
			})
		})

		It("ClusterReplicate", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterReplicate("id")
			}, func() *redis.StatusCmd {
				return client.ClusterReplicate(ctx, "id")
			})
		})

		It("ClusterResetSoft", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterResetSoft()
			}, func() *redis.StatusCmd {
				return client.ClusterResetSoft(ctx)
			})
		})

		It("ClusterResetHard", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterResetHard()
			}, func() *redis.StatusCmd {
				return client.ClusterResetHard(ctx)
			})
		})

		It("ClusterInfo", func() {
			operationStringCmd(clusterMock, func() *ExpectedString {
				return clusterMock.ExpectClusterInfo()
			}, func() *redis.StringCmd {
				return client.ClusterInfo(ctx)
			})
		})

		It("ClusterKeySlot", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectClusterKeySlot("key")
			}, func() *redis.IntCmd {
				return client.ClusterKeySlot(ctx, "key")
			})
		})

		It("ClusterGetKeysInSlot", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectClusterGetKeysInSlot(1, 2)
			}, func() *redis.StringSliceCmd {
				return client.ClusterGetKeysInSlot(ctx, 1, 2)
			})
		})

		It("ClusterCountFailureReports", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectClusterCountFailureReports("id")
			}, func() *redis.IntCmd {
				return client.ClusterCountFailureReports(ctx, "id")
			})
		})

		It("ClusterCountKeysInSlot", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectClusterCountKeysInSlot(1)
			}, func() *redis.IntCmd {
				return client.ClusterCountKeysInSlot(ctx, 1)
			})
		})

		It("ClusterDelSlots", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterDelSlots()
			}, func() *redis.StatusCmd {
				return client.ClusterDelSlots(ctx)
			})
		})

		It("ClusterDelSlotsRange", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterDelSlotsRange(1, 2)
			}, func() *redis.StatusCmd {
				return client.ClusterDelSlotsRange(ctx, 1, 2)
			})
		})

		It("ClusterSaveConfig", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterSaveConfig()
			}, func() *redis.StatusCmd {
				return client.ClusterSaveConfig(ctx)
			})
		})

		It("ClusterSlaves", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectClusterSlaves("id")
			}, func() *redis.StringSliceCmd {
				return client.ClusterSlaves(ctx, "id")
			})
		})

		It("ClusterFailover", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterFailover()
			}, func() *redis.StatusCmd {
				return client.ClusterFailover(ctx)
			})
		})

		It("ClusterAddSlots", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterAddSlots()
			}, func() *redis.StatusCmd {
				return client.ClusterAddSlots(ctx)
			})
		})

		It("ClusterAddSlotsRange", func() {
			operationStatusCmd(clusterMock, func() *ExpectedStatus {
				return clusterMock.ExpectClusterAddSlotsRange(1, 2)
			}, func() *redis.StatusCmd {
				return client.ClusterAddSlotsRange(ctx, 1, 2)
			})
		})

		It("GeoAdd", func() {
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectGeoAdd("Sicily",
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
			operationGeoPosCmd(clusterMock, func() *ExpectedGeoPos {
				return clusterMock.ExpectGeoPos("Sicily", "Palermo", "Catania", "NonExisting")
			}, func() *redis.GeoPosCmd {
				return client.GeoPos(ctx, "Sicily", "Palermo", "Catania", "NonExisting")
			})
		})

		It("GeoRadius", func() {
			operationGeoLocationCmd(clusterMock, func() *ExpectedGeoLocation {
				return clusterMock.ExpectGeoRadius("Sicily", 15, 37, &redis.GeoRadiusQuery{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectGeoRadiusStore("Sicily", 15, 37, &redis.GeoRadiusQuery{
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
			operationGeoLocationCmd(clusterMock, func() *ExpectedGeoLocation {
				return clusterMock.ExpectGeoRadiusByMember("Sicily", "Catania", &redis.GeoRadiusQuery{
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
			operationIntCmd(clusterMock, func() *ExpectedInt {
				return clusterMock.ExpectGeoRadiusByMemberStore("key", "member", &redis.GeoRadiusQuery{
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
			operationFloatCmd(clusterMock, func() *ExpectedFloat {
				return clusterMock.ExpectGeoDist("Sicily", "Palermo", "Catania", "km")
			}, func() *redis.FloatCmd {
				return client.GeoDist(ctx, "Sicily", "Palermo", "Catania", "km")
			})
		})

		It("GeoHash", func() {
			operationStringSliceCmd(clusterMock, func() *ExpectedStringSlice {
				return clusterMock.ExpectGeoHash("Sicily", "Palermo", "Catania")
			}, func() *redis.StringSliceCmd {
				return client.GeoHash(ctx, "Sicily", "Palermo", "Catania")
			})
		})
	})
})
