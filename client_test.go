package redismock

import (
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
)

var _ = Describe("Client", func() {
	var (
		client     *redis.Client
		clientMock ClientMock
	)

	BeforeEach(func() {
		client, clientMock = NewClientMock()
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

		AfterEach(func() {
			hasUnexpectedCall, unexpectedCalls := clientMock.UnexpectedCallsWereCalled()
			Expect(hasUnexpectedCall).To(BeFalse())
			Expect(unexpectedCalls).To(BeNil())
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

		AfterEach(func() {
			hasUnexpectedCall, unexpectedCalls := clientMock.UnexpectedCallsWereCalled()
			Expect(hasUnexpectedCall).To(BeFalse())
			Expect(unexpectedCalls).To(BeNil())
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

		AfterEach(func() {
			hasUnexpectedCall, unexpectedCalls := clientMock.UnexpectedCallsWereCalled()
			Expect(hasUnexpectedCall).To(BeTrue())
			Expect(unexpectedCalls).ShouldNot(BeNil())
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

			hasUnexpectedCall, unexpectedCalls := clientMock.UnexpectedCallsWereCalled()
			Expect(hasUnexpectedCall).To(BeFalse())
			Expect(unexpectedCalls).To(BeNil())
		})

		It("surplus", func() {
			_ = client.Get(ctx, "key")

			set := client.Set(ctx, "key", "1", 1*time.Second)
			Expect(set.Err()).NotTo(HaveOccurred())
			Expect(set.Val()).To(Equal("OK"))

			Expect(clientMock.ExpectationsWereMet()).To(HaveOccurred())

			_ = client.Get(ctx, "key")
			Expect(clientMock.ExpectationsWereMet()).To(HaveOccurred())

			hasUnexpectedCall, unexpectedCalls := clientMock.UnexpectedCallsWereCalled()
			Expect(hasUnexpectedCall).To(BeFalse())
			Expect(unexpectedCalls).To(BeNil())

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

			hasUnexpectedCall, unexpectedCalls := clientMock.UnexpectedCallsWereCalled()
			Expect(hasUnexpectedCall).To(BeTrue())
			Expect(unexpectedCalls).NotTo(BeNil())
		})
	})

	Describe("work not order", func() {

		BeforeEach(func() {
			clientMock.MatchExpectationsInOrder(false)

			clientMock.ExpectSet("key", "1", 1*time.Second).SetVal("OK")
			clientMock.ExpectGet("key").SetVal("1")
			clientMock.ExpectGetSet("key", "0").SetVal("1")
		})

		AfterEach(func() {
			hasUnexpectedCall, unexpectedCalls := clientMock.UnexpectedCallsWereCalled()
			Expect(hasUnexpectedCall).To(BeFalse())
			Expect(unexpectedCalls).To(BeNil())
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

		AfterEach(func() {
			hasUnexpectedCall, unexpectedCalls := clientMock.UnexpectedCallsWereCalled()
			Expect(hasUnexpectedCall).To(BeFalse())
			Expect(unexpectedCalls).To(BeNil())
		})

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

		AfterEach(func() {
			hasUnexpectedCall, unexpectedCalls := clientMock.UnexpectedCallsWereCalled()
			Expect(hasUnexpectedCall).To(BeFalse())
			Expect(unexpectedCalls).To(BeNil())
		})

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
})
