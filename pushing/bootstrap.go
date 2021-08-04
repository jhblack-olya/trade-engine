package pushing

// import "gitlab.com/gae4/trade-engine/conf"

func StartServer() {
	// gbeConfig := conf.GetConfig()

	sub := newSubscription()

	newRedisStream(sub).Start()
}
