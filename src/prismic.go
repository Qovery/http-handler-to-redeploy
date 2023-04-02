package src

type PrismicWebhook struct {
	Type       string            `json:"type"`
	Secret     string            `json:"secret"`
	MasterRef  string            `json:"masterRef"`
	Domain     string            `json:"domain"`
	ApiUrl     string            `json:"apiUrl"`
	Releases   ReleaseUpdate     `json:"releases"`
	Bookmarks  map[string]string `json:"bookmarks"`
	Collection map[string]string `json:"collection"`
	Tags       TagsUpdate        `json:"tags"`
	Documents  []string          `json:"documents"`
}

type ReleaseUpdate struct {
	Addition []Release `json:"addition"`
	Update   []Release `json:"update"`
	Deletion []Release `json:"deletion"`
}

type Release struct {
	ID        string   `json:"id"`
	Ref       string   `json:"ref"`
	Label     string   `json:"label"`
	Documents []string `json:"documents"`
}

type TagsUpdate struct {
	Addition []Tag `json:"addition"`
	Deletion []Tag `json:"deletion"`
}

type Tag struct {
	ID string `json:"id"`
}
