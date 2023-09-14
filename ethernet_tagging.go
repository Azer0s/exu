package exu

type TagData []byte

func WithTagging(tagging VlanTagging, tags ...uint32) TagData {
	if tagging == TaggingUntagged {
		return TagData{}
	}

	if tagging == TaggingTagged {
		return TagData{
			byte(tags[0] >> 8),
			byte(tags[0]),
		}
	}

	return TagData{
		byte(tags[0] >> 8),
		byte(tags[0]),
		byte(tags[1] >> 8),
		byte(tags[1]),
	}
}

func (t TagData) GetTagging() VlanTagging {
	if len(t) == 0 {
		return TaggingUntagged
	}
	if len(t) == 2 {
		return TaggingTagged
	}
	return TaggingDoubleTagged
}
