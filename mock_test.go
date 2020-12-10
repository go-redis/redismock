package redismock

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRedisMock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "redis mock")
}

var _ = Describe("RedisMock", func() {
	var (
		client *redis.Client
		mock   ClientMock
	)

	BeforeEach(func() {
		client, mock = NewClientMock()
	})

	AfterEach(func() {
		Expect(client.Close()).NotTo(HaveOccurred())
		Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())
	})

	Describe("pipeline", func() {
		var pipe redis.Pipeliner

		BeforeEach(func() {
			mock.ExpectGet("key1").SetVal("pipeline get")
			mock.ExpectHGet("hash_key", "hash_field").SetVal("pipeline hash get")
			mock.ExpectSet("set_key", "set value", 1*time.Minute).SetVal("OK")

			pipe = client.Pipeline()
		})

		It("pipeline order", func() {
			mock.MatchExpectationsInOrder(true)

			get := pipe.Get("key1")
			hashGet := pipe.HGet("hash_key", "hash_field")
			set := pipe.Set("set_key", "set value", 1*time.Minute)

			_, err := pipe.Exec()
			Expect(err).NotTo(HaveOccurred())

			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal("pipeline get"))

			Expect(hashGet.Err()).NotTo(HaveOccurred())
			Expect(hashGet.Val()).To(Equal("pipeline hash get"))

			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))
		})

		It("pipeline not order", func() {
			mock.MatchExpectationsInOrder(false)

			hashGet := pipe.HGet("hash_key", "hash_field")
			set := pipe.Set("set_key", "set value", 1*time.Minute)
			get := pipe.Get("key1")

			_, err := pipe.Exec()
			Expect(err).NotTo(HaveOccurred())

			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal("pipeline get"))

			Expect(hashGet.Err()).NotTo(HaveOccurred())
			Expect(hashGet.Val()).To(Equal("pipeline hash get"))

			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))
		})
	})

	Describe("work order", func() {

		BeforeEach(func() {
			mock.ExpectGet("key").RedisNil()
			mock.ExpectSet("key", "1", 1*time.Second).SetVal("OK")
			mock.ExpectGet("key").SetVal("1")
			mock.ExpectGetSet("key", "0").SetVal("1")
		})

		It("ordinary", func() {
			get := client.Get("key")
			Expect(get.Err()).To(Equal(redis.Nil))
			Expect(get.Val()).To(Equal(""))

			set := client.Set("key", "1", 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))

			get = client.Get("key")
			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal("1"))

			getSet := client.GetSet("key", "0")
			Expect(getSet.Err()).NotTo(HaveOccurred())
			Expect(getSet.Val()).To(Equal("1"))
		})

		It("surplus", func() {
			_ = client.Get("key")

			set := client.Set("key", "1", 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))

			Expect(mock.ExpectationsWereMet()).To(HaveOccurred())

			_ = client.Get("key")
			Expect(mock.ExpectationsWereMet()).To(HaveOccurred())

			_ = client.GetSet("key", "0")
		})

		It("not enough", func() {
			_ = client.Get("key")
			_ = client.Set("key", "1", 1*time.Second)
			_ = client.Get("key")
			_ = client.GetSet("key", "0")
			Expect(mock.ExpectationsWereMet()).NotTo(HaveOccurred())

			get := client.HGet("key", "field")
			Expect(get.Err()).To(HaveOccurred())
			Expect(get.Val()).To(Equal(""))
		})
	})

	Describe("work not order", func() {

		BeforeEach(func() {
			mock.MatchExpectationsInOrder(false)

			mock.ExpectSet("key", "1", 1*time.Second).SetVal("OK")
			mock.ExpectGet("key").SetVal("1")
			mock.ExpectGetSet("key", "0").SetVal("1")
		})

		It("ordinary", func() {
			get := client.Get("key")
			Expect(get.Err()).NotTo(HaveOccurred())
			Expect(get.Val()).To(Equal("1"))

			set := client.Set("key", "1", 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))

			getSet := client.GetSet("key", "0")
			Expect(getSet.Err()).NotTo(HaveOccurred())
			Expect(getSet.Val()).To(Equal("1"))
		})
	})

	Describe("work other match", func() {

		It("regexp match", func() {
			mock.Regexp().ExpectSet("key", `^order_id_[0-9]{10}$`, 1*time.Second).SetVal("OK")
			mock.Regexp().ExpectSet("key2", `^order_id_[0-9]{4}\-[0-9]{2}\-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}.+$`, 1*time.Second).SetVal("OK")

			set := client.Set("key", fmt.Sprintf("order_id_%d", time.Now().Unix()), 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))

			// no regexp
			set = client.Set("key2", fmt.Sprintf("order_id_%s", time.Now().Format(time.UnixDate)), 1*time.Second)
			Expect(set.Err()).To(HaveOccurred())
			Expect(set.Val()).To(Equal(""))

			set = client.Set("key2", fmt.Sprintf("order_id_%s", time.Now().Format(time.RFC3339)), 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))
		})

		It("custom match", func() {
			mock.CustomMatch(func(expected, actual []interface{}) error {
				return errors.New("mismatch")
			}).ExpectGet("key").SetVal("OK")

			get := client.Get("key")
			Expect(get.Err()).To(Equal(errors.New("mismatch")))
			Expect(get.Val()).To(Equal(""))

			set := client.Incr("key")
			Expect(set.Err()).To(HaveOccurred())
			Expect(set.Err()).NotTo(Equal(errors.New("mismatch")))
			Expect(set.Val()).To(Equal(int64(0)))

			// no match, no pass
			Expect(mock.ExpectationsWereMet()).To(HaveOccurred())

			// let AfterEach pass
			mock.ClearExpect()
		})

	})

	Describe("work error", func() {

		It("set error", func() {
			mock.ExpectGet("key").SetErr(errors.New("set error"))

			get := client.Get("key")
			Expect(get.Err()).To(Equal(errors.New("set error")))
			Expect(get.Val()).To(Equal(""))
		})

		It("not set", func() {
			mock.ExpectGet("key")

			get := client.Get("key")
			Expect(get.Err()).To(HaveOccurred())
			Expect(get.Val()).To(Equal(""))
		})

		It("set zero", func() {
			mock.ExpectGet("key").SetVal("")

			get := client.Get("key")
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
			mock.ExpectCommand().SetVal(commandsInfo)

			commands, err := client.Command().Result()
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
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectClientGetName()
			}, func() *redis.StringCmd {
				return client.ClientGetName()
			})
		})

		It("Echo", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectEcho("mock")
			}, func() *redis.StringCmd {
				return client.Echo("mock")
			})
		})

		It("Ping", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectPing()
			}, func() *redis.StatusCmd {
				return client.Ping()
			})
		})

		It("Quit", func() {
			//not implemented
		})

		It("Del", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectDel()
			}, func() *redis.IntCmd {
				return client.Del()
			})
		})

		It("Unlink", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectUnlink()
			}, func() *redis.IntCmd {
				return client.Unlink()
			})
		})

		It("Dump", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectDump("key")
			}, func() *redis.StringCmd {
				return client.Dump("key")
			})
		})

		It("Exists", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectExists()
			}, func() *redis.IntCmd {
				return client.Exists()
			})
		})

		It("Expire", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectExpire("key", 1*time.Second)
			}, func() *redis.BoolCmd {
				return client.Expire("key", 1*time.Second)
			})
		})

		It("ExpireAt", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectExpireAt("key", time.Now())
			}, func() *redis.BoolCmd {
				return client.ExpireAt("key", time.Now())
			})
		})

		It("Keys", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectKeys("key")
			}, func() *redis.StringSliceCmd {
				return client.Keys("key")
			})
		})

		It("Migrate", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectMigrate("host", "port", "key", 1, 1*time.Hour)
			}, func() *redis.StatusCmd {
				return client.Migrate("host", "port", "key", 1, 1*time.Hour)
			})
		})

		It("Move", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectMove("key", 1)
			}, func() *redis.BoolCmd {
				return client.Move("key", 1)
			})
		})

		It("ObjectRefCount", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectObjectRefCount("key")
			}, func() *redis.IntCmd {
				return client.ObjectRefCount("key")
			})
		})

		It("ObjectEncoding", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectObjectEncoding("key")
			}, func() *redis.StringCmd {
				return client.ObjectEncoding("key")
			})
		})

		It("ObjectIdleTime", func() {
			operationDurationCmd(mock, func() *ExpectedDuration {
				return mock.ExpectObjectIdleTime("key")
			}, func() *redis.DurationCmd {
				return client.ObjectIdleTime("key")
			})
		})

		It("Persist", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectPersist("key")
			}, func() *redis.BoolCmd {
				return client.Persist("key")
			})
		})

		It("PExpire", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectPExpire("key", 1*time.Minute)
			}, func() *redis.BoolCmd {
				return client.PExpire("key", 1*time.Minute)
			})
		})

		It("PExpireAt", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectPExpireAt("key", time.Now())
			}, func() *redis.BoolCmd {
				return client.PExpireAt("key", time.Now())
			})
		})

		It("PTTL", func() {
			operationDurationCmd(mock, func() *ExpectedDuration {
				return mock.ExpectPTTL("key")
			}, func() *redis.DurationCmd {
				return client.PTTL("key")
			})
		})

		It("RandomKey", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectRandomKey()
			}, func() *redis.StringCmd {
				return client.RandomKey()
			})
		})

		It("Rename", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectRename("key", "new_key")
			}, func() *redis.StatusCmd {
				return client.Rename("key", "new_key")
			})
		})

		It("RenameNX", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectRenameNX("key", "new_key")
			}, func() *redis.BoolCmd {
				return client.RenameNX("key", "new_key")
			})
		})

		It("Restore", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectRestore("key", 1*time.Minute, "value")
			}, func() *redis.StatusCmd {
				return client.Restore("key", 1*time.Minute, "value")
			})
		})

		It("RestoreReplace", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectRestoreReplace("key", 1*time.Minute, "value")
			}, func() *redis.StatusCmd {
				return client.RestoreReplace("key", 1*time.Minute, "value")
			})
		})

		It("Sort", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectSort("key", &redis.Sort{
					Offset: 0,
					Count:  2,
					Order:  "ASC",
				})
			}, func() *redis.StringSliceCmd {
				return client.Sort("key", &redis.Sort{
					Offset: 0,
					Count:  2,
					Order:  "ASC",
				})
			})
		})

		It("SortStore", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectSortStore("key", "store", &redis.Sort{
					Offset: 0,
					Count:  2,
					Order:  "ASC",
				})
			}, func() *redis.IntCmd {
				return client.SortStore("key", "store", &redis.Sort{
					Offset: 0,
					Count:  2,
					Order:  "ASC",
				})
			})
		})

		It("SortInterfaces", func() {
			operationSliceCmd(mock, func() *ExpectedSlice {
				return mock.ExpectSortInterfaces("key", &redis.Sort{
					Get: []string{"object_*"},
				})
			}, func() *redis.SliceCmd {
				return client.SortInterfaces("key", &redis.Sort{
					Get: []string{"object_*"},
				})
			})
		})

		It("Touch", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectTouch()
			}, func() *redis.IntCmd {
				return client.Touch()
			})
		})

		It("TTL", func() {
			operationDurationCmd(mock, func() *ExpectedDuration {
				return mock.ExpectTTL("key")
			}, func() *redis.DurationCmd {
				return client.TTL("key")
			})
		})

		It("Type", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectType("key")
			}, func() *redis.StatusCmd {
				return client.Type("key")
			})
		})

		It("Append", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectAppend("key", "value")
			}, func() *redis.IntCmd {
				return client.Append("key", "value")
			})
		})

		It("Decr", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectDecr("key")
			}, func() *redis.IntCmd {
				return client.Decr("key")
			})
		})

		It("DecrBy", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectDecrBy("key", 1)
			}, func() *redis.IntCmd {
				return client.DecrBy("key", 1)
			})
		})

		It("Get", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectGet("key")
			}, func() *redis.StringCmd {
				return client.Get("key")
			})
		})

		It("GetRange", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectGetRange("key", 1, 10)
			}, func() *redis.StringCmd {
				return client.GetRange("key", 1, 10)
			})
		})

		It("GetSet", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectGetSet("key", 1)
			}, func() *redis.StringCmd {
				return client.GetSet("key", 1)
			})
		})

		It("Incr", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectIncr("key")
			}, func() *redis.IntCmd {
				return client.Incr("key")
			})
		})

		It("IncrBy", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectIncrBy("key", 1)
			}, func() *redis.IntCmd {
				return client.IncrBy("key", 1)
			})
		})

		It("IncrByFloat", func() {
			operationFloatCmd(mock, func() *ExpectedFloat {
				return mock.ExpectIncrByFloat("key", 1)
			}, func() *redis.FloatCmd {
				return client.IncrByFloat("key", 1)
			})
		})

		It("MGet", func() {
			operationSliceCmd(mock, func() *ExpectedSlice {
				return mock.ExpectMGet()
			}, func() *redis.SliceCmd {
				return client.MGet()
			})
		})

		It("MSet", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectMSet()
			}, func() *redis.StatusCmd {
				return client.MSet()
			})
		})

		It("MSetNX", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectMSetNX()
			}, func() *redis.BoolCmd {
				return client.MSetNX()
			})
		})

		It("Set", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectSet("key", "value", 1*time.Minute)
			}, func() *redis.StatusCmd {
				return client.Set("key", "value", 1*time.Minute)
			})
		})

		It("SetNX", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectSetNX("key", "value", 1*time.Minute)
			}, func() *redis.BoolCmd {
				return client.SetNX("key", "value", 1*time.Minute)
			})
		})

		It("SetXX", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectSetXX("key", "value", 1*time.Minute)
			}, func() *redis.BoolCmd {
				return client.SetXX("key", "value", 1*time.Minute)
			})
		})

		It("SetRange", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectSetRange("key", 1, "value")
			}, func() *redis.IntCmd {
				return client.SetRange("key", 1, "value")
			})
		})

		It("StrLen", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectStrLen("key")
			}, func() *redis.IntCmd {
				return client.StrLen("key")
			})
		})

		It("GetBit", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectGetBit("key", 1)
			}, func() *redis.IntCmd {
				return client.GetBit("key", 1)
			})
		})

		It("SetBit", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectSetBit("key", 1, 2)
			}, func() *redis.IntCmd {
				return client.SetBit("key", 1, 2)
			})
		})

		It("BitCount", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectBitCount("key", &redis.BitCount{
					Start: 1,
					End:   2,
				})
			}, func() *redis.IntCmd {
				return client.BitCount("key", &redis.BitCount{
					Start: 1,
					End:   2,
				})
			})
		})

		It("BitOpAnd", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectBitOpAnd("dest", "key1", "key2", "key3")
			}, func() *redis.IntCmd {
				return client.BitOpAnd("dest", "key1", "key2", "key3")
			})
		})

		It("BitOpOr", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectBitOpOr("dest", "key1", "key2", "key3")
			}, func() *redis.IntCmd {
				return client.BitOpOr("dest", "key1", "key2", "key3")
			})
		})

		It("BitOpXor", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectBitOpXor("dest", "key1", "key2", "key3")
			}, func() *redis.IntCmd {
				return client.BitOpXor("dest", "key1", "key2", "key3")
			})
		})

		It("BitOpNot", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectBitOpNot("dest", "key")
			}, func() *redis.IntCmd {
				return client.BitOpNot("dest", "key")
			})
		})

		It("BitPos", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectBitPos("key", 1, 2, 3)
			}, func() *redis.IntCmd {
				return client.BitPos("key", 1, 2, 3)
			})
		})

		It("BitField", func() {
			operationIntSliceCmd(mock, func() *ExpectedIntSlice {
				return mock.ExpectBitField("key", "INCRBY", "i5", 100, 1, "GET", "u4", 0)
			}, func() *redis.IntSliceCmd {
				return client.BitField("key", "INCRBY", "i5", 100, 1, "GET", "u4", 0)
			})
		})

		It("Scan", func() {
			operationScanCmd(mock, func() *ExpectedScan {
				return mock.ExpectScan(1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.Scan(1, "match", 2)
			})
		})

		It("SScan", func() {
			operationScanCmd(mock, func() *ExpectedScan {
				return mock.ExpectSScan("key", 1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.SScan("key", 1, "match", 2)
			})
		})

		It("HScan", func() {
			operationScanCmd(mock, func() *ExpectedScan {
				return mock.ExpectHScan("key", 1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.HScan("key", 1, "match", 2)
			})
		})

		It("ZScan", func() {
			operationScanCmd(mock, func() *ExpectedScan {
				return mock.ExpectZScan("key", 1, "match", 2)
			}, func() *redis.ScanCmd {
				return client.ZScan("key", 1, "match", 2)
			})
		})

		It("HDel", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectHDel("key", "field1", "field2")
			}, func() *redis.IntCmd {
				return client.HDel("key", "field1", "field2")
			})
		})

		It("HExists", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectHExists("key", "field")
			}, func() *redis.BoolCmd {
				return client.HExists("key", "field")
			})
		})

		It("HGet", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectHGet("key", "field")
			}, func() *redis.StringCmd {
				return client.HGet("key", "field")
			})
		})

		It("HGetAll", func() {
			operationStringStringMapCmd(mock, func() *ExpectedStringStringMap {
				return mock.ExpectHGetAll("key")
			}, func() *redis.StringStringMapCmd {
				return client.HGetAll("key")
			})
		})

		It("HIncrBy", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectHIncrBy("key", "field", 1)
			}, func() *redis.IntCmd {
				return client.HIncrBy("key", "field", 1)
			})
		})

		It("HIncrByFloat", func() {
			operationFloatCmd(mock, func() *ExpectedFloat {
				return mock.ExpectHIncrByFloat("key", "field", 1.1)
			}, func() *redis.FloatCmd {
				return client.HIncrByFloat("key", "field", 1.1)
			})
		})

		It("HKeys", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectHKeys("key")
			}, func() *redis.StringSliceCmd {
				return client.HKeys("key")
			})
		})

		It("HLen", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectHLen("key")
			}, func() *redis.IntCmd {
				return client.HLen("key")
			})
		})

		It("HMGet", func() {
			operationSliceCmd(mock, func() *ExpectedSlice {
				return mock.ExpectHMGet("key", "field1", "field2")
			}, func() *redis.SliceCmd {
				return client.HMGet("key", "field1", "field2")
			})
		})

		It("HSet", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectHSet("key", "field1", "value1", "field2", "value2")
			}, func() *redis.IntCmd {
				return client.HSet("key", "field1", "value1", "field2", "value2")
			})
		})

		It("HMSet", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectHMSet("key", "field1", "value1", "field2", "value2")
			}, func() *redis.BoolCmd {
				return client.HMSet("key", "field1", "value1", "field2", "value2")
			})
		})

		It("HSetNX", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectHSetNX("key", "field", "value")
			}, func() *redis.BoolCmd {
				return client.HSetNX("key", "field", "value")
			})
		})

		It("HVals", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectHVals("key")
			}, func() *redis.StringSliceCmd {
				return client.HVals("key")
			})
		})

		It("BLPop", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectBLPop(1*time.Second, "key1", "key2")
			}, func() *redis.StringSliceCmd {
				return client.BLPop(1*time.Second, "key1", "key2")
			})
		})

		It("BRPop", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectBRPop(1*time.Second, "key1", "key2")
			}, func() *redis.StringSliceCmd {
				return client.BRPop(1*time.Second, "key1", "key2")
			})
		})

		It("BRPopLPush", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectBRPopLPush("list1", "list2", 1*time.Minute)
			}, func() *redis.StringCmd {
				return client.BRPopLPush("list1", "list2", 1*time.Minute)
			})
		})

		It("LIndex", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectLIndex("key", 1)
			}, func() *redis.StringCmd {
				return client.LIndex("key", 1)
			})
		})

		It("LInsert", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectLInsert("list", "BEFORE", "World", "There")
			}, func() *redis.IntCmd {
				return client.LInsert("list", "BEFORE", "World", "There")
			})
		})

		It("LInsertBefore", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectLInsertBefore("key", "pivot", "value")
			}, func() *redis.IntCmd {
				return client.LInsertBefore("key", "pivot", "value")
			})
		})

		It("LInsertAfter", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectLInsertAfter("key", "pivot", "value")
			}, func() *redis.IntCmd {
				return client.LInsertAfter("key", "pivot", "value")
			})
		})

		It("LLen", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectLLen("key")
			}, func() *redis.IntCmd {
				return client.LLen("key")
			})
		})

		It("LPop", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectLPop("key")
			}, func() *redis.StringCmd {
				return client.LPop("key")
			})
		})

		It("LPush", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectLPush("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.LPush("key", "value1", "value2")
			})
		})

		It("LPushX", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectLPushX("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.LPushX("key", "value1", "value2")
			})
		})

		It("LRange", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectLRange("key", 1, 2)
			}, func() *redis.StringSliceCmd {
				return client.LRange("key", 1, 2)
			})
		})

		It("LRem", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectLRem("key", 2, "value")
			}, func() *redis.IntCmd {
				return client.LRem("key", 2, "value")
			})
		})

		It("LSet", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectLSet("key", 1, "value")
			}, func() *redis.StatusCmd {
				return client.LSet("key", 1, "value")
			})
		})

		It("LTrim", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectLTrim("key", 1, 2)
			}, func() *redis.StatusCmd {
				return client.LTrim("key", 1, 2)
			})
		})

		It("RPop", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectRPop("key")
			}, func() *redis.StringCmd {
				return client.RPop("key")
			})
		})

		It("RPopLPush", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectRPopLPush("key", "list")
			}, func() *redis.StringCmd {
				return client.RPopLPush("key", "list")
			})
		})

		It("RPush", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectRPush("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.RPush("key", "value1", "value2")
			})
		})

		It("RPushX", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectRPushX("key", "value1", "value2")
			}, func() *redis.IntCmd {
				return client.RPushX("key", "value1", "value2")
			})
		})

		It("SAdd", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectSAdd("key", "add")
			}, func() *redis.IntCmd {
				return client.SAdd("key", "add")
			})
		})

		It("SCard", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectSCard("key")
			}, func() *redis.IntCmd {
				return client.SCard("key")
			})
		})

		It("SDiff", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectSDiff("set1", "set2")
			}, func() *redis.StringSliceCmd {
				return client.SDiff("set1", "set2")
			})
		})

		It("SDiffStore", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectSDiffStore("set", "set1", "set2")
			}, func() *redis.IntCmd {
				return client.SDiffStore("set", "set1", "set2")
			})
		})

		It("SInter", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectSInter()
			}, func() *redis.StringSliceCmd {
				return client.SInter()
			})
		})

		It("SInterStore", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectSInterStore("set", "set1", "set2")
			}, func() *redis.IntCmd {
				return client.SInterStore("set", "set1", "set2")
			})
		})

		It("SIsMember", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectSIsMember("key", "one")
			}, func() *redis.BoolCmd {
				return client.SIsMember("key", "one")
			})
		})

		It("SMembers", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectSMembers("key")
			}, func() *redis.StringSliceCmd {
				return client.SMembers("key")
			})
		})

		It("SMembersMap", func() {
			operationStringStructMapCmd(mock, func() *ExpectedStringStructMap {
				return mock.ExpectSMembersMap("key")
			}, func() *redis.StringStructMapCmd {
				return client.SMembersMap("key")
			})
		})

		It("SMove", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectSMove("set1", "set2", "two")
			}, func() *redis.BoolCmd {
				return client.SMove("set1", "set2", "two")
			})
		})

		It("SPop", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectSPop("key")
			}, func() *redis.StringCmd {
				return client.SPop("key")
			})
		})

		It("SPopN", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectSPopN("key", 1)
			}, func() *redis.StringSliceCmd {
				return client.SPopN("key", 1)
			})
		})

		It("SRandMember", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectSRandMember("key")
			}, func() *redis.StringCmd {
				return client.SRandMember("key")
			})
		})

		It("SRandMemberN", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectSRandMemberN("key", 1)
			}, func() *redis.StringSliceCmd {
				return client.SRandMemberN("key", 1)
			})
		})

		It("SRem", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectSRem("set", "one")
			}, func() *redis.IntCmd {
				return client.SRem("set", "one")
			})
		})

		It("SUnion", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectSUnion()
			}, func() *redis.StringSliceCmd {
				return client.SUnion()
			})
		})

		It("SUnionStore", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectSUnionStore("set", "set1", "set2")
			}, func() *redis.IntCmd {
				return client.SUnionStore("set", "set1", "set2")
			})
		})

		It("XAdd", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectXAdd(&redis.XAddArgs{
					Stream: "stream",
					ID:     "1-0",
					Values: map[string]interface{}{"uno": "un"},
				})
			}, func() *redis.StringCmd {
				return client.XAdd(&redis.XAddArgs{
					Stream: "stream",
					ID:     "1-0",
					Values: map[string]interface{}{"uno": "un"},
				})
			})
		})

		It("XDel", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectXDel("stream", "1-0", "2-0", "3-0")
			}, func() *redis.IntCmd {
				return client.XDel("stream", "1-0", "2-0", "3-0")
			})
		})

		It("XLen", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectXLen("stream")
			}, func() *redis.IntCmd {
				return client.XLen("stream")
			})
		})

		It("XRange", func() {
			operationXMessageSliceCmd(mock, func() *ExpectedXMessageSlice {
				return mock.ExpectXRange("stream", "-", "+")
			}, func() *redis.XMessageSliceCmd {
				return client.XRange("stream", "-", "+")
			})
		})

		It("XRangeN", func() {
			operationXMessageSliceCmd(mock, func() *ExpectedXMessageSlice {
				return mock.ExpectXRangeN("stream", "-", "+", 2)
			}, func() *redis.XMessageSliceCmd {
				return client.XRangeN("stream", "-", "+", 2)
			})
		})

		It("XRevRange", func() {
			operationXMessageSliceCmd(mock, func() *ExpectedXMessageSlice {
				return mock.ExpectXRevRange("stream", "+", "-")
			}, func() *redis.XMessageSliceCmd {
				return client.XRevRange("stream", "+", "-")
			})
		})

		It("XRevRangeN", func() {
			operationXMessageSliceCmd(mock, func() *ExpectedXMessageSlice {
				return mock.ExpectXRevRangeN("stream", "+", "-", 2)
			}, func() *redis.XMessageSliceCmd {
				return client.XRevRangeN("stream", "+", "-", 2)
			})
		})

		It("XRead", func() {
			operationXStreamSliceCmd(mock, func() *ExpectedXStreamSlice {
				return mock.ExpectXRead(&redis.XReadArgs{
					Streams: []string{"stream", "0"},
					Count:   2,
					Block:   100 * time.Millisecond,
				})
			}, func() *redis.XStreamSliceCmd {
				return client.XRead(&redis.XReadArgs{
					Streams: []string{"stream", "0"},
					Count:   2,
					Block:   100 * time.Millisecond,
				})
			})
		})

		It("XReadStreams", func() {
			operationXStreamSliceCmd(mock, func() *ExpectedXStreamSlice {
				return mock.ExpectXReadStreams()
			}, func() *redis.XStreamSliceCmd {
				return client.XReadStreams()
			})
		})

		It("XGroupCreate", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectXGroupCreate("stream", "group", "0")
			}, func() *redis.StatusCmd {
				return client.XGroupCreate("stream", "group", "0")
			})
		})

		It("XGroupCreateMkStream", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectXGroupCreateMkStream("stream", "group", "0")
			}, func() *redis.StatusCmd {
				return client.XGroupCreateMkStream("stream", "group", "0")
			})
		})

		It("XGroupSetID", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectXGroupSetID("stream", "group", "0")
			}, func() *redis.StatusCmd {
				return client.XGroupSetID("stream", "group", "0")
			})
		})

		It("XGroupDestroy", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectXGroupDestroy("stream", "group")
			}, func() *redis.IntCmd {
				return client.XGroupDestroy("stream", "group")
			})
		})

		It("XGroupDelConsumer", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectXGroupDelConsumer("stream", "group", "consumer")
			}, func() *redis.IntCmd {
				return client.XGroupDelConsumer("stream", "group", "consumer")
			})
		})

		It("XReadGroup", func() {
			operationXStreamSliceCmd(mock, func() *ExpectedXStreamSlice {
				return mock.ExpectXReadGroup(&redis.XReadGroupArgs{
					Group:    "group",
					Consumer: "consumer",
					Streams:  []string{"stream", ">"},
				})
			}, func() *redis.XStreamSliceCmd {
				return client.XReadGroup(&redis.XReadGroupArgs{
					Group:    "group",
					Consumer: "consumer",
					Streams:  []string{"stream", ">"},
				})
			})
		})

		It("XAck", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectXAck("stream", "group", "1-0", "2-0", "4-0")
			}, func() *redis.IntCmd {
				return client.XAck("stream", "group", "1-0", "2-0", "4-0")
			})
		})

		It("XPending", func() {
			operationXPendingCmd(mock, func() *ExpectedXPending {
				return mock.ExpectXPending("stream", "group")
			}, func() *redis.XPendingCmd {
				return client.XPending("stream", "group")
			})
		})

		It("XPendingExt", func() {
			operationXPendingExtCmd(mock, func() *ExpectedXPendingExt {
				return mock.ExpectXPendingExt(&redis.XPendingExtArgs{
					Stream:   "stream",
					Group:    "group",
					Start:    "-",
					End:      "+",
					Count:    10,
					Consumer: "consumer",
				})
			}, func() *redis.XPendingExtCmd {
				return client.XPendingExt(&redis.XPendingExtArgs{
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
			operationXMessageSliceCmd(mock, func() *ExpectedXMessageSlice {
				return mock.ExpectXClaim(&redis.XClaimArgs{
					Stream:   "stream",
					Group:    "group",
					Consumer: "consumer",
					Messages: []string{"1-0", "2-0", "3-0"},
				})
			}, func() *redis.XMessageSliceCmd {
				return client.XClaim(&redis.XClaimArgs{
					Stream:   "stream",
					Group:    "group",
					Consumer: "consumer",
					Messages: []string{"1-0", "2-0", "3-0"},
				})
			})
		})

		It("XClaimJustID", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectXClaimJustID(&redis.XClaimArgs{
					Stream:   "stream",
					Group:    "group",
					Consumer: "consumer",
					Messages: []string{"1-0", "2-0", "3-0"},
				})
			}, func() *redis.StringSliceCmd {
				return client.XClaimJustID(&redis.XClaimArgs{
					Stream:   "stream",
					Group:    "group",
					Consumer: "consumer",
					Messages: []string{"1-0", "2-0", "3-0"},
				})
			})
		})

		It("XTrim", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectXTrim("stream", 0)
			}, func() *redis.IntCmd {
				return client.XTrim("stream", 0)
			})
		})

		It("XTrimApprox", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectXTrimApprox("stream", 0)
			}, func() *redis.IntCmd {
				return client.XTrimApprox("stream", 0)
			})
		})

		It("XInfoGroups", func() {
			operationXInfoGroupsCmd(mock, func() *ExpectedXInfoGroups {
				return mock.ExpectXInfoGroups("key")
			}, func() *redis.XInfoGroupsCmd {
				return client.XInfoGroups("key")
			})
		})

		It("BZPopMax", func() {
			operationZWithKeyCmd(mock, func() *ExpectedZWithKey {
				return mock.ExpectBZPopMax(0, "zset1", "zset2")
			}, func() *redis.ZWithKeyCmd {
				return client.BZPopMax(0, "zset1", "zset2")
			})
		})

		It("BZPopMin", func() {
			operationZWithKeyCmd(mock, func() *ExpectedZWithKey {
				return mock.ExpectBZPopMin(0, "zset1", "zset2")
			}, func() *redis.ZWithKeyCmd {
				return client.BZPopMin(0, "zset1", "zset2")
			})
		})

		It("ZAdd", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZAdd("zset", &redis.Z{
					Member: "a",
					Score:  1,
				})
			}, func() *redis.IntCmd {
				return client.ZAdd("zset", &redis.Z{
					Member: "a",
					Score:  1,
				})
			})
		})

		It("ZAddNX", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZAddNX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.IntCmd {
				return client.ZAddNX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZAddXX", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZAddXX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.IntCmd {
				return client.ZAddXX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZAddCh", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZAddCh("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.IntCmd {
				return client.ZAddCh("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZAddNXCh", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZAddNXCh("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.IntCmd {
				return client.ZAddNXCh("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZAddXXCh", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZAddXXCh("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.IntCmd {
				return client.ZAddXXCh("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZIncr", func() {
			operationFloatCmd(mock, func() *ExpectedFloat {
				return mock.ExpectZIncr("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.FloatCmd {
				return client.ZIncr("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZIncrNX", func() {
			operationFloatCmd(mock, func() *ExpectedFloat {
				return mock.ExpectZIncrNX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.FloatCmd {
				return client.ZIncrNX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZIncrXX", func() {
			operationFloatCmd(mock, func() *ExpectedFloat {
				return mock.ExpectZIncrXX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			}, func() *redis.FloatCmd {
				return client.ZIncrXX("zset", &redis.Z{
					Score:  1,
					Member: "one",
				})
			})
		})

		It("ZCard", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZCard("key")
			}, func() *redis.IntCmd {
				return client.ZCard("key")
			})
		})

		It("ZCount", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZCount("zset", "-inf", "+inf")
			}, func() *redis.IntCmd {
				return client.ZCount("zset", "-inf", "+inf")
			})
		})

		It("ZLexCount", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZLexCount("zset", "-", "+")
			}, func() *redis.IntCmd {
				return client.ZLexCount("zset", "-", "+")
			})
		})

		It("ZIncrBy", func() {
			operationFloatCmd(mock, func() *ExpectedFloat {
				return mock.ExpectZIncrBy("zset", 2, "one")
			}, func() *redis.FloatCmd {
				return client.ZIncrBy("zset", 2, "one")
			})
		})

		It("ZInterStore", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZInterStore("out", &redis.ZStore{
					Keys:    []string{"zset1", "zset2"},
					Weights: []float64{2, 3},
				})
			}, func() *redis.IntCmd {
				return client.ZInterStore("out", &redis.ZStore{
					Keys:    []string{"zset1", "zset2"},
					Weights: []float64{2, 3},
				})
			})
		})

		It("ZPopMax", func() {
			operationZSliceCmd(mock, func() *ExpectedZSlice {
				return mock.ExpectZPopMax("key")
			}, func() *redis.ZSliceCmd {
				return client.ZPopMax("key")
			})
		})

		It("ZPopMin", func() {
			operationZSliceCmd(mock, func() *ExpectedZSlice {
				return mock.ExpectZPopMin("key")
			}, func() *redis.ZSliceCmd {
				return client.ZPopMin("key")
			})
		})

		It("ZRange", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectZRange("zset", 0, -1)
			}, func() *redis.StringSliceCmd {
				return client.ZRange("zset", 0, -1)
			})
		})

		It("ZRangeWithScores", func() {
			operationZSliceCmd(mock, func() *ExpectedZSlice {
				return mock.ExpectZRangeWithScores("zset", 0, -1)
			}, func() *redis.ZSliceCmd {
				return client.ZRangeWithScores("zset", 0, -1)
			})
		})

		It("ZRangeByScore", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectZRangeByScore("zset", &redis.ZRangeBy{
					Min: "-inf",
					Max: "+inf",
				})
			}, func() *redis.StringSliceCmd {
				return client.ZRangeByScore("zset", &redis.ZRangeBy{
					Min: "-inf",
					Max: "+inf",
				})
			})
		})

		It("ZRangeByLex", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectZRangeByLex("zset", &redis.ZRangeBy{
					Min: "-",
					Max: "+",
				})
			}, func() *redis.StringSliceCmd {
				return client.ZRangeByLex("zset", &redis.ZRangeBy{
					Min: "-",
					Max: "+",
				})
			})
		})

		It("ZRangeByScoreWithScores", func() {
			operationZSliceCmd(mock, func() *ExpectedZSlice {
				return mock.ExpectZRangeByScoreWithScores("zset", &redis.ZRangeBy{
					Min: "-inf",
					Max: "+inf",
				})
			}, func() *redis.ZSliceCmd {
				return client.ZRangeByScoreWithScores("zset", &redis.ZRangeBy{
					Min: "-inf",
					Max: "+inf",
				})
			})
		})

		It("ZRank", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZRank("zset", "three")
			}, func() *redis.IntCmd {
				return client.ZRank("zset", "three")
			})
		})

		It("ZRem", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZRem("zset", "two")
			}, func() *redis.IntCmd {
				return client.ZRem("zset", "two")
			})
		})

		It("ZRemRangeByRank", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZRemRangeByRank("key", 1, 2)
			}, func() *redis.IntCmd {
				return client.ZRemRangeByRank("key", 1, 2)
			})
		})

		It("ZRemRangeByScore", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZRemRangeByScore("zset", "-inf", "(2")
			}, func() *redis.IntCmd {
				return client.ZRemRangeByScore("zset", "-inf", "(2")
			})
		})

		It("ZRemRangeByLex", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZRemRangeByLex("zset", "[alpha", "[omega")
			}, func() *redis.IntCmd {
				return client.ZRemRangeByLex("zset", "[alpha", "[omega")
			})
		})

		It("ZRevRange", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectZRevRange("zset", 0, -1)
			}, func() *redis.StringSliceCmd {
				return client.ZRevRange("zset", 0, -1)
			})
		})

		It("ZRevRangeWithScores", func() {
			operationZSliceCmd(mock, func() *ExpectedZSlice {
				return mock.ExpectZRevRangeWithScores("zset", 0, -1)
			}, func() *redis.ZSliceCmd {
				return client.ZRevRangeWithScores("zset", 0, -1)
			})
		})

		It("ZRevRangeByScore", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectZRevRangeByScore("zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			}, func() *redis.StringSliceCmd {
				return client.ZRevRangeByScore("zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			})
		})

		It("ZRevRangeByLex", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectZRevRangeByLex("zset", &redis.ZRangeBy{Max: "+", Min: "-"})
			}, func() *redis.StringSliceCmd {
				return client.ZRevRangeByLex("zset", &redis.ZRangeBy{Max: "+", Min: "-"})
			})
		})

		It("ZRevRangeByScoreWithScores", func() {
			operationZSliceCmd(mock, func() *ExpectedZSlice {
				return mock.ExpectZRevRangeByScoreWithScores("zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			}, func() *redis.ZSliceCmd {
				return client.ZRevRangeByScoreWithScores("zset", &redis.ZRangeBy{Max: "+inf", Min: "-inf"})
			})
		})

		It("ZRevRank", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZRevRank("key", "member")
			}, func() *redis.IntCmd {
				return client.ZRevRank("key", "member")
			})
		})

		It("ZScore", func() {
			operationFloatCmd(mock, func() *ExpectedFloat {
				return mock.ExpectZScore("key", "member")
			}, func() *redis.FloatCmd {
				return client.ZScore("key", "member")
			})
		})

		It("ZUnionStore", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectZUnionStore("out", &redis.ZStore{
					Keys:    []string{"zset1", "zset2"},
					Weights: []float64{2, 3},
				})
			}, func() *redis.IntCmd {
				return client.ZUnionStore("out", &redis.ZStore{
					Keys:    []string{"zset1", "zset2"},
					Weights: []float64{2, 3},
				})
			})
		})

		It("PFAdd", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectPFAdd("hll1", "1", "2", "3", "4", "5")
			}, func() *redis.IntCmd {
				return client.PFAdd("hll1", "1", "2", "3", "4", "5")
			})
		})

		It("PFCount", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectPFCount("hll1", "hll2")
			}, func() *redis.IntCmd {
				return client.PFCount("hll1", "hll2")
			})
		})

		It("PFMerge", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectPFMerge("hllMerged", "hll1", "hll2")
			}, func() *redis.StatusCmd {
				return client.PFMerge("hllMerged", "hll1", "hll2")
			})
		})

		It("BgRewriteAOF", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectBgRewriteAOF()
			}, func() *redis.StatusCmd {
				return client.BgRewriteAOF()
			})
		})

		It("BgSave", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectBgSave()
			}, func() *redis.StatusCmd {
				return client.BgSave()
			})
		})

		It("ClientKill", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClientKill("1.1.1.1:1111")
			}, func() *redis.StatusCmd {
				return client.ClientKill("1.1.1.1:1111")
			})
		})

		It("ClientKillByFilter", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectClientKillByFilter("11.11.11.11:1234")
			}, func() *redis.IntCmd {
				return client.ClientKillByFilter("11.11.11.11:1234")
			})
		})

		It("ClientList", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectClientList()
			}, func() *redis.StringCmd {
				return client.ClientList()
			})
		})

		It("ClientPause", func() {
			operationBoolCmd(mock, func() *ExpectedBool {
				return mock.ExpectClientPause(1 * time.Minute)
			}, func() *redis.BoolCmd {
				return client.ClientPause(1 * time.Minute)
			})
		})

		It("ClientID", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectClientID()
			}, func() *redis.IntCmd {
				return client.ClientID()
			})
		})

		It("ConfigGet", func() {
			operationSliceCmd(mock, func() *ExpectedSlice {
				return mock.ExpectConfigGet("*")
			}, func() *redis.SliceCmd {
				return client.ConfigGet("*")
			})
		})

		It("ConfigResetStat", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectConfigResetStat()
			}, func() *redis.StatusCmd {
				return client.ConfigResetStat()
			})
		})

		It("ConfigSet", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectConfigSet("maxmemory", "1024M")
			}, func() *redis.StatusCmd {
				return client.ConfigSet("maxmemory", "1024M")
			})
		})

		It("ConfigRewrite", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectConfigRewrite()
			}, func() *redis.StatusCmd {
				return client.ConfigRewrite()
			})
		})

		It("DBSize", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectDBSize()
			}, func() *redis.IntCmd {
				return client.DBSize()
			})
		})

		It("FlushAll", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectFlushAll()
			}, func() *redis.StatusCmd {
				return client.FlushAll()
			})
		})

		It("FlushAllAsync", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectFlushAllAsync()
			}, func() *redis.StatusCmd {
				return client.FlushAllAsync()
			})
		})

		It("FlushDB", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectFlushDB()
			}, func() *redis.StatusCmd {
				return client.FlushDB()
			})
		})

		It("FlushDBAsync", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectFlushDBAsync()
			}, func() *redis.StatusCmd {
				return client.FlushDBAsync()
			})
		})

		It("Info", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectInfo()
			}, func() *redis.StringCmd {
				return client.Info()
			})
		})

		It("LastSave", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectLastSave()
			}, func() *redis.IntCmd {
				return client.LastSave()
			})
		})

		It("Save", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectSave()
			}, func() *redis.StatusCmd {
				return client.Save()
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
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectSlaveOf("localhost", "8888")
			}, func() *redis.StatusCmd {
				return client.SlaveOf("localhost", "8888")
			})
		})

		It("Time", func() {
			operationTimeCmd(mock, func() *ExpectedTime {
				return mock.ExpectTime()
			}, func() *redis.TimeCmd {
				return client.Time()
			})
		})

		It("DebugObject", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectDebugObject("foo")
			}, func() *redis.StringCmd {
				return client.DebugObject("foo")
			})
		})

		It("ReadOnly", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectReadOnly()
			}, func() *redis.StatusCmd {
				return client.ReadOnly()
			})
		})

		It("ReadWrite", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectReadWrite()
			}, func() *redis.StatusCmd {
				return client.ReadWrite()
			})
		})

		It("MemoryUsage", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectMemoryUsage("foo")
			}, func() *redis.IntCmd {
				return client.MemoryUsage("foo")
			})
		})

		It("Eval", func() {
			operationCmdCmd(mock, func() *ExpectedCmd {
				return mock.ExpectEval("return {KEYS[1],ARGV[1]}", []string{"key"}, "hello")
			}, func() *redis.Cmd {
				return client.Eval("return {KEYS[1],ARGV[1]}", []string{"key"}, "hello")
			})
		})

		It("EvalSha", func() {
			operationCmdCmd(mock, func() *ExpectedCmd {
				return mock.ExpectEvalSha("sha", []string{"key1", "key2"}, "args1", "args2")
			}, func() *redis.Cmd {
				return client.EvalSha("sha", []string{"key1", "key2"}, "args1", "args2")
			})
		})

		It("ScriptExists", func() {
			operationBoolSliceCmd(mock, func() *ExpectedBoolSlice {
				return mock.ExpectScriptExists()
			}, func() *redis.BoolSliceCmd {
				return client.ScriptExists()
			})
		})

		It("ScriptFlush", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectScriptFlush()
			}, func() *redis.StatusCmd {
				return client.ScriptFlush()
			})
		})

		It("ScriptKill", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectScriptKill()
			}, func() *redis.StatusCmd {
				return client.ScriptKill()
			})
		})

		It("ScriptLoad", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectScriptLoad("script")
			}, func() *redis.StringCmd {
				return client.ScriptLoad("script")
			})
		})

		It("Publish", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectPublish("channel", "message")
			}, func() *redis.IntCmd {
				return client.Publish("channel", "message")
			})
		})

		It("PubSubChannels", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectPubSubChannels("pattern")
			}, func() *redis.StringSliceCmd {
				return client.PubSubChannels("pattern")
			})
		})

		It("PubSubNumSub", func() {
			operationStringIntMapCmd(mock, func() *ExpectedStringIntMap {
				return mock.ExpectPubSubNumSub()
			}, func() *redis.StringIntMapCmd {
				return client.PubSubNumSub()
			})
		})

		It("PubSubNumPat", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectPubSubNumPat()
			}, func() *redis.IntCmd {
				return client.PubSubNumPat()
			})
		})

		It("ClusterSlots", func() {
			operationClusterSlotsCmd(mock, func() *ExpectedClusterSlots {
				return mock.ExpectClusterSlots()
			}, func() *redis.ClusterSlotsCmd {
				return client.ClusterSlots()
			})
		})

		It("ClusterNodes", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectClusterNodes()
			}, func() *redis.StringCmd {
				return client.ClusterNodes()
			})
		})

		It("ClusterMeet", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterMeet("1.1.1.1", "1")
			}, func() *redis.StatusCmd {
				return client.ClusterMeet("1.1.1.1", "1")
			})
		})

		It("ClusterForget", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterForget("id")
			}, func() *redis.StatusCmd {
				return client.ClusterForget("id")
			})
		})

		It("ClusterReplicate", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterReplicate("id")
			}, func() *redis.StatusCmd {
				return client.ClusterReplicate("id")
			})
		})

		It("ClusterResetSoft", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterResetSoft()
			}, func() *redis.StatusCmd {
				return client.ClusterResetSoft()
			})
		})

		It("ClusterResetHard", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterResetHard()
			}, func() *redis.StatusCmd {
				return client.ClusterResetHard()
			})
		})

		It("ClusterInfo", func() {
			operationStringCmd(mock, func() *ExpectedString {
				return mock.ExpectClusterInfo()
			}, func() *redis.StringCmd {
				return client.ClusterInfo()
			})
		})

		It("ClusterKeySlot", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectClusterKeySlot("key")
			}, func() *redis.IntCmd {
				return client.ClusterKeySlot("key")
			})
		})

		It("ClusterGetKeysInSlot", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectClusterGetKeysInSlot(1, 2)
			}, func() *redis.StringSliceCmd {
				return client.ClusterGetKeysInSlot(1, 2)
			})
		})

		It("ClusterCountFailureReports", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectClusterCountFailureReports("id")
			}, func() *redis.IntCmd {
				return client.ClusterCountFailureReports("id")
			})
		})

		It("ClusterCountKeysInSlot", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectClusterCountKeysInSlot(1)
			}, func() *redis.IntCmd {
				return client.ClusterCountKeysInSlot(1)
			})
		})

		It("ClusterDelSlots", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterDelSlots()
			}, func() *redis.StatusCmd {
				return client.ClusterDelSlots()
			})
		})

		It("ClusterDelSlotsRange", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterDelSlotsRange(1, 2)
			}, func() *redis.StatusCmd {
				return client.ClusterDelSlotsRange(1, 2)
			})
		})

		It("ClusterSaveConfig", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterSaveConfig()
			}, func() *redis.StatusCmd {
				return client.ClusterSaveConfig()
			})
		})

		It("ClusterSlaves", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectClusterSlaves("id")
			}, func() *redis.StringSliceCmd {
				return client.ClusterSlaves("id")
			})
		})

		It("ClusterFailover", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterFailover()
			}, func() *redis.StatusCmd {
				return client.ClusterFailover()
			})
		})

		It("ClusterAddSlots", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterAddSlots()
			}, func() *redis.StatusCmd {
				return client.ClusterAddSlots()
			})
		})

		It("ClusterAddSlotsRange", func() {
			operationStatusCmd(mock, func() *ExpectedStatus {
				return mock.ExpectClusterAddSlotsRange(1, 2)
			}, func() *redis.StatusCmd {
				return client.ClusterAddSlotsRange(1, 2)
			})
		})

		It("GeoAdd", func() {
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectGeoAdd("Sicily",
					&redis.GeoLocation{Longitude: 13.361389, Latitude: 38.115556, Name: "Palermo"},
					&redis.GeoLocation{Longitude: 15.087269, Latitude: 37.502669, Name: "Tokyo"},
				)
			}, func() *redis.IntCmd {
				return client.GeoAdd("Sicily",
					&redis.GeoLocation{Longitude: 13.361389, Latitude: 38.115556, Name: "Palermo"},
					&redis.GeoLocation{Longitude: 15.087269, Latitude: 37.502669, Name: "Tokyo"})
			})
		})

		It("GeoPos", func() {
			operationGeoPosCmd(mock, func() *ExpectedGeoPos {
				return mock.ExpectGeoPos("Sicily", "Palermo", "Catania", "NonExisting")
			}, func() *redis.GeoPosCmd {
				return client.GeoPos("Sicily", "Palermo", "Catania", "NonExisting")
			})
		})

		It("GeoRadius", func() {
			operationGeoLocationCmd(mock, func() *ExpectedGeoLocation {
				return mock.ExpectGeoRadius("Sicily", 15, 37, &redis.GeoRadiusQuery{
					Radius:      200,
					Unit:        "km",
					WithGeoHash: true,
					WithCoord:   true,
					WithDist:    true,
					Count:       2,
					Sort:        "ASC",
				})
			}, func() *redis.GeoLocationCmd {
				return client.GeoRadius("Sicily", 15, 37, &redis.GeoRadiusQuery{
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
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectGeoRadiusStore("Sicily", 15, 37, &redis.GeoRadiusQuery{
					Radius: 200,
					Store:  "result",
				})
			}, func() *redis.IntCmd {
				return client.GeoRadiusStore("Sicily", 15, 37, &redis.GeoRadiusQuery{
					Radius: 200,
					Store:  "result",
				})
			})
		})

		It("GeoRadiusByMember", func() {
			operationGeoLocationCmd(mock, func() *ExpectedGeoLocation {
				return mock.ExpectGeoRadiusByMember("Sicily", "Catania", &redis.GeoRadiusQuery{
					Radius:      200,
					Unit:        "km",
					WithGeoHash: true,
					WithCoord:   true,
					WithDist:    true,
					Count:       2,
					Sort:        "ASC",
				})
			}, func() *redis.GeoLocationCmd {
				return client.GeoRadiusByMember("Sicily", "Catania", &redis.GeoRadiusQuery{
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
			operationIntCmd(mock, func() *ExpectedInt {
				return mock.ExpectGeoRadiusByMemberStore("key", "member", &redis.GeoRadiusQuery{
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
				return client.GeoRadiusByMemberStore("key", "member", &redis.GeoRadiusQuery{
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
			operationFloatCmd(mock, func() *ExpectedFloat {
				return mock.ExpectGeoDist("Sicily", "Palermo", "Catania", "km")
			}, func() *redis.FloatCmd {
				return client.GeoDist("Sicily", "Palermo", "Catania", "km")
			})
		})

		It("GeoHash", func() {
			operationStringSliceCmd(mock, func() *ExpectedStringSlice {
				return mock.ExpectGeoHash("Sicily", "Palermo", "Catania")
			}, func() *redis.StringSliceCmd {
				return client.GeoHash("Sicily", "Palermo", "Catania")
			})
		})
	})
})

func operationStringCmd(mock ClientMock, expected func() *ExpectedString, actual func() *redis.StringCmd) {
	var (
		setErr = errors.New("string cmd error")
		str    string
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	str, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(str).To(Equal(""))

	mock.ClearExpect()
	expected()
	str, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(str).To(Equal(""))

	mock.ClearExpect()
	expected().SetVal("value")
	str, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(str).To(Equal("value"))
}

func operationStatusCmd(mock ClientMock, expected func() *ExpectedStatus, actual func() *redis.StatusCmd) {
	var (
		setErr = errors.New("status cmd error")
		str    string
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	str, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(str).To(Equal(""))

	mock.ClearExpect()
	expected()
	str, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(str).To(Equal(""))

	mock.ClearExpect()
	expected().SetVal("OK")
	str, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(str).To(Equal("OK"))
}

func operationIntCmd(mock ClientMock, expected func() *ExpectedInt, actual func() *redis.IntCmd) {
	var (
		setErr = errors.New("int cmd error")
		val    int64
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(int64(0)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(int64(0)))

	mock.ClearExpect()
	expected().SetVal(1024)
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(int64(1024)))
}

func operationBoolCmd(mock ClientMock, expected func() *ExpectedBool, actual func() *redis.BoolCmd) {
	var (
		setErr = errors.New("bool cmd error")
		val    bool
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(BeFalse())

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(BeFalse())

	mock.ClearExpect()
	expected().SetVal(true)
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(BeTrue())
}

func operationStringSliceCmd(mock ClientMock, expected func() *ExpectedStringSlice, actual func() *redis.StringSliceCmd) {
	var (
		setErr = errors.New("string slice cmd error")
		val    []string
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]string(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]string(nil)))

	mock.ClearExpect()
	expected().SetVal([]string{"redis", "move"})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]string{"redis", "move"}))
}

func operationDurationCmd(mock ClientMock, expected func() *ExpectedDuration, actual func() *redis.DurationCmd) {
	var (
		setErr = errors.New("duration cmd error")
		val    time.Duration
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(time.Duration(0)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(time.Duration(0)))

	mock.ClearExpect()
	expected().SetVal(2 * time.Hour)
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(2 * time.Hour))
}

func operationSliceCmd(mock ClientMock, expected func() *ExpectedSlice, actual func() *redis.SliceCmd) {
	var (
		setErr = errors.New("slice cmd error")
		val    []interface{}
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]interface{}(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]interface{}(nil)))

	mock.ClearExpect()
	expected().SetVal([]interface{}{"mock", "slice"})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]interface{}{"mock", "slice"}))
}

func operationFloatCmd(mock ClientMock, expected func() *ExpectedFloat, actual func() *redis.FloatCmd) {
	var (
		setErr = errors.New("float cmd error")
		val    float64
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(float64(0)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(float64(0)))

	mock.ClearExpect()
	expected().SetVal(1)
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(float64(1)))
}

func operationIntSliceCmd(mock ClientMock, expected func() *ExpectedIntSlice, actual func() *redis.IntSliceCmd) {
	var (
		setErr = errors.New("int slice cmd error")
		val    []int64
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]int64(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]int64(nil)))

	mock.ClearExpect()
	expected().SetVal([]int64{1, 2, 3, 4})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]int64{1, 2, 3, 4}))
}

