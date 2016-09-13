package tat

// SystemCacheClean clean cache, only for tat admin
func (c *Client) SystemCacheClean() ([]byte, error) {
	return c.simpleGetAndGetBytes("/system/cache/clean")
}
