package redismock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	ctx = context.TODO()
)

func TestRedisMock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "redis mock")
}

func operationStringCmd(base baseMock, expected func() *ExpectedString, actual func() *redis.StringCmd) {
	var (
		setErr = errors.New("string cmd error")
		str    string
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	str, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(str).To(Equal(""))

	base.ClearExpect()
	expected()
	str, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(str).To(Equal(""))

	base.ClearExpect()
	expected().SetVal("value")
	str, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(str).To(Equal("value"))
}

func operationStatusCmd(base baseMock, expected func() *ExpectedStatus, actual func() *redis.StatusCmd) {
	var (
		setErr = errors.New("status cmd error")
		str    string
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	str, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(str).To(Equal(""))

	base.ClearExpect()
	expected()
	str, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(str).To(Equal(""))

	base.ClearExpect()
	expected().SetVal("OK")
	str, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(str).To(Equal("OK"))
}

func operationIntCmd(base baseMock, expected func() *ExpectedInt, actual func() *redis.IntCmd) {
	var (
		setErr = errors.New("int cmd error")
		val    int64
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(int64(0)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(int64(0)))

	base.ClearExpect()
	expected().SetVal(1024)
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(int64(1024)))
}

func operationBoolCmd(base baseMock, expected func() *ExpectedBool, actual func() *redis.BoolCmd) {
	var (
		setErr = errors.New("bool cmd error")
		val    bool
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(BeFalse())

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(BeFalse())

	base.ClearExpect()
	expected().SetVal(true)
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(BeTrue())
}

func operationStringSliceCmd(base baseMock, expected func() *ExpectedStringSlice, actual func() *redis.StringSliceCmd) {
	var (
		setErr = errors.New("string slice cmd error")
		val    []string
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]string(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]string(nil)))

	base.ClearExpect()
	expected().SetVal([]string{"redis", "move"})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]string{"redis", "move"}))
}

func operationDurationCmd(base baseMock, expected func() *ExpectedDuration, actual func() *redis.DurationCmd) {
	var (
		setErr = errors.New("duration cmd error")
		val    time.Duration
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(time.Duration(0)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(time.Duration(0)))

	base.ClearExpect()
	expected().SetVal(2 * time.Hour)
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(2 * time.Hour))
}

func operationSliceCmd(base baseMock, expected func() *ExpectedSlice, actual func() *redis.SliceCmd) {
	var (
		setErr = errors.New("slice cmd error")
		val    []interface{}
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]interface{}(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]interface{}(nil)))

	base.ClearExpect()
	expected().SetVal([]interface{}{"mock", "slice"})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]interface{}{"mock", "slice"}))
}

func operationFloatCmd(base baseMock, expected func() *ExpectedFloat, actual func() *redis.FloatCmd) {
	var (
		setErr = errors.New("float cmd error")
		val    float64
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(float64(0)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(float64(0)))

	base.ClearExpect()
	expected().SetVal(1)
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(float64(1)))
}

func operationFloatSliceCmd(base baseMock, expected func() *ExpectedFloatSlice, actual func() *redis.FloatSliceCmd) {
	var (
		setErr = errors.New("float slice cmd error")
		val    []float64
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]float64(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]float64(nil)))

	base.ClearExpect()
	expected().SetVal([]float64{11.11, 22.22, 99.99999})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]float64{11.11, 22.22, 99.99999}))
}

func operationIntSliceCmd(base baseMock, expected func() *ExpectedIntSlice, actual func() *redis.IntSliceCmd) {
	var (
		setErr = errors.New("int slice cmd error")
		val    []int64
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]int64(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]int64(nil)))

	base.ClearExpect()
	expected().SetVal([]int64{1, 2, 3, 4})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]int64{1, 2, 3, 4}))
}