func operationScanCmd(mock ClientMock, expected func() *ExpectedScan, actual func() *redis.ScanCmd) {
	var (
		setErr = errors.New("scan cmd error")
		page   []string
		cursor uint64
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	page, cursor, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(page).To(Equal([]string(nil)))
	Expect(cursor).To(Equal(uint64(0)))

	mock.ClearExpect()
	expected()
	page, cursor, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(page).To(Equal([]string(nil)))
	Expect(cursor).To(Equal(uint64(0)))

	mock.ClearExpect()
	expected().SetVal([]string{"key1", "key2", "key3"}, 5)
	page, cursor, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(page).To(Equal([]string{"key1", "key2", "key3"}))
	Expect(cursor).To(Equal(uint64(5)))
}

func operationStringStringMapCmd(mock ClientMock, expected func() *ExpectedStringStringMap, actual func() *redis.StringStringMapCmd) {
	var (
		setErr = errors.New("string string map cmd error")
		val    map[string]string
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(map[string]string(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(map[string]string(nil)))

	mock.ClearExpect()
	expected().SetVal(map[string]string{"key": "value", "key2": "value2"})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(map[string]string{"key": "value", "key2": "value2"}))
}

func operationStringStructMapCmd(mock ClientMock, expected func() *ExpectedStringStructMap, actual func() *redis.StringStructMapCmd) {
	var (
		setErr = errors.New("string struct map cmd error")
		val    map[string]struct{}
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(map[string]struct{}(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(map[string]struct{}(nil)))

	mock.ClearExpect()
	expected().SetVal([]string{"key1", "key2"})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(map[string]struct{}{"key1": {}, "key2": {}}))
}

func operationXMessageSliceCmd(mock ClientMock, expected func() *ExpectedXMessageSlice, actual func() *redis.XMessageSliceCmd) {
	var (
		setErr = errors.New("x message slice cmd error")
		val    []redis.XMessage
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.XMessage(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.XMessage(nil)))

	mock.ClearExpect()
	expected().SetVal([]redis.XMessage{
		{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
		{ID: "3-0", Values: map[string]interface{}{"tres": "troix"}},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]redis.XMessage{
		{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
		{ID: "3-0", Values: map[string]interface{}{"tres": "troix"}},
	}))
}

func operationXStreamSliceCmd(mock ClientMock, expected func() *ExpectedXStreamSlice, actual func() *redis.XStreamSliceCmd) {
	var (
		setErr = errors.New("x stream slice cmd error")
		val    []redis.XStream
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.XStream(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.XStream(nil)))

	mock.ClearExpect()
	expected().SetVal([]redis.XStream{{
		Stream: "stream",
		Messages: []redis.XMessage{
			{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
			{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
		}},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]redis.XStream{{
		Stream: "stream",
		Messages: []redis.XMessage{
			{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
			{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
		}},
	}))
}

func operationXPendingCmd(mock ClientMock, expected func() *ExpectedXPending, actual func() *redis.XPendingCmd) {
	var (
		setErr = errors.New("x pending cmd error")
		val    *redis.XPending
		valNil *redis.XPending
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(valNil))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(valNil))

	mock.ClearExpect()
	expected().SetVal(&redis.XPending{
		Count:     3,
		Lower:     "1-0",
		Higher:    "3-0",
		Consumers: map[string]int64{"consumer": 3},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(&redis.XPending{
		Count:     3,
		Lower:     "1-0",
		Higher:    "3-0",
		Consumers: map[string]int64{"consumer": 3},
	}))
}

func operationXPendingExtCmd(mock ClientMock, expected func() *ExpectedXPendingExt, actual func() *redis.XPendingExtCmd) {
	var (
		setErr = errors.New("x pending ext cmd error")
		val    []redis.XPendingExt
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.XPendingExt(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.XPendingExt(nil)))

	mock.ClearExpect()
	expected().SetVal([]redis.XPendingExt{
		{ID: "1-0", Consumer: "consumer", Idle: 0, RetryCount: 1},
		{ID: "2-0", Consumer: "consumer", Idle: 0, RetryCount: 1},
		{ID: "3-0", Consumer: "consumer", Idle: 0, RetryCount: 1},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]redis.XPendingExt{
		{ID: "1-0", Consumer: "consumer", Idle: 0, RetryCount: 1},
		{ID: "2-0", Consumer: "consumer", Idle: 0, RetryCount: 1},
		{ID: "3-0", Consumer: "consumer", Idle: 0, RetryCount: 1},
	}))
}

func operationXInfoGroupsCmd(mock ClientMock, expected func() *ExpectedXInfoGroups, actual func() *redis.XInfoGroupsCmd) {
	var (
		setErr = errors.New("x info group cmd error")
		val    []redis.XInfoGroups
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.XInfoGroups(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.XInfoGroups(nil)))

	mock.ClearExpect()
	expected().SetVal([]redis.XInfoGroups{
		{Name: "name1", Consumers: 1, Pending: 2, LastDeliveredID: "last1"},
		{Name: "name2", Consumers: 1, Pending: 2, LastDeliveredID: "last2"},
		{Name: "name3", Consumers: 1, Pending: 2, LastDeliveredID: "last2"},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]redis.XInfoGroups{
		{Name: "name1", Consumers: 1, Pending: 2, LastDeliveredID: "last1"},
		{Name: "name2", Consumers: 1, Pending: 2, LastDeliveredID: "last2"},
		{Name: "name3", Consumers: 1, Pending: 2, LastDeliveredID: "last2"},
	}))
}

func operationZWithKeyCmd(mock ClientMock, expected func() *ExpectedZWithKey, actual func() *redis.ZWithKeyCmd) {
	var (
		setErr = errors.New("z with key cmd error")
		val    *redis.ZWithKey
		valNil *redis.ZWithKey
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(valNil))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(valNil))

	mock.ClearExpect()
	expected().SetVal(&redis.ZWithKey{
		Z: redis.Z{
			Score:  3,
			Member: "three",
		},
		Key: "zset1",
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(&redis.ZWithKey{
		Z: redis.Z{
			Score:  3,
			Member: "three",
		},
		Key: "zset1",
	}))
}

func operationZSliceCmd(mock ClientMock, expected func() *ExpectedZSlice, actual func() *redis.ZSliceCmd) {
	var (
		setErr = errors.New("z slice cmd error")
		val    []redis.Z
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.Z(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.Z(nil)))

	mock.ClearExpect()
	expected().SetVal([]redis.Z{{
		Score:  5,
		Member: "one",
	}, {
		Score:  10,
		Member: "two",
	}})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]redis.Z{{
		Score:  5,
		Member: "one",
	}, {
		Score:  10,
		Member: "two",
	}}))
}

func operationTimeCmd(mock ClientMock, expected func() *ExpectedTime, actual func() *redis.TimeCmd) {
	var (
		setErr = errors.New("time cmd error")
		val    time.Time
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(time.Time{}))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(time.Time{}))

	mock.ClearExpect()
	now := time.Now()
	expected().SetVal(now)
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(now))
}

