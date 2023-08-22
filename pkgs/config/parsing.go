package config

type RepsonseArray struct {
	Response []string `json:"reaction,omitempty"`
	Reaction []string `json:"response,omitempty"`
}

type base struct {
	Name  string `json:"name,omitempty"`
	Match string `json:"Match,omitempty"`
	RepsonseArray
}

type Command struct {
	base
}

type Keyword struct {
	base
}

type Pattern struct {
	base
}

type Filter struct {
	base
}

type URL struct {
	Name   string `json:"name,omitempty"`
	URL    string `json:"url,omitempty"`
	Format string `json:"format,omitempty"`
}

type Image struct {
	Types []string `json:"types,omitempty"`
	Sites URL      `json:"sites,omitempty"`
}

type BinSite struct {
	Sites  URL `json:"sites,omitempty"`
	Ignore URL `json:"ignore,omitempty"`
}

type Parse struct {
	Ping     RepsonseArray `json:"ping,omitempty"`
	Mention  RepsonseArray `json:"mention,omitempty"`
	Commands []Command     `json:"commands,omitempty"`
	Keywords []Keyword     `json:"keywords,omitempty"`
	Regex    []Pattern     `json:"regex,omitempty"`
	Image    `json:"image,omitempty"`
	BinSite  `json:"bin_sites,omitempty"`
}