func operationScanCmd(base baseMock, expected func() *ExpectedScan, actual func() *redis.ScanCmd) {
	var (
		setErr = errors.New("scan cmd error")
		page   []string
		cursor uint64
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	page, cursor, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(page).To(Equal([]string(nil)))
	Expect(cursor).To(Equal(uint64(0)))

	base.ClearExpect()
	expected()
	page, cursor, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(page).To(Equal([]string(nil)))
	Expect(cursor).To(Equal(uint64(0)))

	base.ClearExpect()
	expected().SetVal([]string{"key1", "key2", "key3"}, 5)
	page, cursor, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(page).To(Equal([]string{"key1", "key2", "key3"}))
	Expect(cursor).To(Equal(uint64(5)))
}

func operationMapStringStringCmd(base baseMock, expected func() *ExpectedMapStringString, actual func() *redis.MapStringStringCmd) {
	var (
		setErr = errors.New("map string string cmd error")
		val    map[string]string
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(map[string]string(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(map[string]string(nil)))

	base.ClearExpect()
	expected().SetVal(map[string]string{"key": "value", "key2": "value2"})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(map[string]string{"key": "value", "key2": "value2"}))
}

func operationStringStructMapCmd(base baseMock, expected func() *ExpectedStringStructMap, actual func() *redis.StringStructMapCmd) {
	var (
		setErr = errors.New("string struct map cmd error")
		val    map[string]struct{}
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(map[string]struct{}(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(map[string]struct{}(nil)))

	base.ClearExpect()
	expected().SetVal([]string{"key1", "key2"})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(map[string]struct{}{"key1": {}, "key2": {}}))
}

func operationXMessageSliceCmd(base baseMock, expected func() *ExpectedXMessageSlice, actual func() *redis.XMessageSliceCmd) {
	var (
		setErr = errors.New("x message slice cmd error")
		val    []redis.XMessage
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.XMessage(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.XMessage(nil)))

	base.ClearExpect()
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

func operationXStreamSliceCmd(base baseMock, expected func() *ExpectedXStreamSlice, actual func() *redis.XStreamSliceCmd) {
	var (
		setErr = errors.New("x stream slice cmd error")
		val    []redis.XStream
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.XStream(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.XStream(nil)))

	base.ClearExpect()
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

func operationXPendingCmd(base baseMock, expected func() *ExpectedXPending, actual func() *redis.XPendingCmd) {
	var (
		setErr = errors.New("x pending cmd error")
		val    *redis.XPending
		valNil *redis.XPending
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(valNil))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(valNil))

	base.ClearExpect()
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

func operationXPendingExtCmd(base baseMock, expected func() *ExpectedXPendingExt, actual func() *redis.XPendingExtCmd) {
	var (
		setErr = errors.New("x pending ext cmd error")
		val    []redis.XPendingExt
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.XPendingExt(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.XPendingExt(nil)))

	base.ClearExpect()
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

func operationXAutoClaimCmd(base baseMock, expected func() *ExpectedXAutoClaim, actual func() *redis.XAutoClaimCmd) {
	var (
		setErr = errors.New("x auto claim cmd error")
		start  string
		val    []redis.XMessage
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, start, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(start).To(Equal(""))
	Expect(val).To(Equal([]redis.XMessage(nil)))

	base.ClearExpect()
	expected()
	val, start, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(start).To(Equal(""))
	Expect(val).To(Equal([]redis.XMessage(nil)))

	base.ClearExpect()
	expected().SetVal([]redis.XMessage{
		{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
		{ID: "3-0", Values: map[string]interface{}{"tres": "troix"}},
	}, "3-0")
	val, start, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(start).To(Equal("3-0"))
	Expect(val).To(Equal([]redis.XMessage{
		{ID: "1-0", Values: map[string]interface{}{"uno": "un"}},
		{ID: "2-0", Values: map[string]interface{}{"dos": "deux"}},
		{ID: "3-0", Values: map[string]interface{}{"tres": "troix"}},
	}))
}

func operationXAutoClaimJustIDCmd(base baseMock, expected func() *ExpectedXAutoClaimJustID, actual func() *redis.XAutoClaimJustIDCmd) {
	var (
		setErr = errors.New("x auto claim just id cmd error")
		start  string
		val    []string
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, start, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(start).To(Equal(""))
	Expect(val).To(Equal([]string(nil)))

	base.ClearExpect()
	expected()
	val, start, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(start).To(Equal(""))
	Expect(val).To(Equal([]string(nil)))

	base.ClearExpect()
	expected().SetVal([]string{"1-0", "2-0", "3-0"}, "3-0")
	val, start, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(start).To(Equal("3-0"))
	Expect(val).To(Equal([]string{"1-0", "2-0", "3-0"}))
}

func operationXInfoGroupsCmd(base baseMock, expected func() *ExpectedXInfoGroups, actual func() *redis.XInfoGroupsCmd) {
	var (
		setErr = errors.New("x info group cmd error")
		val    []redis.XInfoGroup
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.XInfoGroup(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.XInfoGroup(nil)))

	base.ClearExpect()
	expected().SetVal([]redis.XInfoGroup{
		{Name: "name1", Consumers: 1, Pending: 2, LastDeliveredID: "last1"},
		{Name: "name2", Consumers: 1, Pending: 2, LastDeliveredID: "last2"},
		{Name: "name3", Consumers: 1, Pending: 2, LastDeliveredID: "last2"},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]redis.XInfoGroup{
		{Name: "name1", Consumers: 1, Pending: 2, LastDeliveredID: "last1"},
		{Name: "name2", Consumers: 1, Pending: 2, LastDeliveredID: "last2"},
		{Name: "name3", Consumers: 1, Pending: 2, LastDeliveredID: "last2"},
	}))
}

func operationXInfoStreamCmd(base baseMock, expected func() *ExpectedXInfoStream, actual func() *redis.XInfoStreamCmd) {
	var (
		setErr = errors.New("x info stream cmd error")
		val    *redis.XInfoStream
		nilVal *redis.XInfoStream
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(nilVal))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(nilVal))

	base.ClearExpect()
	expected().SetVal(&redis.XInfoStream{
		Length:          1,
		RadixTreeKeys:   10,
		RadixTreeNodes:  20,
		Groups:          30,
		LastGeneratedID: "id",
		FirstEntry: redis.XMessage{
			ID: "first_id",
			Values: map[string]interface{}{
				"first_key": "first_value",
			},
		},
		LastEntry: redis.XMessage{
			ID: "last_id",
			Values: map[string]interface{}{
				"last_key": "last_value",
			},
		},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(&redis.XInfoStream{
		Length:          1,
		RadixTreeKeys:   10,
		RadixTreeNodes:  20,
		Groups:          30,
		LastGeneratedID: "id",
		FirstEntry: redis.XMessage{
			ID: "first_id",
			Values: map[string]interface{}{
				"first_key": "first_value",
			},
		},
		LastEntry: redis.XMessage{
			ID: "last_id",
			Values: map[string]interface{}{
				"last_key": "last_value",
			},
		},
	}))
}

func operationXInfoConsumersCmd(base baseMock, expected func() *ExpectedXInfoConsumers, actual func() *redis.XInfoConsumersCmd) {
	var (
		setErr = errors.New("x info consumers cmd error")
		val    []redis.XInfoConsumer
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.XInfoConsumer(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.XInfoConsumer(nil)))

	base.ClearExpect()
	expected().SetVal([]redis.XInfoConsumer{
		{
			Name:    "c1",
			Pending: 2,
			Idle:    1,
		},
		{
			Name:    "c2",
			Pending: 1,
			Idle:    2,
		},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]redis.XInfoConsumer{
		{
			Name:    "c1",
			Pending: 2,
			Idle:    1,
		},
		{
			Name:    "c2",
			Pending: 1,
			Idle:    2,
		},
	}))
}

func operationXInfoStreamFullCmd(base baseMock, expected func() *ExpectedXInfoStreamFull, actual func() *redis.XInfoStreamFullCmd) {
	var (
		setErr = errors.New("x info stream full cmd error")
		val    *redis.XInfoStreamFull
		valNil *redis.XInfoStreamFull
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(valNil))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(valNil))

	now := time.Now()
	now2 := now.Add(3 * time.Hour)
	now3 := now.Add(5 * time.Hour)
	now4 := now.Add(6 * time.Hour)
	now5 := now.Add(8 * time.Hour)
	now6 := now.Add(9 * time.Hour)
	base.ClearExpect()
	expected().SetVal(&redis.XInfoStreamFull{
		Length:          3,
		RadixTreeKeys:   4,
		RadixTreeNodes:  5,
		LastGeneratedID: "3-3",
		Entries: []redis.XMessage{
			{
				ID: "1-0",
				Values: map[string]interface{}{
					"key1": "val1",
					"key2": "val2",
				},
			},
		},
		Groups: []redis.XInfoStreamGroup{
			{
				Name:            "group1",
				LastDeliveredID: "10-1",
				PelCount:        3,
				Pending: []redis.XInfoStreamGroupPending{
					{
						ID:            "5-1",
						Consumer:      "consumer1",
						DeliveryTime:  now,
						DeliveryCount: 9,
					},
					{
						ID:            "5-2",
						Consumer:      "consumer2",
						DeliveryTime:  now,
						DeliveryCount: 8,
					},
				},
				Consumers: []redis.XInfoStreamConsumer{
					{
						Name:     "consumer3",
						SeenTime: now2,
						PelCount: 7,
						Pending: []redis.XInfoStreamConsumerPending{
							{
								ID:            "6-1",
								DeliveryTime:  now3,
								DeliveryCount: 3,
							},
							{
								ID:            "6-2",
								DeliveryTime:  now4,
								DeliveryCount: 2,
							},
						},
					},
					{
						Name:     "consumer4",
						SeenTime: now,
						PelCount: 6,
						Pending: []redis.XInfoStreamConsumerPending{
							{
								ID:            "7-1",
								DeliveryTime:  now5,
								DeliveryCount: 5,
							},
							{
								ID:            "8-2",
								DeliveryTime:  now6,
								DeliveryCount: 6,
							},
						},
					},
				},
			},
		},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(&redis.XInfoStreamFull{
		Length:          3,
		RadixTreeKeys:   4,
		RadixTreeNodes:  5,
		LastGeneratedID: "3-3",
		Entries: []redis.XMessage{
			{
				ID: "1-0",
				Values: map[string]interface{}{
					"key1": "val1",
					"key2": "val2",
				},
			},
		},
		Groups: []redis.XInfoStreamGroup{
			{
				Name:            "group1",
				LastDeliveredID: "10-1",
				PelCount:        3,
				Pending: []redis.XInfoStreamGroupPending{
					{
						ID:            "5-1",
						Consumer:      "consumer1",
						DeliveryTime:  now,
						DeliveryCount: 9,
					},
					{
						ID:            "5-2",
						Consumer:      "consumer2",
						DeliveryTime:  now,
						DeliveryCount: 8,
					},
				},
				Consumers: []redis.XInfoStreamConsumer{
					{
						Name:     "consumer3",
						SeenTime: now2,
						PelCount: 7,
						Pending: []redis.XInfoStreamConsumerPending{
							{
								ID:            "6-1",
								DeliveryTime:  now3,
								DeliveryCount: 3,
							},
							{
								ID:            "6-2",
								DeliveryTime:  now4,
								DeliveryCount: 2,
							},
						},
					},
					{
						Name:     "consumer4",
						SeenTime: now,
						PelCount: 6,
						Pending: []redis.XInfoStreamConsumerPending{
							{
								ID:            "7-1",
								DeliveryTime:  now5,
								DeliveryCount: 5,
							},
							{
								ID:            "8-2",
								DeliveryTime:  now6,
								DeliveryCount: 6,
							},
						},
					},
				},
			},
		},
	}))
}

func operationZWithKeyCmd(base baseMock, expected func() *ExpectedZWithKey, actual func() *redis.ZWithKeyCmd) {
	var (
		setErr = errors.New("z with key cmd error")
		val    *redis.ZWithKey
		valNil *redis.ZWithKey
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(valNil))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(valNil))

	base.ClearExpect()
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

func operationZSliceCmd(base baseMock, expected func() *ExpectedZSlice, actual func() *redis.ZSliceCmd) {
	var (
		setErr = errors.New("z slice cmd error")
		val    []redis.Z
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.Z(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.Z(nil)))

	base.ClearExpect()
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

func operationTimeCmd(base baseMock, expected func() *ExpectedTime, actual func() *redis.TimeCmd) {
	var (
		setErr = errors.New("time cmd error")
		val    time.Time
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(time.Time{}))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(time.Time{}))

	base.ClearExpect()
	now := time.Now()
	expected().SetVal(now)
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(now))
}

func operationCmdCmd(base baseMock, expected func() *ExpectedCmd, actual func() *redis.Cmd) {
	var (
		setErr = errors.New("cmd error")
		val    interface{}
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(BeNil())

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(BeNil())

	base.ClearExpect()
	expected().SetVal(interface{}(1024))
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(interface{}(1024)))
}

func operationBoolSliceCmd(base baseMock, expected func() *ExpectedBoolSlice, actual func() *redis.BoolSliceCmd) {
	var (
		setErr = errors.New("bool slice cmd error")
		val    []bool
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]bool(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]bool(nil)))

	base.ClearExpect()
	expected().SetVal([]bool{true, false, true})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]bool{true, false, true}))
}

func operationMapStringIntCmd(base baseMock, expected func() *ExpectedMapStringInt, actual func() *redis.MapStringIntCmd) {
	var (
		setErr = errors.New("map string int cmd error")
		val    map[string]int64
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal(map[string]int64(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal(map[string]int64(nil)))

	base.ClearExpect()
	expected().SetVal(map[string]int64{"key": 1, "key2": 2})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal(map[string]int64{"key": 1, "key2": 2}))
}

func operationClusterSlotsCmd(base baseMock, expected func() *ExpectedClusterSlots, actual func() *redis.ClusterSlotsCmd) {
	var (
		setErr = errors.New("cluster slots cmd error")
		val    []redis.ClusterSlot
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.ClusterSlot(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.ClusterSlot(nil)))

	base.ClearExpect()
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

func operationGeoLocationCmd(base baseMock, expected func() *ExpectedGeoLocation, actual func() *redis.GeoLocationCmd) {
	var (
		setErr = errors.New("geo location cmd error")
		val    []redis.GeoLocation
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.GeoLocation(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.GeoLocation(nil)))

	base.ClearExpect()
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

func operationGeoPosCmd(base baseMock, expected func() *ExpectedGeoPos, actual func() *redis.GeoPosCmd) {
	var (
		setErr = errors.New("geo pos cmd error")
		val    []*redis.GeoPos
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]*redis.GeoPos(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]*redis.GeoPos(nil)))

	base.ClearExpect()
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

func operationGeoSearchLocationCmd(base baseMock, expected func() *ExpectedGeoSearchLocation, actual func() *redis.GeoSearchLocationCmd) {
	var (
		setErr = errors.New("geo search location cmd error")
		val    []redis.GeoLocation
		err    error
	)

	base.ClearExpect()
	expected().SetErr(setErr)
	val, err = actual().Result()
	Expect(err).To(Equal(setErr))
	Expect(val).To(Equal([]redis.GeoLocation(nil)))

	base.ClearExpect()
	expected()
	val, err = actual().Result()
	Expect(err).To(HaveOccurred())
	Expect(val).To(Equal([]redis.GeoLocation(nil)))

	base.ClearExpect()
	expected().SetVal([]redis.GeoLocation{
		{
			Name:      "Catania",
			Longitude: 15.08726745843887329,
			Latitude:  37.50266842333162032,
			Dist:      56.4413,
			GeoHash:   3479447370796909,
		},
		{
			Name:      "Palermo",
			Longitude: 13.36138933897018433,
			Latitude:  38.11555639549629859,
			Dist:      190.4424,
			GeoHash:   3479099956230698,
		},
	})
	val, err = actual().Result()
	Expect(err).NotTo(HaveOccurred())
	Expect(val).To(Equal([]redis.GeoLocation{
		{
			Name:      "Catania",
			Longitude: 15.08726745843887329,
			Latitude:  37.50266842333162032,
			Dist:      56.4413,
			GeoHash:   3479447370796909,
		},
		{
			Name:      "Palermo",
			Longitude: 13.36138933897018433,
			Latitude:  38.11555639549629859,
			Dist:      190.4424,
			GeoHash:   3479099956230698,
		},
	}))
}
