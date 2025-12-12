package form

type EmailCollectionType string

const (
	EmailCollectionTypeUnspecified    EmailCollectionType = "EMAIL_COLLECTION_TYPE_UNSPECIFIED"
	EmailCollectionTypeDoNotCollect   EmailCollectionType = "DO_NOT_COLLECT"
	EmailCollectionTypeVerified       EmailCollectionType = "VERIFIED"
	EmailCollectionTypeResponderInput EmailCollectionType = "RESPONDER_INPUT"
)
