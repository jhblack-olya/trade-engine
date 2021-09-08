/*
Copyright (C) 2021 Global Art Exchange, LLC (GAX). All Rights Reserved.
You may not use, distribute and modify this code without a license;
To obtain a license write to legal@gax.llc
*/

package pushing

import "sync"

// Each connection corresponds to a client, and the client is responsible for the data I / O of the connection
type Client struct {
	id       int64
	writeCh  chan interface{}
	sub      *subscription
	channels map[string]struct{}
	mu       sync.Mutex
}
