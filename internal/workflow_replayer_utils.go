package internal

import (
	"fmt"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/util"
	"strings"
)

func matchReplayWithHistory(replayDecisions []*s.Decision, historyEvents []*s.HistoryEvent) error {
	di := 0
	hi := 0
	hSize := len(historyEvents)
	dSize := len(replayDecisions)
matchLoop:
	for hi < hSize || di < dSize {
		var e *s.HistoryEvent
		if hi < hSize {
			e = historyEvents[hi]
			if skipDeterministicCheckForUpsertChangeVersion(historyEvents, hi) {
				hi += 2
				continue matchLoop
			}
			if skipDeterministicCheckForEvent(e) {
				hi++
				continue matchLoop
			}
		}

		var d *s.Decision
		if di < dSize {
			d = replayDecisions[di]
			if skipDeterministicCheckForDecision(d) {
				di++
				continue matchLoop
			}
		}

		if d == nil {
			return fmt.Errorf("nondeterministic workflow: missing replay decision for %s", util.HistoryEventToString(e))
		}

		if e == nil {
			return fmt.Errorf("nondeterministic workflow: extra replay decision for %s", util.DecisionToString(d))
		}

		if !isDecisionMatchEvent(d, e, false) {
			return fmt.Errorf("nondeterministic workflow: history event is %s, replay decision is %s",
				util.HistoryEventToString(e), util.DecisionToString(d))
		}

		di++
		hi++
	}
	return nil
}

func lastPartOfName(name string) string {
	name = strings.TrimSuffix(name, "-fm")
	lastDotIdx := strings.LastIndex(name, ".")
	if lastDotIdx < 0 || lastDotIdx == len(name)-1 {
		return name
	}
	return name[lastDotIdx+1:]
}
