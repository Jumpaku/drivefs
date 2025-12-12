package form

type PublishState string

const (
	PublishStateUnpublished  PublishState = "unpublished"
	PublishStateAccepting    PublishState = "accepting"
	PublishStateNotAccepting PublishState = "not_accepting"
)
