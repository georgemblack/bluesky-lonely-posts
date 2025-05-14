package app

type APIFeedSkeletonResponse struct {
	Feed   []APIPost `json:"feed"`
	Cursor string    `json:"cursor,omitempty"`
}

type APIPost struct {
	Post string `json:"post"`
}

type APIDIDDocResponse struct {
	Context []string     `json:"@context"`
	ID      string       `json:"id"`
	Service []APIService `json:"service"`
}

type APIService struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}
