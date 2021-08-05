package pushing

func StartServer() {
	sub := newSubscription()

	newRedisStream(sub).Start()
}
