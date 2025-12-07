package station

type Info struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	FaviconURL  string `json:"faviconURL"`
	LogoURL     string `json:"logoURL"`
	Location    string `json:"location"`
	Timezone    string `json:"timezone"`
	Links       string `json:"links"`
}

type Property struct {
	Key   string
	Value string
}

type Store interface {
	StationProperties() ([]*Property, error)
	UpsertStationProperty(key, value string) (*Property, error)
	DeleteStationProperty(key string) error
}
