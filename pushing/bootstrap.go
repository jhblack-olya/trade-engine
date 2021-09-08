/*
Copyright (C) 2021 Global Art Exchange, LLC (GAX). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package pushing

func StartServer() {
	sub := newSubscription()

	newRedisStream(sub).Start()
}
