package core

type Metadata struct {
	Audio  *MetadataAudio   `json:"audio,omitempty"`
	Stcker *MetadataSticker `json:"stcker,omitempty"`
}

type MetadataAudio struct {
	Title     string `json:"title,omitempty"`
	Performer string `json:"performer,omitempty"`
}

type MetadataSticker struct {
	SetName string `json:"set_name,omitempty"`
	Emoji   string `json:"emoji,omitempty"`
}

func NewMetadataAudio(title, performer string) Metadata {
	return Metadata{
		Audio: &MetadataAudio{
			Title:     title,
			Performer: performer,
		},
	}
}
