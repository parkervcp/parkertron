package main

type match struct {
	Match    []string `json:"match,omitempty"`
	Reaction []string `json:"reaction,omitempty"`
	Response []string `json:"response,omitempty"`
	Exact    bool     `json:"exact,omitempty"`
}

type permGroups struct {
	Group     string   `json:"group,omitempty"`
	User      []string `json:"user,omitempty"`
	Role      []string `json:"role,omitempty"`
	Commands  []string `json:"commands,omitempty"`
	Blacklist bool     `json:"blacklisted,omitempty"`
}

type command struct {
	Command  string   `json:"command,omitempty"`
	Response []string `json:"response,omitempty"`
	Reaction []string `json:"reaction,omitempty"`
}

type keyword struct {
	Keywords []string `json:"keywords,omitempty"`
	Reaction []string `json:"reaction,omitempty"`
	Response []string `json:"response,omitempty"`
	Exact    bool     `json:"exact,omitempty"`
}

type mentions struct {
	Ping    []string `json:"ping,omitempty"`
	Mention []string `json:"mention,omitempty"`
}

type parsing struct {
	Image parsingImageConfig   `json:"image,omitempty"`
	Paste []parsingPasteConfig `json:"paste,omitempty"`
}

type parsingPasteConfig struct {
	Name   string `json:"name,omitempty"`
	URL    string `json:"url,omitempty"`
	Format string `json:"format,omitempty"`
}

type parsingImageConfig struct {
	Filetypes []string `json:"filetypes,omitempty"`
}
