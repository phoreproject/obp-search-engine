package mock

var connectionData = map[string][]string{"1": {"2", "3"}, "3": {"4", "1"}, "2": {"1"}, "4": {"3"}}

type MockRPC struct{}

func (m MockRPC) GetConnections(s string) ([]string, error) {
	return connectionData[s], nil
}
