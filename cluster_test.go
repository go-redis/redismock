package redismock

import (
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
)

var _ = Describe("Cluster", func() {
	var (
		client      *redis.ClusterClient
		clusterMock ClusterClientMock
	)

	BeforeEach(func() {
		client, clusterMock = NewClusterMock()
	})

	AfterEach(func() {
		Expect(client.Close()).NotTo(HaveOccurred())
		Expect(clusterMock.ExpectationsWereMet()).NotTo(HaveOccurred())
	})

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
})
