package telemetry

// Subscriber consumes the telemetry the bus produces: individual events as they
// happen, a path summary when a level ends, and the run profile as it updates.
// Implementations should be cheap and must not block.
type Subscriber interface {
	OnEvent(PlayerEvent)
	OnPathSummary(PathSummary)
	OnRunProfile(RunProfile)
}

// NopSubscriber is the default subscriber: it accepts everything and does
// nothing. It exists so the bus always has a valid sink wired in.
type NopSubscriber struct{}

func (NopSubscriber) OnEvent(PlayerEvent)       {}
func (NopSubscriber) OnPathSummary(PathSummary) {}
func (NopSubscriber) OnRunProfile(RunProfile)   {}
