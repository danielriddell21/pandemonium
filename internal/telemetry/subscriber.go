package telemetry

type Subscriber interface {
	OnEvent(PlayerEvent)
	OnPathSummary(PathSummary)
	OnRunProfile(RunProfile)
}

type NopSubscriber struct{}

func (NopSubscriber) OnEvent(PlayerEvent) {}

func (NopSubscriber) OnPathSummary(PathSummary) {}

func (NopSubscriber) OnRunProfile(RunProfile) {}
