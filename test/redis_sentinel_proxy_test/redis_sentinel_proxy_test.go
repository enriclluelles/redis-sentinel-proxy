package redis_sentinel_proxy_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "")
}

var _ = Describe("redis-sentinel-proxy :: tests", func() {
	var rspClient, rmClient *redis.Client
	BeforeEach(func() {
		rspClient = connect("redis-sentinel-proxy:9999")
		rmClient = connect("redis-master:6379")
	})

	AfterEach(func() {
		defer rmClient.Close()
		defer rmClient.Close()
	})

	Context("ping-pong RSP", func() {
		It("returns PONG", func() {
			ping(rspClient)
		})
	})

	Context("ping-pong master", func() {
		It("returns PONG", func() {
			ping(rmClient)
		})
	})

	Context("set from RSP", func() {
		BeforeEach(func() {
			setKey(rmClient, "test-key-rsp", "test-value-1")
		})

		AfterEach(func() {
			delKey(rmClient, "test-key-rsp")
		})

		It("returns key from RSP", func() {
			getKey(rspClient, "test-key-rsp", "test-value-1")
		})

		It("returns key from master", func() {
			getKey(rmClient, "test-key-rsp", "test-value-1")
		})
	})

	Context("set from master", func() {
		BeforeEach(func() {
			setKey(rmClient, "test-key-rm", "test-value-2")
		})

		AfterEach(func() {
			delKey(rmClient, "test-key-rm")
		})

		It("returns key from RSP", func() {
			getKey(rspClient, "test-key-rm", "test-value-2")
		})

		It("returns key from master", func() {
			getKey(rmClient, "test-key-rm", "test-value-2")
		})
	})

})

func connect(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
}

func ping(client *redis.Client) {
	ctx, cancel := ctxTimeout()
	defer cancel()
	ans, err := client.Ping(ctx).Result()
	errAnswer(ans, "PONG", err)
}

func setKey(client *redis.Client, key, value string) {
	ctx, cancel := ctxTimeout()
	defer cancel()

	ans, err := client.Set(ctx, key, value, time.Hour).Result()
	errAnswer(ans, "OK", err)
}

func getKey(client *redis.Client, key, value string) {
	ctx, cancel := ctxTimeout()
	defer cancel()

	ans, err := client.Get(ctx, key).Result()
	errAnswer(ans, value, err)
}

func delKey(client *redis.Client, key string) {
	ctx, cancel := ctxTimeout()
	defer cancel()

	ans, err := client.Del(ctx, key).Result()
	errAnswer(ans, int64(1), err)
}

func ctxTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*3)
}

func errAnswer(ans, expected interface{}, err error) {
	Expect(err).NotTo(HaveOccurred())
	Expect(ans).To(Equal(expected))
}