func operationCmdCmd(mock ClientMock, expected func() *ExpectedCmd, actual func() *redis.Cmd) {
	var (
		setErr = errors.New("cmd error")
		val    interface{}
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(BeNil())

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(BeNil())

	mock.ClearExpect()
	expected().SetVal(interface{}(1024))
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(interface{}(1024)))
}

func operationBoolSliceCmd(mock ClientMock, expected func() *ExpectedBoolSlice, actual func() *redis.BoolSliceCmd) {
	var (
		setErr = errors.New("bool slice cmd error")
		val    []bool
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]bool(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]bool(nil)))

	mock.ClearExpect()
	expected().SetVal([]bool{true, false, true})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]bool{true, false, true}))
}

func operationStringIntMapCmd(mock ClientMock, expected func() *ExpectedStringIntMap, actual func() *redis.StringIntMapCmd) {
	var (
		setErr = errors.New("string int map cmd error")
		val    map[string]int64
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(map[string]int64(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(map[string]int64(nil)))

	mock.ClearExpect()
	expected().SetVal(map[string]int64{"key": 1, "key2": 2})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(map[string]int64{"key": 1, "key2": 2}))
}

func operationClusterSlotsCmd(mock ClientMock, expected func() *ExpectedClusterSlots, actual func() *redis.ClusterSlotsCmd) {
	var (
		setErr = errors.New("cluster slots cmd error")
		val    []redis.ClusterSlot
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.ClusterSlot(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.ClusterSlot(nil)))

	mock.ClearExpect()
	expected().SetVal([]redis.ClusterSlot{
		{Start: 1, End: 2, Nodes: []redis.ClusterNode{
			{ID: "1", Addr: "1.1.1.1"},
			{ID: "2", Addr: "2.2.2.2"},
		}},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]redis.ClusterSlot{
		{Start: 1, End: 2, Nodes: []redis.ClusterNode{
			{ID: "1", Addr: "1.1.1.1"},
			{ID: "2", Addr: "2.2.2.2"},
		}},
	}))
}

