package redis_sentinel_proxy_test

import (
	"context"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
)

const (
	redisSentinelProxyAddr = "redis-sentinel-proxy:9999"
	redisMasterAddr        = "redis-master:6379"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "")
}

var _ = Describe("redis-sentinel-proxy :: tests", func() {
	Context("Default behaviour", Ordered, func() {
		var rspClient, rmClient *redis.Client
		BeforeEach(func() {
			rspClient = connect(redisSentinelProxyAddr)
			rmClient = connect(redisMasterAddr)
		})

		AfterEach(func() {
			Expect(rmClient.Close()).To(Succeed())
			Expect(rspClient.Close()).To(Succeed())
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

	Context("Count available connections", Ordered, func() {
		var checkClient *redis.Client
		BeforeEach(func() {
			checkClient = connect(redisMasterAddr)
		})

		AfterEach(func() {
			Expect(checkClient.Close()).To(Succeed())
		})

		It("Counts clients", func() {
			clients := make([]*redis.Client, 40)
			for i := range clients {
				client := connect(redisSentinelProxyAddr)
				ping(client)
				clients[i] = client
			}
			listClients(checkClient, 40, 50)

			for i := range clients {
				Expect(clients[i].Close()).To(Succeed())
			}
			time.Sleep(time.Second * 5)
			listClients(checkClient, 0, 10)
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

func listClients(client *redis.Client, low, high int) {
	ctx, cancel := ctxTimeout()
	defer cancel()

	ans, err := client.ClientList(ctx).Result()
	Expect(err).NotTo(HaveOccurred())
	lines := strings.Split(ans, "\n")
	Expect(len(lines)).To(And(BeNumerically(">", low), BeNumerically("<", high)))
}

func ctxTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second*3)
}

func errAnswer(ans, expected interface{}, err error) {
	Expect(err).NotTo(HaveOccurred())
	Expect(ans).To(Equal(expected))
}
