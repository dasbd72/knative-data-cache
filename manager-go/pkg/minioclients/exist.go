package minioclients

// return if endpoint exists
func (mcs *MinioClients) Exist(endpoint string) bool {
	mcs.mux.Lock()
	_, ok := mcs.entries[endpoint]
	mcs.mux.Unlock()

	return ok
}