func operationGeoLocationCmd(mock ClientMock, expected func() *ExpectedGeoLocation, actual func() *redis.GeoLocationCmd) {
	var (
		setErr = errors.New("geo location cmd error")
		val    []redis.GeoLocation
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.GeoLocation(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.GeoLocation(nil)))

	mock.ClearExpect()
	expected().SetVal([]redis.GeoLocation{
		{Longitude: 13.361389, Latitude: 38.115556, Name: "Palermo"},
		{Longitude: 15.087269, Latitude: 37.502669, Name: "Catania"},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]redis.GeoLocation{
		{Longitude: 13.361389, Latitude: 38.115556, Name: "Palermo"},
		{Longitude: 15.087269, Latitude: 37.502669, Name: "Catania"},
	}))
}

func operationGeoPosCmd(mock ClientMock, expected func() *ExpectedGeoPos, actual func() *redis.GeoPosCmd) {
	var (
		setErr = errors.New("geo pos cmd error")
		val    []*redis.GeoPos
		err    error
	)

	mock.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]*redis.GeoPos(nil)))

	mock.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]*redis.GeoPos(nil)))

	mock.ClearExpect()
	expected().SetVal([]*redis.GeoPos{
		{
			Longitude: 13.361389338970184,
			Latitude:  38.1155563954963,
		},
		{
			Longitude: 15.087267458438873,
			Latitude:  37.50266842333162,
		},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]*redis.GeoPos{
		{
			Longitude: 13.361389338970184,
			Latitude:  38.1155563954963,
		},
		{
			Longitude: 15.087267458438873,
			Latitude:  37.50266842333162,
		},
	}))
}
