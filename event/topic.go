package event

// Topic of an event
type Topic uint

// Common topics
const (
	TopicStarted Topic = iota
	TopicCompleted
	TopicDownloading
	TopicOnline
)

// AllTopics is a slice of available topics
var AllTopics = []Topic{
	TopicStarted,
	TopicCompleted,
	TopicDownloading,
	TopicOnline,
}

// String returns a string of topic
func (t Topic) String() string {
	return [...]string{
		"started",
		"completed",
		"downloading",
		"online",
	}[t]
}
